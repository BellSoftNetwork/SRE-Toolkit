package reporter

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/k8s-diff/domain"
	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/k8s-diff/utils"
)

// MarkdownReporter 마크다운 리포터
type MarkdownReporter struct {
	outputDir string
}

// NewMarkdownReporter 새 마크다운 리포터 생성
func NewMarkdownReporter(outputDir string) *MarkdownReporter {
	return &MarkdownReporter{
		outputDir: outputDir,
	}
}

// Generate 마크다운 리포트 생성
func (r *MarkdownReporter) Generate(results map[string]domain.ComparisonResult, sourceCluster, targetCluster domain.ClusterInfo) error {
	// 출력 디렉토리 생성
	if err := os.MkdirAll(r.outputDir, 0755); err != nil {
		return fmt.Errorf("출력 디렉토리 생성 실패: %w", err)
	}

	// 마크다운 파일 생성
	timestamp := time.Now().Format("20060102_150405")
	filename := filepath.Join(r.outputDir, fmt.Sprintf("%s.md", timestamp))

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("마크다운 파일 생성 실패: %w", err)
	}
	defer file.Close()

	// 리포트 작성
	r.writeHeader(file, sourceCluster, targetCluster)
	r.writeSummary(file, results)
	r.writeNamespaceDetails(file, results)
	r.writeFooter(file)

	fmt.Printf("\n✅ Markdown 리포트 생성됨: %s\n", filename)
	return nil
}

// writeHeader 헤더 작성
func (r *MarkdownReporter) writeHeader(file *os.File, sourceCluster, targetCluster domain.ClusterInfo) {
	fmt.Fprintf(file, "# Kubernetes 클러스터 비교 리포트\n\n")
	fmt.Fprintf(file, "**생성 시간:** %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, "## 클러스터 정보\n\n")
	fmt.Fprintf(file, "| 구분 | 클러스터 | 컨텍스트 |\n")
	fmt.Fprintf(file, "|------|----------|----------|\n")
	fmt.Fprintf(file, "| **소스** | %s | %s |\n", utils.ExtractClusterName(sourceCluster.Name), sourceCluster.Context)
	fmt.Fprintf(file, "| **타겟** | %s | %s |\n\n", utils.ExtractClusterName(targetCluster.Name), targetCluster.Context)
}

// writeSummary 요약 정보 작성
func (r *MarkdownReporter) writeSummary(file *os.File, results map[string]domain.ComparisonResult) {
	totalSourceOnly, totalTargetOnly, totalModified := 0, 0, 0

	for _, result := range results {
		totalSourceOnly += len(result.OnlyInSource)
		totalTargetOnly += len(result.OnlyInTarget)
		totalModified += len(result.ModifiedResources)
	}

	fmt.Fprintf(file, "## 전체 요약\n\n")
	fmt.Fprintf(file, "| 항목 | 개수 |\n")
	fmt.Fprintf(file, "|------|------|\n")
	fmt.Fprintf(file, "| 소스에만 있는 리소스 | %d |\n", totalSourceOnly)
	fmt.Fprintf(file, "| 타겟에만 있는 리소스 | %d |\n", totalTargetOnly)
	if totalModified > 0 {
		fmt.Fprintf(file, "| 수정된 리소스 | %d |\n", totalModified)
	}
	fmt.Fprintf(file, "\n")
}

// writeNamespaceDetails 네임스페이스별 상세 정보 작성
func (r *MarkdownReporter) writeNamespaceDetails(file *os.File, results map[string]domain.ComparisonResult) {
	fmt.Fprintf(file, "## 네임스페이스별 리소스 비교\n\n")

	// 네임스페이스 정렬
	namespaces := make([]string, 0, len(results))
	for ns := range results {
		namespaces = append(namespaces, ns)
	}
	sort.Strings(namespaces)

	for _, ns := range namespaces {
		result := results[ns]

		// 차이가 없는 네임스페이스는 스킵
		if len(result.OnlyInSource) == 0 && len(result.OnlyInTarget) == 0 && len(result.ModifiedResources) == 0 {
			continue
		}

		fmt.Fprintf(file, "### 네임스페이스: %s\n\n", ns)

		// 모든 리소스를 하나의 맵으로 병합
		allResources := make(map[string]resourceComparison)

		// 소스에만 있는 리소스
		for _, res := range result.OnlyInSource {
			key := fmt.Sprintf("%s/%s", res.Kind, res.Name)
			allResources[key] = resourceComparison{
				Resource: res,
				Status:   "source-only",
			}
		}

		// 타겟에만 있는 리소스
		for _, res := range result.OnlyInTarget {
			key := fmt.Sprintf("%s/%s", res.Kind, res.Name)
			allResources[key] = resourceComparison{
				Resource: res,
				Status:   "target-only",
			}
		}

		// 리소스 키 정렬
		var keys []string
		for k := range allResources {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// 테이블 작성
		fmt.Fprintf(file, "| 타입 | 이름 | API 버전 | 상태 | 생성 시간 |\n")
		fmt.Fprintf(file, "|------|------|----------|------|----------|\n")

		for _, key := range keys {
			comp := allResources[key]
			res := comp.Resource

			statusIcon := r.getStatusIcon(comp.Status)
			creationTime := ""
			if !res.CreationTime.IsZero() {
				creationTime = res.CreationTime.Format("2006-01-02 15:04:05")
			}

			fmt.Fprintf(file, "| %s | %s | %s | %s | %s |\n",
				res.Kind,
				res.Name,
				res.APIVersion,
				statusIcon,
				creationTime,
			)
		}
		fmt.Fprintf(file, "\n")
	}
}

// getStatusIcon 상태에 따른 아이콘 반환
func (r *MarkdownReporter) getStatusIcon(status string) string {
	switch status {
	case "source-only":
		return "🔴 소스만"
	case "target-only":
		return "🟢 타겟만"
	case "modified":
		return "🔵 수정됨"
	default:
		return status
	}
}

// resourceComparison 리소스 비교 정보
type resourceComparison struct {
	Resource domain.KubernetesResource
	Status   string
}

// writeFooter 푸터 작성
func (r *MarkdownReporter) writeFooter(file *os.File) {
	fmt.Fprintf(file, "---\n\n")
	fmt.Fprintf(file, "*이 리포트는 [K8s-Diff](https://gitlab.bellsoft.net/devops/sre-workbench)에 의해 자동 생성되었습니다.*\n")
}
