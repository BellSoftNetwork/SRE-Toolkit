package k8sclient

import (
	"fmt"
	"os"
	"path/filepath"

	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/k8s-diff/config"
	"gitlab.bellsoft.net/devops/sre-workbench/go/pkg/k8s/client"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// ClientWrapper k8s 클라이언트 래퍼
type ClientWrapper struct {
	*client.Client
	context string
}

// NewClientWithContext 특정 컨텍스트로 클라이언트 생성
func NewClientWithContext(cfg *config.Config, context string) (*ClientWrapper, error) {
	// SkipResourceTypes를 map으로 변환
	skipMap := make(map[string]bool)
	for _, rt := range cfg.SkipResourceTypes {
		skipMap[rt] = true
	}

	// kubeconfig 파일 경로 결정
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	if kubeconfigEnv := os.Getenv("KUBECONFIG"); kubeconfigEnv != "" {
		kubeconfig = kubeconfigEnv
	}

	// 임시 kubeconfig 파일 생성
	tempKubeconfig, err := createTempKubeconfigForContext(kubeconfig, context)
	if err != nil {
		return nil, fmt.Errorf("임시 kubeconfig 생성 실패: %w", err)
	}
	defer os.Remove(tempKubeconfig)

	// 임시 kubeconfig로 환경변수 설정
	originalKubeconfig := os.Getenv("KUBECONFIG")
	os.Setenv("KUBECONFIG", tempKubeconfig)
	defer os.Setenv("KUBECONFIG", originalKubeconfig)

	clientConfig := &client.ClientConfig{
		ImportantResourceTypes: cfg.ImportantResourceTypes,
		SkipResourceTypes:      skipMap,
	}

	k8sClient, err := client.NewClient(clientConfig)
	if err != nil {
		return nil, err
	}

	return &ClientWrapper{
		Client:  k8sClient,
		context: context,
	}, nil
}

// createTempKubeconfigForContext 특정 컨텍스트만 있는 임시 kubeconfig 생성
func createTempKubeconfigForContext(kubeconfigPath, contextName string) (string, error) {
	// 기존 kubeconfig 로드
	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return "", fmt.Errorf("kubeconfig 로드 실패: %w", err)
	}

	// 해당 컨텍스트가 있는지 확인
	if _, exists := config.Contexts[contextName]; !exists {
		return "", fmt.Errorf("컨텍스트 '%s'를 찾을 수 없습니다", contextName)
	}

	// 현재 컨텍스트를 원하는 컨텍스트로 설정
	config.CurrentContext = contextName

	// 임시 파일에 저장
	tempFile, err := os.CreateTemp("", "kubeconfig-*.yaml")
	if err != nil {
		return "", fmt.Errorf("임시 파일 생성 실패: %w", err)
	}
	tempFile.Close()

	if err := clientcmd.WriteToFile(*config, tempFile.Name()); err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("kubeconfig 저장 실패: %w", err)
	}

	return tempFile.Name(), nil
}

// GetContext 클라이언트의 컨텍스트 반환
func (c *ClientWrapper) GetContext() string {
	return c.context
}
