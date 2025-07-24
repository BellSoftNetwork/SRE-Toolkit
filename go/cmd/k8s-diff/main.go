package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/k8s-diff/config"
	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/k8s-diff/domain"
	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/k8s-diff/k8sclient"
	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/k8s-diff/reporter"
	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/k8s-diff/service"
	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/k8s-diff/utils"
	"gitlab.bellsoft.net/devops/sre-workbench/go/pkg/utils/color"
)

type CLIFlags struct {
	SourceContext    *string
	TargetContext    *string
	Namespace        *string
	AllNamespaces    *bool
	SkipConfirm      *bool
	Parallel         *int
	ConfigFile       *string
	OutputFormat     *string
	CompareContents  *bool
	FastScan         *bool
	StrictAPIVersion *bool
}

func main() {
	flags := parseCommandLineFlags()
	startTime := time.Now()

	// 설정 로드
	cfg := loadConfiguration(flags)
	applyFlagsToConfig(cfg, flags)

	// 클러스터 클라이언트 생성
	sourceClient, targetClient, err := k8sclient.CreateClientPair(cfg, *flags.SourceContext, *flags.TargetContext)
	if err != nil {
		exitWithError("클라이언트 생성 실패: %v", err)
	}

	// 서비스 생성
	svc := createComparisonService(cfg, sourceClient, targetClient, flags)

	// 클러스터 정보 표시
	sourceInfo, targetInfo := svc.GetClusterInfo()
	displayApplicationHeader(sourceInfo, targetInfo)

	// 네임스페이스 결정
	namespaces := resolveTargetNamespaces(svc, flags)
	validNamespaces := validateNamespaces(svc, namespaces)

	// 실행 확인
	if shouldRequestConfirmation(flags) {
		confirmExecutionOrExit(validNamespaces)
	}

	// 비교 실행
	executeComparison(svc, validNamespaces, flags, sourceInfo, targetInfo, startTime)
}

func parseCommandLineFlags() *CLIFlags {
	flags := &CLIFlags{
		SourceContext:    flag.String("source", "", "소스 클러스터 컨텍스트 (필수)"),
		TargetContext:    flag.String("target", "", "타겟 클러스터 컨텍스트 (필수)"),
		Namespace:        flag.String("n", "", "네임스페이스 목록 (쉼표로 구분)"),
		AllNamespaces:    flag.Bool("A", false, "모든 네임스페이스 비교"),
		SkipConfirm:      flag.Bool("y", false, "확인 없이 바로 실행"),
		Parallel:         flag.Int("P", 20, "최대 동시 처리 수"),
		ConfigFile:       flag.String("f", "rules.yaml", "설정 파일 경로"),
		OutputFormat:     flag.String("o", "console,html,markdown", "출력 형식 (console,html,markdown)"),
		CompareContents:  flag.Bool("c", false, "리소스 내용 비교 (해시 기반)"),
		FastScan:         flag.Bool("fast", false, "빠른 스캔 모드 (중요 리소스만)"),
		StrictAPIVersion: flag.Bool("strict-api", false, "정밀 분석: API 버전이 다른 경우 다른 리소스로 처리"),
	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "사용법: %s [옵션]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "두 Kubernetes 클러스터 간의 리소스 차이를 비교합니다.\n\n")
		fmt.Fprintf(os.Stderr, "옵션:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n예제:\n")
		fmt.Fprintf(os.Stderr, "  # 두 클러스터 비교\n")
		fmt.Fprintf(os.Stderr, "  %s -source context1 -target context2\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # 특정 네임스페이스만 비교\n")
		fmt.Fprintf(os.Stderr, "  %s -source context1 -target context2 -n default,kube-system\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # 모든 네임스페이스 비교 (확인 없이)\n")
		fmt.Fprintf(os.Stderr, "  %s -source context1 -target context2 -A -y\n", os.Args[0])
	}

	flag.Parse()

	// 필수 플래그 검증
	if *flags.SourceContext == "" || *flags.TargetContext == "" {
		fmt.Fprintf(os.Stderr, "❌ 소스와 타겟 컨텍스트는 필수입니다.\n\n")
		flag.Usage()
		os.Exit(1)
	}

	return flags
}

func loadConfiguration(flags *CLIFlags) *config.Config {
	cfg, err := config.LoadConfigFromFile(*flags.ConfigFile)
	if err != nil {
		exitWithError("설정 파일 로드 실패 (%s): %v", *flags.ConfigFile, err)
	}
	printSuccess("설정 파일 로드됨: %s", *flags.ConfigFile)
	return cfg
}

func applyFlagsToConfig(cfg *config.Config, flags *CLIFlags) {
	if *flags.CompareContents {
		cfg.CompareResourceContents = true
	}

	if *flags.StrictAPIVersion {
		cfg.StrictAPIVersion = true
		printInfo("🔍 정밀 분석 모드: API 버전이 다른 경우 다른 리소스로 처리")
	}

	if *flags.FastScan && len(cfg.ImportantResourceTypes) > 0 {
		cfg.ResourceTypes = cfg.ImportantResourceTypes
		printInfo("⚡ 빠른 스캔 모드 활성화 (중요 리소스 %d개만 검사)", len(cfg.ResourceTypes))
	}

	if *flags.Parallel > 0 {
		cfg.MaxConcurrent = *flags.Parallel
	}
}

func createComparisonService(cfg *config.Config, sourceClient, targetClient *k8sclient.K8sClientWrapper, flags *CLIFlags) *service.ScannerService {
	svc := service.NewScannerService(cfg, sourceClient, targetClient)

	// 출력 형식에 따라 리포터 추가
	formats := strings.Split(*flags.OutputFormat, ",")
	for _, format := range formats {
		switch strings.TrimSpace(format) {
		case "console":
			svc.AddReporter(reporter.NewConsoleReporter())
		case "html":
			svc.AddReporter(reporter.NewHTMLReporter("reports"))
		case "markdown":
			svc.AddReporter(reporter.NewMarkdownReporter("reports"))
		}
	}

	return svc
}

func displayApplicationHeader(source, target domain.ClusterInfo) {
	fmt.Printf("%s🔍 K8s-Diff - Kubernetes 클러스터 비교 도구%s\n", color.Bold, color.NC)

	// 클러스터 이름에서 실제 이름 추출
	sourceName := utils.ExtractClusterName(source.Name)
	targetName := utils.ExtractClusterName(target.Name)

	fmt.Printf("소스 클러스터: %s%s%s (컨텍스트: %s)\n", color.Cyan, sourceName, color.NC, source.Context)
	fmt.Printf("타겟 클러스터: %s%s%s (컨텍스트: %s)\n", color.Cyan, targetName, color.NC, target.Context)
}

func resolveTargetNamespaces(svc *service.ScannerService, flags *CLIFlags) []string {
	if *flags.Namespace != "" {
		return strings.Split(*flags.Namespace, ",")
	}

	if *flags.AllNamespaces {
		namespaces, err := svc.GetAllNamespaces()
		if err != nil {
			exitWithError("네임스페이스 조회 실패: %v", err)
		}
		return namespaces
	}

	// 기본적으로 default 네임스페이스만 비교
	return []string{"default"}
}

func validateNamespaces(svc *service.ScannerService, namespaces []string) []string {
	validNamespaces, err := svc.ValidateNamespaces(namespaces)
	if err != nil {
		exitWithError("%v", err)
	}
	return validNamespaces
}

func shouldRequestConfirmation(flags *CLIFlags) bool {
	return !*flags.SkipConfirm
}

func confirmExecutionOrExit(namespaces []string) {
	displayNamespaceSummary(namespaces)
	if !promptUserConfirmation() {
		fmt.Println("\n실행 취소됨")
		os.Exit(0)
	}
}

func displayNamespaceSummary(namespaces []string) {
	fmt.Printf("\n%s📋 비교할 네임스페이스 (%d개):%s\n", color.Cyan, len(namespaces), color.NC)

	for i, ns := range namespaces {
		fmt.Printf("  %d. %s\n", i+1, ns)
		if i >= 9 && len(namespaces) > 10 {
			fmt.Printf("  ... 그리고 %d개 더\n", len(namespaces)-10)
			break
		}
	}
}

func promptUserConfirmation() bool {
	fmt.Printf("\n%s계속하시겠습니까? (y/N): %s", color.Yellow, color.NC)

	var response string
	fmt.Scanln(&response)

	return strings.ToLower(strings.TrimSpace(response)) == "y"
}

func executeComparison(svc *service.ScannerService, namespaces []string, flags *CLIFlags,
	sourceInfo, targetInfo domain.ClusterInfo, startTime time.Time) {

	maxConcurrent := limitConcurrency(*flags.Parallel)

	printInfo("⏳ 클러스터 비교 시작... (동시 처리: %d)", maxConcurrent)

	results, err := svc.CompareNamespaces(namespaces, maxConcurrent)
	if err != nil {
		exitWithError("%v", err)
	}

	if err := svc.GenerateReports(results, sourceInfo, targetInfo); err != nil {
		exitWithError("리포트 생성 실패: %v", err)
	}

	elapsed := time.Since(startTime)
	printSuccess("\n✨ 비교 완료! (소요 시간: %s)", elapsed.Round(time.Second))
}

func limitConcurrency(requested int) int {
	const maxAllowed = 30
	if requested > maxAllowed {
		printWarning("동시 처리 수를 %d로 제한합니다", maxAllowed)
		return maxAllowed
	}
	return requested
}

func printInfo(format string, args ...interface{}) {
	fmt.Printf("%s"+format+"%s\n", append([]interface{}{color.Cyan}, append(args, color.NC)...)...)
}

func printSuccess(format string, args ...interface{}) {
	fmt.Printf("%s✓ "+format+"%s\n", append([]interface{}{color.Green}, append(args, color.NC)...)...)
}

func printWarning(format string, args ...interface{}) {
	fmt.Printf("%s⚠️ "+format+"%s\n", append([]interface{}{color.Yellow}, append(args, color.NC)...)...)
}

func exitWithError(format string, args ...interface{}) {
	fmt.Printf("%s❌ "+format+"%s\n", append([]interface{}{color.Red}, append(args, color.NC)...)...)
	os.Exit(1)
}
