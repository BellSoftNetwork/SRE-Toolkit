package domain

import (
	"fmt"
	"time"
)

// KubernetesResource k8s 리소스 정보
type KubernetesResource struct {
	Namespace       string
	Kind            string
	APIVersion      string
	Name            string
	Labels          map[string]string
	Annotations     map[string]string
	CreationTime    time.Time
	UID             string
	ResourceHash    string           // 리소스의 내용을 기반으로 한 해시값
	OwnerReferences []OwnerReference // 부모 리소스 정보
}

// OwnerReference 리소스의 소유자 정보
type OwnerReference struct {
	APIVersion string
	Kind       string
	Name       string
	UID        string
}

// ResourceKey 리소스의 고유 키 생성
func (r *KubernetesResource) ResourceKey() string {
	return fmt.Sprintf("%s/%s/%s/%s", r.Namespace, r.APIVersion, r.Kind, r.Name)
}

// ResourceKeyWithoutAPIVersion API 버전 제외한 리소스 키 생성
func (r *KubernetesResource) ResourceKeyWithoutAPIVersion() string {
	return fmt.Sprintf("%s/%s/%s", r.Namespace, r.Kind, r.Name)
}

// HasOwnerReference 부모 리소스가 있는지 확인
func (r *KubernetesResource) HasOwnerReference() bool {
	return len(r.OwnerReferences) > 0
}

// IsOwnedByKind 특정 Kind의 부모가 있는지 확인
func (r *KubernetesResource) IsOwnedByKind(kind string) bool {
	for _, owner := range r.OwnerReferences {
		if owner.Kind == kind {
			return true
		}
	}
	return false
}

// ComparisonResult 비교 결과
type ComparisonResult struct {
	OnlyInSource      []KubernetesResource // 소스에만 있는 리소스
	OnlyInTarget      []KubernetesResource // 타겟에만 있는 리소스
	ModifiedResources []ResourceDiff       // 수정된 리소스
	TotalSource       int
	TotalTarget       int
}

// ResourceDiff 리소스 차이점
type ResourceDiff struct {
	Resource        KubernetesResource
	SourceHash      string
	TargetHash      string
	DifferentFields []string
}

// NamespaceComparison 네임스페이스별 비교 결과
type NamespaceComparison struct {
	Namespace string
	Result    ComparisonResult
	Error     error
}

// ClusterInfo 클러스터 정보
type ClusterInfo struct {
	Context string
	Name    string
}
