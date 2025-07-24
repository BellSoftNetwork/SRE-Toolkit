package service

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/k8s-diff/analyzer"
	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/k8s-diff/config"
	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/k8s-diff/domain"
	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/k8s-diff/reporter"
	"gitlab.bellsoft.net/devops/sre-workbench/go/pkg/k8s/interface"
	"gitlab.bellsoft.net/devops/sre-workbench/go/pkg/utils/color"
)

// ScannerService 클러스터 스캔 서비스
type ScannerService struct {
	sourceClient k8sinterface.K8sClient
	targetClient k8sinterface.K8sClient
	analyzer     *analyzer.Analyzer
	config       *config.Config
	reporters    []reporter.Reporter
}

// NewScannerService 새 스캐너 서비스 생성
func NewScannerService(cfg *config.Config, sourceClient, targetClient k8sinterface.K8sClient) *ScannerService {
	return &ScannerService{
		sourceClient: sourceClient,
		targetClient: targetClient,
		analyzer:     analyzer.NewAnalyzer(cfg),
		config:       cfg,
		reporters:    []reporter.Reporter{},
	}
}

// AddReporter 리포터 추가
func (s *ScannerService) AddReporter(r reporter.Reporter) {
	s.reporters = append(s.reporters, r)
}

// GetClusterInfo 클러스터 정보 가져오기
func (s *ScannerService) GetClusterInfo() (source, target domain.ClusterInfo) {
	sourceContext, sourceCluster := s.sourceClient.GetCurrentContext()
	targetContext, targetCluster := s.targetClient.GetCurrentContext()

	return domain.ClusterInfo{Context: sourceContext, Name: sourceCluster},
		domain.ClusterInfo{Context: targetContext, Name: targetCluster}
}

// GetAllNamespaces 모든 네임스페이스 가져오기
func (s *ScannerService) GetAllNamespaces() ([]string, error) {
	sourceNamespaces, err := s.sourceClient.GetAllNamespaces()
	if err != nil {
		return nil, fmt.Errorf("소스 클러스터 네임스페이스 조회 실패: %w", err)
	}

	targetNamespaces, err := s.targetClient.GetAllNamespaces()
	if err != nil {
		return nil, fmt.Errorf("타겟 클러스터 네임스페이스 조회 실패: %w", err)
	}

	// 두 클러스터의 모든 네임스페이스 통합
	namespaceSet := make(map[string]bool)
	for _, ns := range sourceNamespaces {
		if !s.isNamespaceExcluded(ns) {
			namespaceSet[ns] = true
		}
	}
	for _, ns := range targetNamespaces {
		if !s.isNamespaceExcluded(ns) {
			namespaceSet[ns] = true
		}
	}

	namespaces := make([]string, 0, len(namespaceSet))
	for ns := range namespaceSet {
		namespaces = append(namespaces, ns)
	}
	sort.Strings(namespaces)

	return namespaces, nil
}

// isNamespaceExcluded 네임스페이스 제외 여부 확인
func (s *ScannerService) isNamespaceExcluded(namespace string) bool {
	for _, rule := range s.config.ExclusionRules {
		if rule.Namespace == namespace && rule.Kind == "*" && rule.Name == "*" {
			return true
		}
	}
	return false
}

// ValidateNamespaces 네임스페이스 유효성 검증
func (s *ScannerService) ValidateNamespaces(namespaces []string) ([]string, error) {
	fmt.Printf("\n%s⏳ 네임스페이스 검증 중...%s\n", color.Cyan, color.NC)

	sourceValidation, err := s.sourceClient.ValidateNamespacesBatch(namespaces)
	if err != nil {
		return nil, fmt.Errorf("소스 클러스터 네임스페이스 검증 실패: %w", err)
	}

	targetValidation, err := s.targetClient.ValidateNamespacesBatch(namespaces)
	if err != nil {
		return nil, fmt.Errorf("타겟 클러스터 네임스페이스 검증 실패: %w", err)
	}

	var validNamespaces []string
	for _, ns := range namespaces {
		sourceValid := sourceValidation[ns]
		targetValid := targetValidation[ns]

		if sourceValid || targetValid {
			validNamespaces = append(validNamespaces, ns)
			if !sourceValid {
				fmt.Printf("%s⚠️ 네임스페이스 '%s'는 타겟 클러스터에만 존재%s\n", color.Yellow, ns, color.NC)
			} else if !targetValid {
				fmt.Printf("%s⚠️ 네임스페이스 '%s'는 소스 클러스터에만 존재%s\n", color.Yellow, ns, color.NC)
			}
		} else {
			fmt.Printf("%s❌ 네임스페이스 '%s'가 두 클러스터 모두에 존재하지 않습니다%s\n", color.Red, ns, color.NC)
		}
	}

	fmt.Printf("%s✓%s %d개 네임스페이스 검증 완료\n", color.Green, color.NC, len(validNamespaces))
	return validNamespaces, nil
}

// CompareNamespaces 네임스페이스별 리소스 비교
func (s *ScannerService) CompareNamespaces(namespaces []string, maxConcurrent int) (map[string]domain.ComparisonResult, error) {
	fmt.Printf("\n%s⏳ 리소스 타입 조회 중...%s\n", color.Cyan, color.NC)

	resourceTypes, err := s.getResourceTypes()
	if err != nil {
		return nil, err
	}

	fmt.Printf("%s✓%s %d개 리소스 타입 발견\n", color.Green, color.NC, len(resourceTypes))

	if len(namespaces) < maxConcurrent {
		maxConcurrent = len(namespaces)
	}

	fmt.Printf("\n%s⏳ %d개 네임스페이스 비교 시작 (최대 %d개 동시 처리)...%s\n",
		color.Cyan, len(namespaces), maxConcurrent, color.NC)

	workChan := make(chan string, len(namespaces))
	for _, ns := range namespaces {
		workChan <- ns
	}
	close(workChan)

	resultsChan := make(chan domain.NamespaceComparison, len(namespaces))
	var wg sync.WaitGroup
	var progress int32

	wg.Add(maxConcurrent)
	for i := 0; i < maxConcurrent; i++ {
		go func() {
			defer wg.Done()
			for namespace := range workChan {
				s.compareNamespace(namespace, resourceTypes, resultsChan, &progress, len(namespaces))
			}
		}()
	}

	allResults := make(map[string]domain.ComparisonResult)
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

	fmt.Printf("\n\n%s✓%s 모든 네임스페이스 비교 완료\n", color.Green, color.NC)
	return allResults, nil
}

// getResourceTypes 리소스 타입 가져오기
func (s *ScannerService) getResourceTypes() ([]string, error) {
	// 설정에서 리소스 타입 사용
	if len(s.config.ResourceTypes) > 0 {
		return s.config.ResourceTypes, nil
	}

	// 소스 클러스터에서 리소스 타입 조회
	resourceTypes, err := s.sourceClient.GetResourceTypes(true)
	if err != nil {
		return nil, fmt.Errorf("리소스 타입 조회 실패: %w", err)
	}

	// 스킵할 리소스 타입 필터링
	var filtered []string
	skipMap := make(map[string]bool)
	for _, skip := range s.config.SkipResourceTypes {
		skipMap[skip] = true
	}

	for _, rt := range resourceTypes {
		if !skipMap[rt] {
			filtered = append(filtered, rt)
		}
	}

	return filtered, nil
}

// compareNamespace 단일 네임스페이스 비교
func (s *ScannerService) compareNamespace(namespace string, resourceTypes []string,
	results chan<- domain.NamespaceComparison, progress *int32, total int) {

	// 소스 클러스터에서 리소스 가져오기
	sourceResources, err := s.getNamespaceResources(s.sourceClient, namespace, resourceTypes)
	if err != nil {
		results <- domain.NamespaceComparison{
			Namespace: namespace,
			Error:     fmt.Errorf("소스 클러스터 리소스 조회 실패: %w", err),
		}
		return
	}

	// 타겟 클러스터에서 리소스 가져오기
	targetResources, err := s.getNamespaceResources(s.targetClient, namespace, resourceTypes)
	if err != nil {
		results <- domain.NamespaceComparison{
			Namespace: namespace,
			Error:     fmt.Errorf("타겟 클러스터 리소스 조회 실패: %w", err),
		}
		return
	}

	// 리소스 비교
	result := s.analyzer.CompareResources(sourceResources, targetResources)

	current := atomic.AddInt32(progress, 1)
	fmt.Printf("\r%s진행중: %d/%d 완료%s", color.Cyan, current, total, color.NC)

	results <- domain.NamespaceComparison{
		Namespace: namespace,
		Result:    result,
		Error:     nil,
	}
}

// getNamespaceResources 네임스페이스의 모든 리소스 가져오기
func (s *ScannerService) getNamespaceResources(client k8sinterface.K8sClient, namespace string, resourceTypes []string) ([]domain.KubernetesResource, error) {
	var allResources []domain.KubernetesResource
	var mu sync.Mutex

	// 배치 처리
	batches := s.createBatches(resourceTypes)
	var wg sync.WaitGroup

	for _, batch := range batches {
		wg.Add(1)
		go func(batch []string) {
			defer wg.Done()

			resources, err := client.GetResourcesBatch(batch, namespace)
			if err != nil {
				return
			}

			var localResources []domain.KubernetesResource
			for _, res := range resources {
				if kr := analyzer.MapToResource(res, namespace, s.config); kr != nil {
					if !s.analyzer.IsExcluded(*kr) {
						localResources = append(localResources, *kr)
					}
				}
			}

			if len(localResources) > 0 {
				mu.Lock()
				allResources = append(allResources, localResources...)
				mu.Unlock()
			}
		}(batch)
	}

	wg.Wait()
	return allResources, nil
}

// createBatches 리소스 타입을 배치로 분할
func (s *ScannerService) createBatches(resourceTypes []string) [][]string {
	batchSize := s.config.BatchSize
	if batchSize <= 0 {
		batchSize = 10
	}

	var batches [][]string
	for i := 0; i < len(resourceTypes); i += batchSize {
		end := i + batchSize
		if end > len(resourceTypes) {
			end = len(resourceTypes)
		}
		batches = append(batches, resourceTypes[i:end])
	}
	return batches
}

// GenerateReports 리포트 생성
func (s *ScannerService) GenerateReports(results map[string]domain.ComparisonResult, sourceCluster, targetCluster domain.ClusterInfo) error {
	fmt.Printf("\n%s📊 클러스터 비교 결과%s\n", color.Bold, color.NC)
	fmt.Println("=" + strings.Repeat("=", 99))

	for _, reporter := range s.reporters {
		if err := reporter.Generate(results, sourceCluster, targetCluster); err != nil {
			fmt.Printf("%s⚠️ 리포트 생성 실패: %v%s\n", color.Yellow, err, color.NC)
		}
	}

	return nil
}
