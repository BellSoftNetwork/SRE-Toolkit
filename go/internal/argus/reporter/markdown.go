package reporter

import (
	"fmt"
	"gitlab.bellsoft.net/devops/sre-toolkit/go/internal/argus/domain"
	"os"
	"sort"
	"strings"
	"time"
)

type MarkdownReporter struct {
	reportDir string
}

func NewMarkdownReporter(reportDir string) *MarkdownReporter {
	return &MarkdownReporter{reportDir: reportDir}
}

func (r *MarkdownReporter) Generate(allResults map[string]domain.AnalysisResult, context, cluster string, startTime time.Time) error {
	if err := os.MkdirAll(r.reportDir, 0755); err != nil {
		return fmt.Errorf("보고서 디렉토리 생성 실패: %w", err)
	}

	content := r.generateMarkdownContent(allResults, context, cluster, startTime)

	fileName := fmt.Sprintf("%s/%s.md", r.reportDir, startTime.Format("20060102_150405"))

	if err := os.WriteFile(fileName, []byte(content), 0644); err != nil {
		return fmt.Errorf("보고서 파일 저장 실패: %w", err)
	}

	fmt.Printf("\n📄 마크다운 보고서 저장됨: %s\n", fileName)
	return nil
}

func (r *MarkdownReporter) generateMarkdownContent(allResults map[string]domain.AnalysisResult, context, cluster string, startTime time.Time) string {
	var sb strings.Builder
	elapsed := time.Since(startTime)

	sb.WriteString("# Argus 분석 리포트\n\n")

	sb.WriteString("## 실행 정보\n\n")
	sb.WriteString(fmt.Sprintf("- **생성 시간**: %s\n", startTime.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("- **컨텍스트**: %s\n", context))
	sb.WriteString(fmt.Sprintf("- **클러스터**: %s\n", cluster))
	sb.WriteString(fmt.Sprintf("- **실행 시간**: %.2f초\n\n", elapsed.Seconds()))

	var sortedNamespaces []string
	for ns := range allResults {
		sortedNamespaces = append(sortedNamespaces, ns)
	}
	sort.Strings(sortedNamespaces)

	var unmanagedNamespaces []string
	for _, namespace := range sortedNamespaces {
		result := allResults[namespace]
		if result.ManualResources > 0 {
			unmanagedNamespaces = append(unmanagedNamespaces, namespace)
		}
	}

	r.writeOverallStatistics(&sb, allResults, sortedNamespaces)

	r.writeSummaryTable(&sb, allResults, sortedNamespaces)

	if len(unmanagedNamespaces) > 0 {
		sb.WriteString("## ArgoCD 미관리 네임스페이스\n\n")
		for _, ns := range unmanagedNamespaces {
			result := allResults[ns]
			sb.WriteString(fmt.Sprintf("- **%s**: %d개 수동 리소스\n", ns, result.ManualResources))
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("## ✅ 모든 네임스페이스가 완전히 관리됨\n\n")
		sb.WriteString("모든 리소스가 ArgoCD를 통해 관리되고 있습니다!\n\n")
	}

	totalManual := 0
	for _, result := range allResults {
		totalManual += result.ManualResources
	}

	if totalManual > 0 {
		sb.WriteString("## 💡 권장사항\n\n")
		sb.WriteString("1. 수동 생성된 리소스를 GitOps 워크플로우로 마이그레이션\n")
		sb.WriteString("2. 불필요한 리소스 정리\n")
		sb.WriteString("3. ArgoCD Application 정의에 누락된 리소스 추가\n\n")
	}

	if len(unmanagedNamespaces) > 0 {
		sb.WriteString("## 목차\n\n")
		sb.WriteString("### ArgoCD 미관리 네임스페이스\n")
		for _, ns := range unmanagedNamespaces {
			result := allResults[ns]
			sb.WriteString(fmt.Sprintf("- [%s](#%s) - %d개 수동 리소스\n", ns, ns, result.ManualResources))
		}
		sb.WriteString("\n")
	}

	r.writeManualResourceDetails(&sb, allResults, sortedNamespaces)

	return sb.String()
}

func (r *MarkdownReporter) writeSummaryTable(sb *strings.Builder, allResults map[string]domain.AnalysisResult, sortedNamespaces []string) {
	sb.WriteString("## 최종 결과 요약\n\n")
	sb.WriteString("| 네임스페이스 | ArgoCD 관리 | 전체 리소스 | 최상위 리소스 | ArgoCD 관리 중 | 수동 생성 | 기본 리소스 (제외) |\n")
	sb.WriteString("| --- | --- | --- | --- | --- | --- | --- |\n")

	totalManual := 0
	totalArgoCD := 0
	totalResources := 0
	totalRootResources := 0
	totalExcluded := 0

	for _, namespace := range sortedNamespaces {
		result := allResults[namespace]

		var argoCDManaged string
		if result.RootResources == 0 {
			argoCDManaged = "➖"
		} else if result.ManualResources == 0 {
			argoCDManaged = "✅"
		} else if result.ArgoCDManaged > 0 {
			argoCDManaged = "⚠️"
		} else {
			argoCDManaged = "❌"
		}

		sb.WriteString(fmt.Sprintf("| %s | %s | %d | %d | %d | %d | %d |\n",
			namespace,
			argoCDManaged,
			result.TotalResources,
			result.RootResources,
			result.ArgoCDManaged,
			result.ManualResources,
			result.ExcludedDefaults,
		))

		totalManual += result.ManualResources
		totalArgoCD += result.ArgoCDManaged
		totalResources += result.TotalResources
		totalRootResources += result.RootResources
		totalExcluded += result.ExcludedDefaults
	}

	sb.WriteString(fmt.Sprintf("| **총계** | - | **%d** | **%d** | **%d** | **%d** | **%d** |\n\n",
		totalResources,
		totalRootResources,
		totalArgoCD,
		totalManual,
		totalExcluded,
	))
}

func (r *MarkdownReporter) writeOverallStatistics(sb *strings.Builder, allResults map[string]domain.AnalysisResult, sortedNamespaces []string) {
	totalNamespaces := len(sortedNamespaces)
	totalManual := 0
	totalArgoCD := 0
	totalResources := 0
	totalRootResources := 0
	totalExcluded := 0
	completelyManagedNamespaces := 0
	partiallyManagedNamespaces := 0
	unmanagedNamespaces := 0

	for _, result := range allResults {
		totalManual += result.ManualResources
		totalArgoCD += result.ArgoCDManaged
		totalResources += result.TotalResources
		totalRootResources += result.RootResources
		totalExcluded += result.ExcludedDefaults

		if result.RootResources == 0 {
			unmanagedNamespaces++
		} else if result.ManualResources == 0 {
			completelyManagedNamespaces++
		} else if result.ArgoCDManaged > 0 {
			partiallyManagedNamespaces++
		} else {
			unmanagedNamespaces++
		}
	}

	sb.WriteString("## 전체 통계\n\n")
	sb.WriteString("### 기본 정보\n")
	sb.WriteString(fmt.Sprintf("- **검사한 네임스페이스**: %d개\n", totalNamespaces))
	sb.WriteString(fmt.Sprintf("- **전체 리소스**: %d개\n", totalResources))
	sb.WriteString(fmt.Sprintf("- **최상위 리소스**: %d개\n", totalRootResources))
	sb.WriteString(fmt.Sprintf("- **ArgoCD 관리 리소스**: %d개\n", totalArgoCD))
	sb.WriteString(fmt.Sprintf("- **수동 생성 리소스**: %d개\n", totalManual))
	sb.WriteString(fmt.Sprintf("- **제외된 기본 리소스**: %d개\n\n", totalExcluded))

	sb.WriteString("### 네임스페이스 관리 상태\n")
	sb.WriteString(fmt.Sprintf("- **✅ 완전 관리**: %d개 네임스페이스\n", completelyManagedNamespaces))
	sb.WriteString(fmt.Sprintf("- **⚠️ 부분 관리**: %d개 네임스페이스\n", partiallyManagedNamespaces))
	sb.WriteString(fmt.Sprintf("- **❌ 미관리**: %d개 네임스페이스\n\n", unmanagedNamespaces))
}

func (r *MarkdownReporter) writeManualResourceDetails(sb *strings.Builder, allResults map[string]domain.AnalysisResult, sortedNamespaces []string) {
	sb.WriteString("## 수동 생성된 리소스 상세\n\n")

	hasManualResources := false
	for _, namespace := range sortedNamespaces {
		result := allResults[namespace]
		if len(result.ManualResourceList) > 0 {
			hasManualResources = true
			sb.WriteString(fmt.Sprintf("### %s\n\n", namespace))
			sb.WriteString("| API Version | Kind | Name | Created |\n")
			sb.WriteString("| --- | --- | --- | --- |\n")

			resources := result.ManualResourceList
			sort.Slice(resources, func(i, j int) bool {
				if resources[i].Identifier.APIVersion != resources[j].Identifier.APIVersion {
					return resources[i].Identifier.APIVersion < resources[j].Identifier.APIVersion
				}
				if resources[i].Identifier.Kind != resources[j].Identifier.Kind {
					return resources[i].Identifier.Kind < resources[j].Identifier.Kind
				}
				return resources[i].Identifier.Name < resources[j].Identifier.Name
			})

			for _, resource := range resources {
				created := resource.CreatedAt
				if len(created) > 19 {
					created = created[:19]
				}
				sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
					resource.Identifier.APIVersion,
					resource.Identifier.Kind,
					resource.Identifier.Name,
					created,
				))
			}
			sb.WriteString("\n")
		}
	}

	if !hasManualResources {
		sb.WriteString("✅ 모든 리소스가 ArgoCD로 관리되고 있습니다!\n\n")
	}
}
