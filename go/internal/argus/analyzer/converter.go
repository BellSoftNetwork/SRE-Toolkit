package analyzer

import (
	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/argus/config"
	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/argus/domain"
)

func MapToResource(obj map[string]interface{}, namespace string, cfg *config.Config) *domain.KubernetesResource {
	metadata, ok := obj["metadata"].(map[string]interface{})
	if !ok {
		return nil
	}

	resource := &domain.KubernetesResource{
		Identifier: domain.ResourceIdentifier{
			APIVersion: getString(obj, "apiVersion"),
			Kind:       getString(obj, "kind"),
			Name:       getString(metadata, "name"),
			Namespace:  getStringOr(metadata, "namespace", namespace),
		},
		CreatedAt:       getString(metadata, "creationTimestamp"),
		Labels:          getStringMap(metadata, "labels"),
		Annotations:     getStringMap(metadata, "annotations"),
		OwnerReferences: getSlice(metadata, "ownerReferences"),
		Config:          cfg,
	}

	return resource
}

// 유틸리티 함수들
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getStringOr(m map[string]interface{}, key, defaultValue string) string {
	if v := getString(m, key); v != "" {
		return v
	}
	return defaultValue
}

func getStringMap(m map[string]interface{}, key string) map[string]string {
	result := make(map[string]string)
	if v, ok := m[key].(map[string]interface{}); ok {
		for k, val := range v {
			if s, ok := val.(string); ok {
				result[k] = s
			}
		}
	}
	return result
}

func getSlice(m map[string]interface{}, key string) []interface{} {
	if v, ok := m[key].([]interface{}); ok {
		return v
	}
	return nil
}
