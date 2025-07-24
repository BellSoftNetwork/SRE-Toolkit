package reporter

import (
	"fmt"
	"sort"
	"strings"

	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/k8s-diff/domain"
	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/k8s-diff/utils"
	"gitlab.bellsoft.net/devops/sre-workbench/go/pkg/utils/color"
)

// ConsoleReporter 콘솔 리포터
type ConsoleReporter struct{}

// NewConsoleReporter 새 콘솔 리포터 생성
func NewConsoleReporter() *ConsoleReporter {
	return &ConsoleReporter{}
}

// Generate 콘솔 리포트 생성
func (r *ConsoleReporter) Generate(results map[string]domain.ComparisonResult, sourceCluster, targetCluster domain.ClusterInfo) error {
	fmt.Printf("\n%s📊 클러스터 비교 요약%s\n", color.Bold, color.NC)
	fmt.Printf("소스 클러스터: %s%s%s\n", color.Cyan, utils.ExtractClusterName(sourceCluster.Name), color.NC)
	fmt.Printf("타겟 클러스터: %s%s%s\n", color.Cyan, utils.ExtractClusterName(targetCluster.Name), color.NC)
	fmt.Println(strings.Repeat("=", 100))

	// 전체 통계
	totalSourceOnly, totalTargetOnly, totalModified := r.calculateTotals(results)
	r.printSummary(totalSourceOnly, totalTargetOnly, totalModified)

	// 네임스페이스별 상세 비교
	r.printNamespaceComparison(results)

	return nil
}

// calculateTotals 전체 통계 계산
func (r *ConsoleReporter) calculateTotals(results map[string]domain.ComparisonResult) (sourceOnly, targetOnly, modified int) {
	for _, result := range results {
		sourceOnly += len(result.OnlyInSource)
		targetOnly += len(result.OnlyInTarget)
		modified += len(result.ModifiedResources)
	}
	return
}

// printSummary 요약 정보 출력
func (r *ConsoleReporter) printSummary(sourceOnly, targetOnly, modified int) {
	fmt.Printf("\n%s전체 요약:%s\n", color.Bold, color.NC)
	fmt.Printf("  • 소스에만 있는 리소스: %s%d개%s\n", color.Yellow, sourceOnly, color.NC)
	fmt.Printf("  • 타겟에만 있는 리소스: %s%d개%s\n", color.Green, targetOnly, color.NC)
	if modified > 0 {
		fmt.Printf("  • 수정된 리소스: %s%d개%s\n", color.Blue, modified, color.NC)
	}
}

// printNamespaceComparison 네임스페이스별 비교 출력
func (r *ConsoleReporter) printNamespaceComparison(results map[string]domain.ComparisonResult) {
	fmt.Printf("\n%s네임스페이스별 리소스 비교:%s\n", color.Bold, color.NC)

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

		fmt.Printf("\n%s[네임스페이스: %s]%s\n", color.Yellow, ns, color.NC)
		fmt.Println(strings.Repeat("-", 100))
		fmt.Printf("%-20s | %-30s | %-15s | %-10s | %-20s\n", "타입", "이름", "API 버전", "상태", "생성 시간")
		fmt.Println(strings.Repeat("-", 100))

		// 모든 리소스를 하나의 슬라이스로 병합
		type resourceEntry struct {
			resource domain.KubernetesResource
			status   string
		}
		var allResources []resourceEntry

		// 소스에만 있는 리소스
		for _, res := range result.OnlyInSource {
			allResources = append(allResources, resourceEntry{
				resource: res,
				status:   "source-only",
			})
		}

		// 타겟에만 있는 리소스
		for _, res := range result.OnlyInTarget {
			allResources = append(allResources, resourceEntry{
				resource: res,
				status:   "target-only",
			})
		}

		// 리소스 정렬
		sort.Slice(allResources, func(i, j int) bool {
			if allResources[i].resource.Kind != allResources[j].resource.Kind {
				return allResources[i].resource.Kind < allResources[j].resource.Kind
			}
			return allResources[i].resource.Name < allResources[j].resource.Name
		})

		// 출력
		for _, entry := range allResources {
			res := entry.resource
			statusDisplay := r.getStatusDisplay(entry.status)
			creationTime := ""
			if !res.CreationTime.IsZero() {
				creationTime = res.CreationTime.Format("2006-01-02 15:04")
			}

			fmt.Printf("%-20s | %-30s | %-15s | %-10s | %-20s\n",
				truncateString(res.Kind, 20),
				truncateString(res.Name, 30),
				truncateString(res.APIVersion, 15),
				statusDisplay,
				creationTime,
			)
		}
	}
}

// getStatusDisplay 상태에 따른 표시 문자열
func (r *ConsoleReporter) getStatusDisplay(status string) string {
	switch status {
	case "source-only":
		return fmt.Sprintf("%s🔴 소스만%s", color.Red, color.NC)
	case "target-only":
		return fmt.Sprintf("%s🟢 타겟만%s", color.Green, color.NC)
	case "modified":
		return fmt.Sprintf("%s🔵 수정됨%s", color.Blue, color.NC)
	default:
		return status
	}
}

// truncateString 문자열 잘라내기
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
