package domain

import "gitlab.bellsoft.net/devops/sre-toolkit/go/internal/argus/config"

type ResourceIdentifier struct {
	APIVersion string
	Kind       string
	Name       string
	Namespace  string
}

type KubernetesResource struct {
	Identifier      ResourceIdentifier
	CreatedAt       string
	Labels          map[string]string
	Annotations     map[string]string
	OwnerReferences []interface{}
	Config          *config.Config
}

func (r *KubernetesResource) IsRootResource() bool {
	return len(r.OwnerReferences) == 0
}

func (r *KubernetesResource) IsArgoCDManaged() bool {
	if r.Config != nil {
		for _, label := range r.Config.GetManagedLabels() {
			if _, ok := r.Labels[label]; ok {
				return true
			}
		}
		for _, ann := range r.Config.GetSyncAnnotations() {
			if _, ok := r.Annotations[ann]; ok {
				return true
			}
		}
	}
	if _, ok := r.Labels["argocd.argoproj.io/instance"]; ok {
		return true
	}
	if _, ok := r.Annotations["argocd.argoproj.io/tracking-id"]; ok {
		return true
	}
	return false
}

type AnalysisResult struct {
	TotalResources     int
	RootResources      int
	ArgoCDManaged      int
	ManualResources    int
	ExcludedDefaults   int
	ManualResourceList []KubernetesResource
	ArgoCDResourceList []KubernetesResource
}

type NamespaceAnalysis struct {
	Namespace string
	Result    AnalysisResult
	Error     error
}
