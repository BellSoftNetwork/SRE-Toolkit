package reporter

import (
	"gitlab.bellsoft.net/devops/sre-toolkit/go/internal/argus/domain"
	"time"
)

type Reporter interface {
	Generate(allResults map[string]domain.AnalysisResult, context, cluster string, startTime time.Time) error
}
