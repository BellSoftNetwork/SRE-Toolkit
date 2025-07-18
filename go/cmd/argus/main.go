package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/argus/config"
	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/argus/reporter"
	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/argus/service"
	"gitlab.bellsoft.net/devops/sre-workbench/go/pkg/k8s/client"
	"gitlab.bellsoft.net/devops/sre-workbench/go/pkg/utils/color"
)

type CLIFlags struct {
	Namespace     *string
	Regex         *string
	Exclude       *string
	SkipConfirm   *bool
	Parallel      *int
	ConfigFile    *string
	BatchSize     *int
	FastScan      *bool
	GenerateImage *bool
	Timeout       *int
	Retry         *int
}

func main() {
	flags := parseCommandLineFlags()
	startTime := time.Now()

	cfg := loadConfiguration(flags)
	applyPerformanceSettings(cfg, flags)

	k8sClient := createKubernetesClient(cfg)
	svc := createAnalysisService(cfg, k8sClient, flags)

	context, cluster := svc.GetCurrentContext()
	displayApplicationHeader(context, cluster)
	displayAPISettings(*flags.Timeout, *flags.Retry)

	namespaces := resolveTargetNamespaces(svc, cfg, flags)
	validNamespaces := validateNamespaces(svc, namespaces)

	if shouldRequestConfirmation(flags) {
		confirmExecutionOrExit(validNamespaces)
	}

	executeResourceAnalysis(svc, validNamespaces, flags, context, cluster, startTime)
}

func parseCommandLineFlags() *CLIFlags {
	flags := &CLIFlags{
		Namespace:     flag.String("n", "", "네임스페이스 목록 (쉼표로 구분)"),
		Regex:         flag.String("r", "", "네임스페이스 필터링 정규식"),
		Exclude:       flag.String("exclude", "", "제외할 네임스페이스 패턴 (정규식)"),
		SkipConfirm:   flag.Bool("y", false, "확인 없이 바로 실행"),
		Parallel:      flag.Int("P", config.DefaultMaxConcurrent, "최대 동시 처리 수"),
		ConfigFile:    flag.String("f", "rules.yaml", "설정 파일 경로"),
		BatchSize:     flag.Int("batch-size", 0, "리소스 타입 배치 크기 (0=자동)"),
		FastScan:      flag.Bool("fast", false, "빠른 스캔 모드 (중요 리소스만 검사)"),
		GenerateImage: flag.Bool("image", false, "이미지 파일 생성"),
		Timeout:       flag.Int("timeout", 30, "API 요청 타임아웃 (초)"),
		Retry:         flag.Int("retry", 3, "타임아웃 시 재시도 횟수"),
	}
	flag.Parse()
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

func applyPerformanceSettings(cfg *config.Config, flags *CLIFlags) {
	if *flags.BatchSize > 0 {
		cfg.BatchSize = *flags.BatchSize
	}

	if *flags.FastScan {
		enableFastScanMode(cfg, flags)
	} else {
		cfg.ImportantResourceTypes = nil
	}
}

func enableFastScanMode(cfg *config.Config, flags *CLIFlags) {
	printInfo("⚡ 빠른 스캔 모드 활성화 (중요 리소스 %d개만 검사)", len(cfg.ImportantResourceTypes))
	if *flags.Parallel == config.DefaultMaxConcurrent {
		*flags.Parallel = cfg.Performance.FastScanConcurrent
	}
}

func createKubernetesClient(cfg *config.Config) *client.Client {
	printInfo("🚀 Kubernetes Go Client 사용")
	clientConfig := &client.ClientConfig{
		ImportantResourceTypes: cfg.ImportantResourceTypes,
		SkipResourceTypes:      cfg.SkipResourceTypes,
	}
	k8sClient, err := client.NewClient(clientConfig)
	if err != nil {
		exitWithError("Kubernetes 클라이언트 초기화 실패: %v", err)
	}
	return k8sClient
}

func createAnalysisService(cfg *config.Config, k8sClient *client.Client, flags *CLIFlags) *service.ScannerService {
	svc := service.NewScannerService(cfg, k8sClient)

	svc.AddReporter(reporter.NewConsoleReporter())
	svc.AddReporter(reporter.NewMarkdownReporter("reports"))
	svc.AddReporter(reporter.NewHTMLReporter("reports"))

	if *flags.GenerateImage {
		svc.AddReporter(reporter.NewImageReporter("reports"))
	}

	return svc
}

func displayApplicationHeader(context, cluster string) {
	fmt.Printf("%s🚀 Argus - Kubernetes 리소스 분석기%s\n", color.Bold, color.NC)
	fmt.Printf("현재 컨텍스트: %s%s%s\n", color.Cyan, context, color.NC)
	fmt.Printf("클러스터: %s%s%s\n", color.Cyan, cluster, color.NC)
}

func displayAPISettings(timeout, retry int) {
	fmt.Printf("%s⚙️  API 타임아웃: %d초, 재시도: %d회%s\n", color.Cyan, timeout, retry, color.NC)
}

func resolveTargetNamespaces(svc *service.ScannerService, cfg *config.Config, flags *CLIFlags) []string {
	namespaces, err := determineNamespaces(svc, cfg, flags)
	if err != nil {
		exitWithError("%v", err)
	}
	return namespaces
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
	if !confirmExecution(namespaces) {
		fmt.Println("\n실행 취소됨")
		os.Exit(0)
	}
}

func executeResourceAnalysis(svc *service.ScannerService, namespaces []string, flags *CLIFlags, context, cluster string, startTime time.Time) {
	maxConcurrent := limitConcurrency(*flags.Parallel)

	printInfo("⏳ 리소스 검사 시작... (동시 처리: %d)", maxConcurrent)

	allResults, err := svc.AnalyzeNamespaces(namespaces, maxConcurrent)
	if err != nil {
		exitWithError("%v", err)
	}

	if err := svc.GenerateReports(allResults, context, cluster, startTime); err != nil {
		exitWithError("보고서 생성 실패: %v", err)
	}
}

func limitConcurrency(requested int) int {
	const maxAllowed = 25
	if requested > maxAllowed {
		printWarning("동시 처리 수를 %d로 제한합니다 (안정성 보장)", maxAllowed)
		return maxAllowed
	}
	return requested
}

func determineNamespaces(svc *service.ScannerService, cfg *config.Config, flags *CLIFlags) ([]string, error) {
	args := flag.Args()

	if len(args) > 0 {
		return strings.Split(args[0], ","), nil
	}

	if *flags.Namespace != "" {
		return strings.Split(*flags.Namespace, ","), nil
	}

	return getAllNamespacesWithFilters(svc, flags)
}

func getAllNamespacesWithFilters(svc *service.ScannerService, flags *CLIFlags) ([]string, error) {
	printInfo("⏳ 모든 네임스페이스 조회 중...")

	allNamespaces, err := svc.GetAllNamespaces()
	if err != nil {
		return nil, fmt.Errorf("네임스페이스 조회 실패: %w", err)
	}

	namespaces := filterNamespacesByRegex(allNamespaces, flags)
	namespaces = excludeNamespacesByPattern(namespaces, flags)

	if len(namespaces) == 0 {
		return nil, fmt.Errorf("조건에 맞는 네임스페이스가 없습니다")
	}

	return namespaces, nil
}

func filterNamespacesByRegex(namespaces []string, flags *CLIFlags) []string {
	if *flags.Regex == "" {
		return namespaces
	}

	regex, err := regexp.Compile(*flags.Regex)
	if err != nil {
		printWarning("잘못된 정규식 패턴: %v", err)
		return namespaces
	}

	var filtered []string
	for _, ns := range namespaces {
		if regex.MatchString(ns) {
			filtered = append(filtered, ns)
		}
	}
	return filtered
}

func excludeNamespacesByPattern(namespaces []string, flags *CLIFlags) []string {
	if *flags.Exclude == "" {
		return namespaces
	}

	excludeRegex, err := regexp.Compile(*flags.Exclude)
	if err != nil {
		printWarning("잘못된 제외 패턴: %v", err)
		return namespaces
	}

	var filtered []string
	for _, ns := range namespaces {
		if !excludeRegex.MatchString(ns) {
			filtered = append(filtered, ns)
		}
	}
	return filtered
}

func confirmExecution(namespaces []string) bool {
	displayNamespaceSummary(namespaces)
	return promptUserConfirmation()
}

func displayNamespaceSummary(namespaces []string) {
	fmt.Printf("\n%s📋 검사할 네임스페이스 (%d개):%s\n", color.Cyan, len(namespaces), color.NC)

	sort.Strings(namespaces)
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
