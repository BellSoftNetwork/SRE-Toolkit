package k8sinterface

type K8sClient interface {
	GetCurrentContext() (string, string)
	GetAllNamespaces() ([]string, error)
	GetResourceTypes(namespaced bool) ([]string, error)
	GetResourcesBatch(resourceTypes []string, namespace string) ([]map[string]interface{}, error)
	GetResources(resourceType, namespace string) ([]map[string]interface{}, error)
	ValidateNamespacesBatch(namespaces []string) (map[string]bool, error)
}
