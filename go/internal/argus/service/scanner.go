package service

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gitlab.bellsoft.net/devops/sre-toolkit/go/internal/argus/analyzer"
	"gitlab.bellsoft.net/devops/sre-toolkit/go/internal/argus/config"
	"gitlab.bellsoft.net/devops/sre-toolkit/go/internal/argus/domain"
	"gitlab.bellsoft.net/devops/sre-toolkit/go/internal/argus/reporter"
	"gitlab.bellsoft.net/devops/sre-toolkit/go/pkg/k8s/interface"
	"gitlab.bellsoft.net/devops/sre-toolkit/go/pkg/utils/color"
)

type ScannerService struct {
	k8sClient k8sinterface.K8sClient
	analyzer  *analyzer.Analyzer
	config    *config.Config
	reporters []reporter.Reporter
}

func NewScannerService(cfg *config.Config, k8sClient k8sinterface.K8sClient) *ScannerService {
	return &ScannerService{
		k8sClient: k8sClient,
		analyzer:  analyzer.NewAnalyzer(cfg),
		config:    cfg,
		reporters: []reporter.Reporter{},
	}
}

func (s *ScannerService) AddReporter(r reporter.Reporter) {
	s.reporters = append(s.reporters, r)
}

func (s *ScannerService) GetCurrentContext() (string, string) {
	return s.k8sClient.GetCurrentContext()
}

func (s *ScannerService) GetAllNamespaces() ([]string, error) {
	allNamespaces, err := s.k8sClient.GetAllNamespaces()
	if err != nil {
		return nil, err
	}

	return s.filterExcludedNamespaces(allNamespaces), nil
}

func (s *ScannerService) filterExcludedNamespaces(namespaces []string) []string {
	var filtered []string
	for _, ns := range namespaces {
		if !s.isNamespaceExcluded(ns) {
			filtered = append(filtered, ns)
		}
	}
	return filtered
}

func (s *ScannerService) isNamespaceExcluded(namespace string) bool {
	for _, rule := range s.config.ExclusionRules {
		if s.isWholeNamespaceRule(rule) && rule.Match(namespace, "*", "*") {
			return true
		}
	}
	return false
}

func (s *ScannerService) isWholeNamespaceRule(rule config.ExclusionRule) bool {
	return rule.Kind == "*" && rule.Name == "*"
}

func (s *ScannerService) ValidateNamespaces(namespaces []string) ([]string, error) {
	s.printValidationStart()

	validationResults, err := s.k8sClient.ValidateNamespacesBatch(namespaces)
	if err != nil {
		return nil, fmt.Errorf("네임스페이스 검증 실패: %w", err)
	}

	validNamespaces := s.extractValidNamespaces(namespaces, validationResults)
	if len(validNamespaces) == 0 {
		return nil, fmt.Errorf("유효한 네임스페이스가 없습니다")
	}

	s.printValidationComplete(len(validNamespaces))
	return validNamespaces, nil
}

func (s *ScannerService) printValidationStart() {
	fmt.Printf("\n%s⏳ 네임스페이스 검증 중...%s\n", color.Cyan, color.NC)
}

func (s *ScannerService) printValidationComplete(count int) {
	fmt.Printf("%s✓%s %d개 네임스페이스 검증 완료\n", color.Green, color.NC, count)
}

func (s *ScannerService) extractValidNamespaces(namespaces []string, validationResults map[string]bool) []string {
	var valid []string
	for _, ns := range namespaces {
		if validationResults[ns] {
			valid = append(valid, ns)
		} else {
			s.printInvalidNamespace(ns)
		}
	}
	return valid
}

func (s *ScannerService) printInvalidNamespace(namespace string) {
	fmt.Printf("%s❌ 네임스페이스 '%s'가 존재하지 않습니다%s\n", color.Red, namespace, color.NC)
}

func (s *ScannerService) AnalyzeNamespaces(namespaces []string, maxConcurrent int) (map[string]domain.AnalysisResult, error) {
	s.printResourceTypeQueryStart()
	resourceTypes, err := s.k8sClient.GetResourceTypes(true)
	if err != nil {
		return nil, fmt.Errorf("리소스 타입 조회 실패: %w", err)
	}
	s.printResourceTypeCount(len(resourceTypes))
	maxConcurrent = s.optimizeConcurrency(namespaces, maxConcurrent)
	s.printParallelProcessingStart(len(namespaces), maxConcurrent)

	workChan := s.createWorkChannel(namespaces)
	resultsChan := make(chan domain.NamespaceAnalysis, len(namespaces))

	var progress int32
	allResults := s.processNamespacesInParallel(workChan, resultsChan, resourceTypes, maxConcurrent, &progress, len(namespaces))

	fmt.Printf("\n\n%s✓%s 모든 네임스페이스 분석 완료\n", color.Green, color.NC)
	return allResults, nil
}

func (s *ScannerService) printResourceTypeQueryStart() {
	fmt.Printf("\n%s⏳ 리소스 타입 조회 중...%s\n", color.Cyan, color.NC)
}

func (s *ScannerService) printResourceTypeCount(count int) {
	fmt.Printf("%s✓%s %d개 리소스 타입 발견\n", color.Green, color.NC, count)
}

func (s *ScannerService) printParallelProcessingStart(namespaceCount, concurrency int) {
	fmt.Printf("\n%s⏳ %d개 네임스페이스 병렬 처리 시작 (최대 %d개 동시 처리)...%s\n",
		color.Cyan, namespaceCount, concurrency, color.NC)
}

func (s *ScannerService) optimizeConcurrency(namespaces []string, requested int) int {
	concurrent := requested

	if len(namespaces) < concurrent {
		concurrent = len(namespaces)
	}

	const largeNamespaceThreshold = 50
	const maxForLargeNamespaces = 10
	const absoluteMax = 15

	if len(namespaces) > largeNamespaceThreshold {
		concurrent = maxForLargeNamespaces
	}

	if concurrent > absoluteMax {
		concurrent = absoluteMax
	}

	return concurrent
}

func (s *ScannerService) createWorkChannel(namespaces []string) chan string {
	workChan := make(chan string, len(namespaces))
	for _, ns := range namespaces {
		workChan <- ns
	}
	close(workChan)
	return workChan
}

func (s *ScannerService) processNamespacesInParallel(
	workChan chan string,
	resultsChan chan domain.NamespaceAnalysis,
	resourceTypes []string,
	maxConcurrent int,
	progress *int32,
	totalNamespaces int,
) map[string]domain.AnalysisResult {
	var wg sync.WaitGroup
	wg.Add(maxConcurrent)

	for i := 0; i < maxConcurrent; i++ {
		go func() {
			defer wg.Done()
			for namespace := range workChan {
				s.processNamespace(resourceTypes, namespace, resultsChan, progress, totalNamespaces)
			}
		}()
	}

	allResults := make(map[string]domain.AnalysisResult)
	done := make(chan bool)

	go func() {
		for result := range resultsChan {
			if result.Error == nil {
				allResults[result.Namespace] = result.Result
			}
		}
		done <- true
	}()

	wg.Wait()
	close(resultsChan)
	<-done

	return allResults
}

func (s *ScannerService) processNamespace(
	resourceTypes []string,
	namespace string,
	results chan<- domain.NamespaceAnalysis,
	progress *int32,
	totalNamespaces int,
) {
	var allResources []domain.KubernetesResource
	var resourceMutex sync.Mutex

	batches := s.createResourceTypeBatches(resourceTypes)
	s.processBatchesInParallel(batches, namespace, &allResources, &resourceMutex)

	result := s.analyzer.AnalyzeResources(allResources)
	s.updateProgress(progress, totalNamespaces)

	results <- domain.NamespaceAnalysis{
		Namespace: namespace,
		Result:    result,
		Error:     nil,
	}
}

func (s *ScannerService) createResourceTypeBatches(resourceTypes []string) [][]string {
	batchSize := s.calculateBatchSize(len(resourceTypes))

	var batches [][]string
	for i := 0; i < len(resourceTypes); i += batchSize {
		end := min(i+batchSize, len(resourceTypes))
		batches = append(batches, resourceTypes[i:end])
	}
	return batches
}

func (s *ScannerService) calculateBatchSize(resourceTypeCount int) int {
	if s.config.BatchSize > 0 {
		return s.config.BatchSize
	}

	const (
		minBatchSize  = 5
		maxBatchSize  = 15
		targetBatches = 5
	)

	batchSize := resourceTypeCount / targetBatches
	if batchSize < minBatchSize {
		batchSize = minBatchSize
	}
	if batchSize > maxBatchSize {
		batchSize = maxBatchSize
	}

	return batchSize
}

func (s *ScannerService) processBatchesInParallel(
	batches [][]string,
	namespace string,
	allResources *[]domain.KubernetesResource,
	resourceMutex *sync.Mutex,
) {
	batchChan := make(chan []string, len(batches))
	for _, batch := range batches {
		batchChan <- batch
	}
	close(batchChan)

	const maxBatchConcurrency = 2
	concurrency := min(maxBatchConcurrency, len(batches))

	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			s.processBatchWorker(batchChan, namespace, allResources, resourceMutex)
		}()
	}

	wg.Wait()
}

func (s *ScannerService) processBatchWorker(
	batchChan chan []string,
	namespace string,
	allResources *[]domain.KubernetesResource,
	resourceMutex *sync.Mutex,
) {
	for batch := range batchChan {
		resources, err := s.k8sClient.GetResourcesBatch(batch, namespace)
		if err != nil {
			continue
		}

		localResources := s.convertToKubernetesResources(resources, namespace)
		if len(localResources) > 0 {
			resourceMutex.Lock()
			*allResources = append(*allResources, localResources...)
			resourceMutex.Unlock()
		}
	}
}

func (s *ScannerService) convertToKubernetesResources(resources []map[string]interface{}, namespace string) []domain.KubernetesResource {
	var kubeResources []domain.KubernetesResource
	for _, res := range resources {
		if kr := analyzer.MapToResource(res, namespace, s.config); kr != nil {
			kubeResources = append(kubeResources, *kr)
		}
	}
	return kubeResources
}

func (s *ScannerService) updateProgress(progress *int32, total int) {
	current := atomic.AddInt32(progress, 1)
	fmt.Printf("\r%s진행중: %d/%d 완료%s", color.Cyan, current, total, color.NC)
}

func (s *ScannerService) GenerateReports(allResults map[string]domain.AnalysisResult, context, cluster string, startTime time.Time) error {
	s.printReportHeader()

	for _, reporter := range s.reporters {
		if err := reporter.Generate(allResults, context, cluster, startTime); err != nil {
			s.printReportGenerationWarning(err)
		}
	}

	return nil
}

func (s *ScannerService) printReportHeader() {
	fmt.Printf("\n%s📊 네임스페이스별 분석 결과%s\n", color.Bold, color.NC)
	fmt.Println("=" + strings.Repeat("=", 99))
}

func (s *ScannerService) printReportGenerationWarning(err error) {
	fmt.Printf("%s⚠️ 리포트 생성 실패: %v%s\n", color.Yellow, err, color.NC)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
