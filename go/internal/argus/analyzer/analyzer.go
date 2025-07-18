package analyzer

import (
	"gitlab.bellsoft.net/devops/sre-toolkit/go/internal/argus/config"
	"gitlab.bellsoft.net/devops/sre-toolkit/go/internal/argus/domain"
)

type Analyzer struct {
	config *config.Config
}

func NewAnalyzer(cfg *config.Config) *Analyzer {
	return &Analyzer{config: cfg}
}

func (a *Analyzer) AnalyzeResources(resources []domain.KubernetesResource) domain.AnalysisResult {
	result := domain.AnalysisResult{
		TotalResources:     len(resources),
		ManualResourceList: []domain.KubernetesResource{},
		ArgoCDResourceList: []domain.KubernetesResource{},
	}

	for _, resource := range resources {
		if !resource.IsRootResource() {
			continue
		}

		result.RootResources++

		if a.shouldExcludeResource(&resource) {
			result.ExcludedDefaults++
			continue
		}

		a.categorizeResource(&resource, &result)
	}

	return result
}

func (a *Analyzer) categorizeResource(resource *domain.KubernetesResource, result *domain.AnalysisResult) {
	if resource.IsArgoCDManaged() {
		result.ArgoCDManaged++
		result.ArgoCDResourceList = append(result.ArgoCDResourceList, *resource)
	} else {
		result.ManualResources++
		result.ManualResourceList = append(result.ManualResourceList, *resource)
	}
}

func (a *Analyzer) shouldExcludeResource(resource *domain.KubernetesResource) bool {
	if a.matchesExclusionRule(resource) {
		return true
	}

	if a.isExcludedSecret(resource) {
		return true
	}

	if a.isRancherManaged(resource) {
		return true
	}

	if a.hasAutoManagedAnnotation(resource) {
		return true
	}

	if a.isStatefulSetPVC(resource) {
		return true
	}

	return false
}

func (a *Analyzer) matchesExclusionRule(resource *domain.KubernetesResource) bool {
	for _, rule := range a.config.ExclusionRules {
		if rule.Match(resource.Identifier.Namespace, resource.Identifier.Kind, resource.Identifier.Name) {
			return true
		}
	}
	return false
}

func (a *Analyzer) isExcludedSecret(resource *domain.KubernetesResource) bool {
	if resource.Identifier.Kind != "Secret" {
		return false
	}

	if a.matchesSecretPattern(resource.Identifier.Name) {
		return true
	}

	return a.hasCertManagerAnnotation(resource)
}

func (a *Analyzer) matchesSecretPattern(name string) bool {
	for _, pattern := range a.config.SecretPatterns {
		if pattern.MatchString(name) {
			return true
		}
	}
	return false
}

func (a *Analyzer) hasCertManagerAnnotation(resource *domain.KubernetesResource) bool {
	for annotation := range a.config.CertManagerAnnotations {
		if _, ok := resource.Annotations[annotation]; ok {
			return true
		}
	}
	return false
}

func (a *Analyzer) isRancherManaged(resource *domain.KubernetesResource) bool {
	patterns, ok := a.config.RancherManagedPatterns[resource.Identifier.Kind]
	if !ok {
		return false
	}

	for _, pattern := range patterns {
		if pattern.MatchString(resource.Identifier.Name) {
			return true
		}
	}
	return false
}

func (a *Analyzer) hasAutoManagedAnnotation(resource *domain.KubernetesResource) bool {
	for annotation := range a.config.AutoManagedAnnotations {
		if _, ok := resource.Annotations[annotation]; ok {
			return true
		}
	}
	return false
}

func (a *Analyzer) isStatefulSetPVC(resource *domain.KubernetesResource) bool {
	if resource.Identifier.Kind != "PersistentVolumeClaim" {
		return false
	}

	if a.config.StatefulSetPVCPattern.MatchString(resource.Identifier.Name) {
		return true
	}

	return a.hasStatefulSetLabels(resource)
}

func (a *Analyzer) hasStatefulSetLabels(resource *domain.KubernetesResource) bool {
	_, hasInstance := resource.Labels["app.kubernetes.io/instance"]
	_, hasComponent := resource.Labels["app.kubernetes.io/component"]
	return hasInstance && hasComponent
}
