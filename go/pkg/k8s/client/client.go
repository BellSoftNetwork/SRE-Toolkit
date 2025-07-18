package client

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/flowcontrol"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
)

type ClientConfig struct {
	ImportantResourceTypes []string
	SkipResourceTypes      map[string]bool
}

type Client struct {
	config                     *ClientConfig
	clientset                  kubernetes.Interface
	dynamicClient              dynamic.Interface
	discoveryClient            discovery.DiscoveryInterface
	cachedResourceTypes        []metav1.APIResource
	cachedResourceTypesOnce    sync.Once
	cachedAPIResourceLists     []*metav1.APIResourceList
	cachedAPIResourceListsOnce sync.Once
	cachedNamespaces           []string
	cachedNamespacesTime       time.Time
	namespacesMutex            sync.RWMutex
	emptyResourceCache         sync.Map // namespace:resourceType -> bool
}

func init() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "false")
	flag.Set("alsologtostderr", "false")
	flag.Set("v", "0")
}

func NewClient(cfg *ClientConfig) (*Client, error) {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	if kubeconfigEnv := clientcmd.NewDefaultClientConfigLoadingRules().ExplicitPath; kubeconfigEnv != "" {
		kubeconfig = kubeconfigEnv
	}

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("kubeconfig 로드 실패: %w", err)
		}
	}

	restConfig.QPS = 0
	restConfig.Burst = 0
	restConfig.Timeout = 60 * time.Second

	restConfig.RateLimiter = flowcontrol.NewFakeAlwaysRateLimiter()

	restConfig.DisableCompression = true
	restConfig.UserAgent = "argus/1.0"

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("clientset 생성 실패: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("dynamic client 생성 실패: %w", err)
	}

	return &Client{
		config:          cfg,
		clientset:       clientset,
		dynamicClient:   dynamicClient,
		discoveryClient: clientset.Discovery(),
	}, nil
}

func (c *Client) GetCurrentContext() (string, string) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{},
	)
	rawConfig, _ := clientConfig.RawConfig()
	context := rawConfig.CurrentContext

	if currentContext, exists := rawConfig.Contexts[context]; exists {
		return context, currentContext.Cluster
	}

	return context, ""
}

func (c *Client) GetAllNamespaces() ([]string, error) {
	c.namespacesMutex.RLock()
	if time.Since(c.cachedNamespacesTime) < 5*time.Minute && len(c.cachedNamespaces) > 0 {
		namespaces := make([]string, len(c.cachedNamespaces))
		copy(namespaces, c.cachedNamespaces)
		c.namespacesMutex.RUnlock()
		return namespaces, nil
	}
	c.namespacesMutex.RUnlock()

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

	c.namespacesMutex.Lock()
	c.cachedNamespaces = namespaces
	c.cachedNamespacesTime = time.Now()
	c.namespacesMutex.Unlock()

	return namespaces, nil
}

func (c *Client) getAPIResourceLists() ([]*metav1.APIResourceList, error) {
	var err error
	c.cachedAPIResourceListsOnce.Do(func() {
		_, apiResourceLists, e := c.discoveryClient.ServerGroupsAndResources()
		if e != nil {
			err = e
			return
		}
		c.cachedAPIResourceLists = apiResourceLists
	})

	return c.cachedAPIResourceLists, err
}

func (c *Client) GetResourceTypes(namespaced bool) ([]string, error) {
	if len(c.config.ImportantResourceTypes) > 0 {
		return c.config.ImportantResourceTypes, nil
	}

	var err error
	c.cachedResourceTypesOnce.Do(func() {
		apiResourceLists, e := c.getAPIResourceLists()
		if e != nil {
			err = e
			return
		}

		var resources []metav1.APIResource
		for _, apiResourceList := range apiResourceLists {
			for _, apiResource := range apiResourceList.APIResources {
				hasListVerb := false
				for _, verb := range apiResource.Verbs {
					if verb == "list" {
						hasListVerb = true
						break
					}
				}
				if !hasListVerb {
					continue
				}

				if namespaced && !apiResource.Namespaced {
					continue
				}

				resourceName := apiResource.Name
				if apiResource.Group != "" {
					resourceName = fmt.Sprintf("%s.%s", apiResource.Name, apiResource.Group)
				}
				if c.config.SkipResourceTypes[resourceName] || c.config.SkipResourceTypes[apiResource.Name] {
					continue
				}

				resources = append(resources, apiResource)
			}
		}
		c.cachedResourceTypes = resources
	})

	if err != nil {
		return nil, err
	}

	resourceMap := make(map[string]bool)
	for _, r := range c.cachedResourceTypes {
		if r.Group != "" {
			resourceMap[fmt.Sprintf("%s.%s", r.Name, r.Group)] = true
		} else {
			resourceMap[r.Name] = true
		}
	}

	var resourceNames []string
	for name := range resourceMap {
		resourceNames = append(resourceNames, name)
	}

	return resourceNames, nil
}

func (c *Client) GetResourcesBatch(resourceTypes []string, namespace string) ([]map[string]interface{}, error) {
	var allResources []map[string]interface{}
	var mu sync.Mutex
	var wg sync.WaitGroup

	seenUIDs := make(map[string]bool)

	maxWorkers := 20
	if len(resourceTypes) < maxWorkers {
		maxWorkers = len(resourceTypes)
	}
	if len(resourceTypes) < 10 {
		maxWorkers = len(resourceTypes)
	}
	if len(resourceTypes) > 50 && maxWorkers > 10 {
		maxWorkers = 10
	}

	workChan := make(chan string, len(resourceTypes))
	for _, rt := range resourceTypes {
		workChan <- rt
	}
	close(workChan)

	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for resourceType := range workChan {
				resources, err := c.GetResources(resourceType, namespace)
				if err != nil {
					continue
				}
				if len(resources) > 0 {
					localResources := make([]map[string]interface{}, 0, len(resources))
					localUIDs := make(map[string]bool)

					for _, resource := range resources {
						if metadata, ok := resource["metadata"].(map[string]interface{}); ok {
							if uid, ok := metadata["uid"].(string); ok {
								if !localUIDs[uid] {
									localUIDs[uid] = true
									localResources = append(localResources, resource)
								}
							} else {
								localResources = append(localResources, resource)
							}
						} else {
							localResources = append(localResources, resource)
						}
					}

					if len(localResources) > 0 {
						mu.Lock()
						for _, res := range localResources {
							if metadata, ok := res["metadata"].(map[string]interface{}); ok {
								if uid, ok := metadata["uid"].(string); ok {
									if !seenUIDs[uid] {
										seenUIDs[uid] = true
										allResources = append(allResources, res)
									}
								}
							}
						}
						mu.Unlock()
					}
				}
			}
		}()
	}

	wg.Wait()
	return allResources, nil
}

func (c *Client) GetResources(resourceType, namespace string) ([]map[string]interface{}, error) {
	cacheKey := fmt.Sprintf("%s:%s", namespace, resourceType)
	if empty, ok := c.emptyResourceCache.Load(cacheKey); ok && empty.(bool) {
		return nil, nil
	}

	var gvr schema.GroupVersionResource

	apiResource, gv, err := c.findAPIResource(resourceType)
	if err != nil {
		return nil, nil
	}

	gvr.Resource = apiResource.Name
	gvr.Group = gv.Group
	gvr.Version = gv.Version

	var list *unstructured.UnstructuredList
	maxRetries := 3
	retryDelay := time.Second

	for i := 0; i < maxRetries; i++ {
		timeout := time.Duration(20+i*10) * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)

		listOpts := metav1.ListOptions{
			Limit: 500,
		}

		var allItems []unstructured.Unstructured
		for {
			tempList, tempErr := c.dynamicClient.Resource(gvr).Namespace(namespace).List(ctx, listOpts)
			if tempErr != nil {
				err = tempErr
				break
			}

			allItems = append(allItems, tempList.Items...)

			if tempList.GetContinue() == "" {
				list = &unstructured.UnstructuredList{
					Object: tempList.Object,
					Items:  allItems,
				}
				err = nil
				break
			}

			listOpts.Continue = tempList.GetContinue()
		}

		cancel()

		if err == nil {
			break
		}

		if ctx.Err() != nil || strings.Contains(err.Error(), "timeout") ||
			strings.Contains(err.Error(), "deadline exceeded") ||
			strings.Contains(err.Error(), "connection reset") ||
			strings.Contains(err.Error(), "temporary") {
			if i < maxRetries-1 {
				time.Sleep(retryDelay)
				retryDelay *= 2
				continue
			}
			return nil, nil
		}

		return nil, err
	}

	if len(list.Items) == 0 {
		c.emptyResourceCache.Store(cacheKey, true)
		return nil, nil
	}

	var resources []map[string]interface{}
	for _, item := range list.Items {
		resources = append(resources, item.Object)
	}

	return resources, nil
}

func (c *Client) findAPIResource(resourceType string) (*metav1.APIResource, *schema.GroupVersion, error) {
	apiResourceLists, err := c.getAPIResourceLists()
	if err != nil {
		return nil, nil, err
	}

	for _, apiResourceList := range apiResourceLists {
		gv, _ := schema.ParseGroupVersion(apiResourceList.GroupVersion)
		for _, apiResource := range apiResourceList.APIResources {
			fullName := apiResource.Name
			if gv.Group != "" {
				fullName = fmt.Sprintf("%s.%s", apiResource.Name, gv.Group)
			}

			if fullName == resourceType || apiResource.Name == resourceType {
				return &apiResource, &gv, nil
			}
		}
	}

	return nil, nil, fmt.Errorf("리소스 타입을 찾을 수 없음: %s", resourceType)
}

func (c *Client) ValidateNamespacesBatch(namespaces []string) (map[string]bool, error) {
	allNamespaces, err := c.GetAllNamespaces()
	if err != nil {
		return nil, err
	}

	existingNs := make(map[string]bool)
	for _, ns := range allNamespaces {
		existingNs[ns] = true
	}

	result := make(map[string]bool)
	for _, ns := range namespaces {
		result[ns] = existingNs[ns]
	}

	return result, nil
}
