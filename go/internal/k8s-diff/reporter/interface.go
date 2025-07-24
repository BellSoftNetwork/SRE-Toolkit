package reporter

import (
	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/k8s-diff/domain"
)

// Reporter 리포트 생성 인터페이스
type Reporter interface {
	Generate(results map[string]domain.ComparisonResult, sourceCluster, targetCluster domain.ClusterInfo) error
}
