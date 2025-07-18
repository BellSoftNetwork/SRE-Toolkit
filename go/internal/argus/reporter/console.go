package reporter

import (
	"fmt"
	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/argus/domain"
	"gitlab.bellsoft.net/devops/sre-workbench/go/pkg/utils/color"
	"sort"
	"strings"
	"time"
)

type ConsoleReporter struct{}

func NewConsoleReporter() *ConsoleReporter {
	return &ConsoleReporter{}
}

func (r *ConsoleReporter) Generate(allResults map[string]domain.AnalysisResult, context, cluster string, startTime time.Time) error {
	var sortedNamespaces []string
	for ns := range allResults {
		sortedNamespaces = append(sortedNamespaces, ns)
	}
	sort.Strings(sortedNamespaces)

	r.printOverallStatistics(allResults, sortedNamespaces)

	r.printSummaryTable(allResults, sortedNamespaces)

	hasManualResources := false
	var manualNamespaces []string
	for _, namespace := range sortedNamespaces {
		result := allResults[namespace]
		if result.ManualResources > 0 {
			hasManualResources = true
			manualNamespaces = append(manualNamespaces, namespace)
		}
	}

	if hasManualResources {
		fmt.Printf("\n%s⚠️ 수동 생성된 리소스가 있는 네임스페이스:%s\n", color.Yellow, color.NC)
		for _, ns := range manualNamespaces {
			result := allResults[ns]
			fmt.Printf("  - %s: %d개\n", ns, result.ManualResources)
		}
		fmt.Printf("\n%s💡 상세 내용은 생성된 마크다운 보고서를 확인하세요%s\n", color.Cyan, color.NC)
	} else {
		fmt.Printf("\n%s✅ 모든 네임스페이스가 ArgoCD로 완전히 관리되고 있습니다!%s\n", color.Green, color.NC)
	}

	elapsed := time.Since(startTime)
	fmt.Printf("\n%s⏱️ 실행 시간: %.2f초%s\n", color.Cyan, elapsed.Seconds(), color.NC)

	return nil
}

func (r *ConsoleReporter) printOverallStatistics(allResults map[string]domain.AnalysisResult, sortedNamespaces []string) {
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

	fmt.Printf("\n%s📊 전체 통계%s\n", color.Bold, color.NC)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("검사한 네임스페이스: %d개\n", totalNamespaces)
	fmt.Printf("전체 리소스: %d개\n", totalResources)
	fmt.Printf("최상위 리소스: %d개\n", totalRootResources)
	fmt.Printf("ArgoCD 관리 리소스: %d개\n", totalArgoCD)
	fmt.Printf("수동 생성 리소스: %d개\n", totalManual)
	fmt.Printf("제외된 기본 리소스: %d개\n", totalExcluded)
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("%s완전 관리%s 네임스페이스: %d개\n", color.Green, color.NC, completelyManagedNamespaces)
	fmt.Printf("%s부분 관리%s 네임스페이스: %d개\n", color.Yellow, color.NC, partiallyManagedNamespaces)
	fmt.Printf("%s미관리%s 네임스페이스: %d개\n", color.Red, color.NC, unmanagedNamespaces)
	fmt.Println(strings.Repeat("=", 60))
}

func (r *ConsoleReporter) printSummaryTable(allResults map[string]domain.AnalysisResult, sortedNamespaces []string) {
	fmt.Printf("\n%s📊 최종 결과 요약%s\n", color.Bold, color.NC)
	fmt.Println(strings.Repeat("=", 120))

	fmt.Printf("%-30s %-12s %-12s %-15s %-15s %-12s %-15s\n",
		"네임스페이스", "ArgoCD 관리", "전체 리소스", "최상위 리소스", "ArgoCD 관리 중", "수동 생성", "기본 리소스")
	fmt.Println(strings.Repeat("-", 120))

	totalManual := 0
	totalArgoCD := 0
	totalResources := 0
	totalRootResources := 0
	totalExcluded := 0

	for _, namespace := range sortedNamespaces {
		result := allResults[namespace]

		var argoCDManaged string
		if result.RootResources == 0 {
			argoCDManaged = fmt.Sprintf("%s-%s", color.Dim, color.NC)
		} else if result.ManualResources == 0 {
			argoCDManaged = fmt.Sprintf("%s✅%s", color.Green, color.NC)
		} else if result.ArgoCDManaged > 0 {
			argoCDManaged = fmt.Sprintf("%s⚠️%s", color.Yellow, color.NC)
		} else {
			argoCDManaged = fmt.Sprintf("%s❌%s", color.Red, color.NC)
		}

		fmt.Printf("%-30s %-20s %-12d %-15d %-15d %-12d %-15d\n",
			namespace,
			argoCDManaged,
			result.TotalResources,
			result.RootResources,
			result.ArgoCDManaged,
			result.ManualResources,
			result.ExcludedDefaults,
		)

		totalManual += result.ManualResources
		totalArgoCD += result.ArgoCDManaged
		totalResources += result.TotalResources
		totalRootResources += result.RootResources
		totalExcluded += result.ExcludedDefaults
	}

	fmt.Println(strings.Repeat("-", 120))
	fmt.Printf("%s%-30s %-20s %-12d %-15d %-15d %-12d %-15d%s\n",
		color.Bold,
		"총계",
		"-",
		totalResources,
		totalRootResources,
		totalArgoCD,
		totalManual,
		totalExcluded,
		color.NC,
	)
}
