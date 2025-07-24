package analyzer

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/k8s-diff/config"
	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/k8s-diff/domain"
)

// Analyzer 리소스 분석기
type Analyzer struct {
	config *config.Config
}

// NewAnalyzer 새 분석기 생성
func NewAnalyzer(cfg *config.Config) *Analyzer {
	return &Analyzer{
		config: cfg,
	}
}

// CompareResources 두 클러스터의 리소스 비교
func (a *Analyzer) CompareResources(sourceResources, targetResources []domain.KubernetesResource) domain.ComparisonResult {
	sourceMap := a.createResourceMap(sourceResources)
	targetMap := a.createResourceMap(targetResources)

	result := domain.ComparisonResult{
		TotalSource: len(sourceResources),
		TotalTarget: len(targetResources),
	}

	// 소스에만 있는 리소스 찾기
	for key, resource := range sourceMap {
		if _, exists := targetMap[key]; !exists {
			result.OnlyInSource = append(result.OnlyInSource, resource)
		}
	}

	// 타겟에만 있는 리소스 찾기
	for key, resource := range targetMap {
		if _, exists := sourceMap[key]; !exists {
			result.OnlyInTarget = append(result.OnlyInTarget, resource)
		}
	}

	// 수정된 리소스 찾기 (향후 구현 예정)
	if a.config.CompareResourceContents {
		for key, sourceResource := range sourceMap {
			if targetResource, exists := targetMap[key]; exists {
				if sourceResource.ResourceHash != targetResource.ResourceHash {
					result.ModifiedResources = append(result.ModifiedResources, domain.ResourceDiff{
						Resource:   sourceResource,
						SourceHash: sourceResource.ResourceHash,
						TargetHash: targetResource.ResourceHash,
					})
				}
			}
		}
	}

	return result
}

// createResourceMap 리소스 맵 생성
func (a *Analyzer) createResourceMap(resources []domain.KubernetesResource) map[string]domain.KubernetesResource {
	resourceMap := make(map[string]domain.KubernetesResource)
	for _, resource := range resources {
		var key string
		if a.config.StrictAPIVersion {
			// 정밀 분석: API 버전 포함
			key = resource.ResourceKey()
		} else {
			// 기본: Kind만 비교
			key = resource.ResourceKeyWithoutAPIVersion()
		}
		resourceMap[key] = resource
	}
	return resourceMap
}

// IsExcluded 리소스 제외 여부 확인
func (a *Analyzer) IsExcluded(resource domain.KubernetesResource) bool {
	// 제외 규칙 확인
	for _, rule := range a.config.ExclusionRules {
		if rule.Match(resource.Namespace, resource.Kind, resource.Name) {
			return true
		}
	}

	// 부모가 있는 리소스 제외 (최상위 리소스만 비교)
	if a.shouldExcludeByOwner(resource) {
		return true
	}

	// 패턴 기반 제외
	if a.shouldExcludeByPattern(resource) {
		return true
	}

	return false
}

// shouldExcludeByOwner 부모가 있는 리소스 제외
func (a *Analyzer) shouldExcludeByOwner(resource domain.KubernetesResource) bool {
	// Job이 CronJob에 의해 생성된 경우 제외
	if resource.Kind == "Job" && resource.IsOwnedByKind("CronJob") {
		return true
	}

	// ReplicaSet이 Deployment에 의해 생성된 경우 제외
	if resource.Kind == "ReplicaSet" && resource.IsOwnedByKind("Deployment") {
		return true
	}

	// Pod는 항상 제외 (직접 생성된 것이 아닌 경우가 대부분)
	if resource.Kind == "Pod" && resource.HasOwnerReference() {
		return true
	}

	return false
}

// shouldExcludeByPattern 패턴 기반 제외
func (a *Analyzer) shouldExcludeByPattern(resource domain.KubernetesResource) bool {
	// rb- 로 시작하는 RoleBinding 제외
	if resource.Kind == "RoleBinding" && strings.HasPrefix(resource.Name, "rb-") {
		return true
	}

	// Rancher 관련 패턴
	if resource.Kind == "RoleBinding" || resource.Kind == "ClusterRoleBinding" {
		if strings.HasPrefix(resource.Name, "rb-") ||
			strings.HasPrefix(resource.Name, "crb-") ||
			strings.HasPrefix(resource.Name, "psp:") {
			return true
		}
	}

	// default-token- 으로 시작하는 Secret 제외
	if resource.Kind == "Secret" && strings.HasPrefix(resource.Name, "default-token-") {
		return true
	}

	// sh.helm.release 로 시작하는 Secret 제외 (Helm 관련)
	if resource.Kind == "Secret" && strings.HasPrefix(resource.Name, "sh.helm.release.") {
		return true
	}

	// VerticalPodAutoscalerCheckpoint 제외
	if resource.Kind == "VerticalPodAutoscalerCheckpoint" {
		return true
	}

	return false
}

// MapToResource k8s API 응답을 도메인 객체로 변환
func MapToResource(data map[string]interface{}, namespace string, cfg *config.Config) *domain.KubernetesResource {
	metadata, ok := data["metadata"].(map[string]interface{})
	if !ok {
		return nil
	}

	name, _ := metadata["name"].(string)
	if name == "" {
		return nil
	}

	kind, _ := data["kind"].(string)
	apiVersion, _ := data["apiVersion"].(string)

	resource := &domain.KubernetesResource{
		Namespace:       namespace,
		Kind:            kind,
		APIVersion:      apiVersion,
		Name:            name,
		Labels:          extractStringMap(metadata["labels"]),
		Annotations:     extractStringMap(metadata["annotations"]),
		OwnerReferences: extractOwnerReferences(metadata["ownerReferences"]),
	}

	if uid, ok := metadata["uid"].(string); ok {
		resource.UID = uid
	}

	if creationTimestamp, ok := metadata["creationTimestamp"].(string); ok {
		if t, err := time.Parse(time.RFC3339, creationTimestamp); err == nil {
			resource.CreationTime = t
		}
	}

	// 리소스 해시 계산
	if cfg.CompareResourceContents {
		resource.ResourceHash = calculateResourceHash(data)
	}

	return resource
}

// extractStringMap 인터페이스 맵에서 문자열 맵 추출
func extractStringMap(data interface{}) map[string]string {
	result := make(map[string]string)
	if m, ok := data.(map[string]interface{}); ok {
		for k, v := range m {
			if str, ok := v.(string); ok {
				result[k] = str
			}
		}
	}
	return result
}

// calculateResourceHash 리소스의 해시값 계산
func calculateResourceHash(data map[string]interface{}) string {
	// metadata에서 변경 가능한 필드들 제거
	cleanData := make(map[string]interface{})
	for k, v := range data {
		if k != "metadata" {
			cleanData[k] = v
		} else if metadata, ok := v.(map[string]interface{}); ok {
			cleanMetadata := make(map[string]interface{})
			for mk, mv := range metadata {
				// 자주 변경되는 메타데이터 필드 제외
				if mk != "resourceVersion" && mk != "uid" && mk != "generation" &&
					mk != "creationTimestamp" && mk != "managedFields" {
					cleanMetadata[mk] = mv
				}
			}
			if len(cleanMetadata) > 0 {
				cleanData[k] = cleanMetadata
			}
		}
	}

	// 안정적인 JSON 직렬화를 위해 키 정렬
	jsonData, _ := json.Marshal(sortKeys(cleanData))
	hash := sha256.Sum256(jsonData)
	return fmt.Sprintf("%x", hash)
}

// sortKeys 맵의 키를 정렬하여 일관된 순서 보장
func sortKeys(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		sorted := make(map[string]interface{})
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			sorted[k] = sortKeys(v[k])
		}
		return sorted
	case []interface{}:
		sorted := make([]interface{}, len(v))
		for i, item := range v {
			sorted[i] = sortKeys(item)
		}
		return sorted
	default:
		return v
	}
}

// extractOwnerReferences ownerReferences 추출
func extractOwnerReferences(data interface{}) []domain.OwnerReference {
	var result []domain.OwnerReference
	if owners, ok := data.([]interface{}); ok {
		for _, owner := range owners {
			if ownerMap, ok := owner.(map[string]interface{}); ok {
				ref := domain.OwnerReference{}
				if apiVersion, ok := ownerMap["apiVersion"].(string); ok {
					ref.APIVersion = apiVersion
				}
				if kind, ok := ownerMap["kind"].(string); ok {
					ref.Kind = kind
				}
				if name, ok := ownerMap["name"].(string); ok {
					ref.Name = name
				}
				if uid, ok := ownerMap["uid"].(string); ok {
					ref.UID = uid
				}
				result = append(result, ref)
			}
		}
	}
	return result
}
