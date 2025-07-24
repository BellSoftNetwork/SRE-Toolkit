package k8sclient

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/k8s-diff/config"
	"gitlab.bellsoft.net/devops/sre-workbench/go/pkg/k8s/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sClientWrapper k8s 클라이언트 래퍼
type K8sClientWrapper struct {
	*client.Client
	context       string
	clientset     kubernetes.Interface
	dynamicClient dynamic.Interface
}

// CreateClientPair 소스와 타겟 클라이언트 쌍 생성
func CreateClientPair(cfg *config.Config, sourceContext, targetContext string) (*K8sClientWrapper, *K8sClientWrapper, error) {
	sourceClient, err := CreateClientWithContext(cfg, sourceContext)
	if err != nil {
		return nil, nil, fmt.Errorf("소스 클라이언트 생성 실패: %w", err)
	}

	targetClient, err := CreateClientWithContext(cfg, targetContext)
	if err != nil {
		return nil, nil, fmt.Errorf("타겟 클라이언트 생성 실패: %w", err)
	}

	return sourceClient, targetClient, nil
}

// CreateClientWithContext 특정 컨텍스트로 클라이언트 생성
func CreateClientWithContext(cfg *config.Config, contextName string) (*K8sClientWrapper, error) {
	// kubeconfig 로드
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{
		CurrentContext: contextName,
	}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	restConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("kubeconfig 로드 실패: %w", err)
	}

	// 성능 설정
	restConfig.QPS = 100
	restConfig.Burst = 100
	restConfig.Timeout = 60 * time.Second

	// clientset 생성
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("clientset 생성 실패: %w", err)
	}

	// dynamic client 생성
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("dynamic client 생성 실패: %w", err)
	}

	// 기존 client 구조체 생성 (호환성을 위해)
	skipMap := make(map[string]bool)
	for _, rt := range cfg.SkipResourceTypes {
		skipMap[rt] = true
	}

	clientConfig := &client.ClientConfig{
		ImportantResourceTypes: cfg.ImportantResourceTypes,
		SkipResourceTypes:      skipMap,
	}

	// 임시로 환경변수 설정하여 기존 client 생성
	tempFile, err := createTempKubeconfig(contextName)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempFile)

	originalKubeconfig := os.Getenv("KUBECONFIG")
	os.Setenv("KUBECONFIG", tempFile)
	baseClient, err := client.NewClient(clientConfig)
	os.Setenv("KUBECONFIG", originalKubeconfig)

	if err != nil {
		return nil, err
	}

	return &K8sClientWrapper{
		Client:        baseClient,
		context:       contextName,
		clientset:     clientset,
		dynamicClient: dynamicClient,
	}, nil
}

// createTempKubeconfig 임시 kubeconfig 파일 생성
func createTempKubeconfig(contextName string) (string, error) {
	// kubectl config view로 특정 컨텍스트만 추출
	cmd := exec.Command("kubectl", "config", "view",
		"--minify",
		"--context="+contextName,
		"--flatten",
	)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("kubectl config view 실행 실패: %w", err)
	}

	// 임시 파일 생성
	tempFile, err := os.CreateTemp("", "kubeconfig-*.yaml")
	if err != nil {
		return "", err
	}
	tempFile.Close()

	if err := os.WriteFile(tempFile.Name(), output, 0600); err != nil {
		os.Remove(tempFile.Name())
		return "", err
	}

	return tempFile.Name(), nil
}

// GetAllResourceTypes 모든 리소스 타입 가져오기 (CRD 포함)
func (c *K8sClientWrapper) GetAllResourceTypes() ([]string, error) {
	// Discovery 클라이언트로 모든 API 리소스 가져오기
	discoveryClient := c.clientset.Discovery()

	apiResourceLists, err := discoveryClient.ServerPreferredNamespacedResources()
	if err != nil {
		// 일부 API 그룹에서 에러가 발생해도 계속 진행
		// (예: metrics.k8s.io 등이 없을 수 있음)
	}

	var resourceTypes []string
	seen := make(map[string]bool)

	for _, apiResourceList := range apiResourceLists {
		// 메트릭 관련 API 그룹 제외
		gv, _ := schema.ParseGroupVersion(apiResourceList.GroupVersion)
		if strings.HasSuffix(gv.Group, "metrics.k8s.io") {
			continue
		}

		for _, apiResource := range apiResourceList.APIResources {
			// List 권한이 있는 리소스만
			if !contains(apiResource.Verbs, "list") {
				continue
			}

			// 서브리소스 제외 (예: pods/log, pods/status)
			if strings.Contains(apiResource.Name, "/") {
				continue
			}

			// 리소스 타입 구성
			var resourceType string
			if apiResourceList.GroupVersion == "v1" {
				resourceType = apiResource.Name
			} else {
				resourceType = apiResource.Name + "." + gv.Group
			}

			// 중복 제거
			if !seen[resourceType] {
				seen[resourceType] = true
				resourceTypes = append(resourceTypes, resourceType)
			}
		}
	}

	return resourceTypes, nil
}

// GetResourcesInNamespace 네임스페이스의 특정 리소스 타입 가져오기
func (c *K8sClientWrapper) GetResourcesInNamespace(resourceType, namespace string) ([]unstructured.Unstructured, error) {
	// 리소스 타입 파싱
	parts := strings.SplitN(resourceType, ".", 2)
	var gvr schema.GroupVersionResource

	if len(parts) == 1 {
		// core API group (v1)
		gvr = schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: parts[0],
		}
	} else {
		// 다른 API group
		// 가장 선호되는 버전 찾기
		var err error
		gvr, err = c.findPreferredVersion(parts[1], parts[0])
		if err != nil {
			return nil, err
		}
	}

	// 리소스 조회
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	list, err := c.dynamicClient.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

// findPreferredVersion API 그룹의 선호 버전 찾기
func (c *K8sClientWrapper) findPreferredVersion(group string, resource string) (schema.GroupVersionResource, error) {
	discoveryClient := c.clientset.Discovery()

	apiGroupList, err := discoveryClient.ServerGroups()
	if err != nil {
		return schema.GroupVersionResource{}, err
	}

	for _, apiGroup := range apiGroupList.Groups {
		if apiGroup.Name == group {
			// PreferredVersion 사용
			gv, err := schema.ParseGroupVersion(apiGroup.PreferredVersion.GroupVersion)
			if err != nil {
				return schema.GroupVersionResource{}, err
			}
			return schema.GroupVersionResource{
				Group:    gv.Group,
				Version:  gv.Version,
				Resource: resource,
			}, nil
		}
	}

	return schema.GroupVersionResource{}, fmt.Errorf("API 그룹 %s를 찾을 수 없습니다", group)
}

// GetContext 컨텍스트 반환
func (c *K8sClientWrapper) GetContext() string {
	return c.context
}

// contains 문자열 슬라이스에 값이 있는지 확인
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetCurrentContext 현재 컨텍스트 반환 (인터페이스 구현)
func (c *K8sClientWrapper) GetCurrentContext() (string, string) {
	// 컨텍스트명을 그대로 클러스터명으로 사용
	// ARN 형태의 컨텍스트명은 나중에 ExtractClusterName으로 처리됨
	return c.context, c.context
}

// GetAllNamespaces 모든 네임스페이스 가져오기 (인터페이스 구현)
func (c *K8sClientWrapper) GetAllNamespaces() ([]string, error) {
	// 기존 Client의 메서드 사용
	if c.Client != nil {
		return c.Client.GetAllNamespaces()
	}

	// 또는 직접 구현
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	nsList, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var namespaces []string
	for _, ns := range nsList.Items {
		namespaces = append(namespaces, ns.Name)
	}
	return namespaces, nil
}

// GetResourceTypes 리소스 타입 가져오기 (인터페이스 구현)
func (c *K8sClientWrapper) GetResourceTypes(namespaced bool) ([]string, error) {
	// 모든 리소스 타입 가져오기 (CRD 포함)
	return c.GetAllResourceTypes()
}

// GetResourcesBatch 배치로 리소스 가져오기 (인터페이스 구현)
func (c *K8sClientWrapper) GetResourcesBatch(resourceTypes []string, namespace string) ([]map[string]interface{}, error) {
	var allResources []map[string]interface{}
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, resourceType := range resourceTypes {
		wg.Add(1)
		go func(rt string) {
			defer wg.Done()

			resources, err := c.GetResources(rt, namespace)
			if err != nil {
				// 에러 무시 (권한 없는 리소스 등)
				return
			}

			mu.Lock()
			allResources = append(allResources, resources...)
			mu.Unlock()
		}(resourceType)
	}

	wg.Wait()
	return allResources, nil
}

// GetResources 특정 리소스 타입의 리소스 가져오기 (인터페이스 구현)
func (c *K8sClientWrapper) GetResources(resourceType, namespace string) ([]map[string]interface{}, error) {
	resources, err := c.GetResourcesInNamespace(resourceType, namespace)
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for _, res := range resources {
		result = append(result, res.UnstructuredContent())
	}
	return result, nil
}

// ValidateNamespacesBatch 네임스페이스 유효성 검증 (인터페이스 구현)
func (c *K8sClientWrapper) ValidateNamespacesBatch(namespaces []string) (map[string]bool, error) {
	result := make(map[string]bool)

	// 현재 존재하는 모든 네임스페이스 가져오기
	existingNamespaces, err := c.GetAllNamespaces()
	if err != nil {
		return nil, err
	}

	// 맵으로 변환
	existingMap := make(map[string]bool)
	for _, ns := range existingNamespaces {
		existingMap[ns] = true
	}

	// 검증
	for _, ns := range namespaces {
		result[ns] = existingMap[ns]
	}

	return result, nil
}
