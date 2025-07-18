package service

import (
	"errors"
	"testing"
	"time"

	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/argus/config"
	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/argus/domain"
	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/argus/reporter"
)

// Mock K8sClient
type mockK8sClient struct {
	currentContext    string
	currentCluster    string
	namespaces        []string
	resourceTypes     []string
	resources         []map[string]interface{}
	validationResults map[string]bool
	returnError       bool
	resourceTypeError bool
	validationError   bool
	getBatchError     bool
}

func (m *mockK8sClient) GetCurrentContext() (string, string) {
	return m.currentContext, m.currentCluster
}

func (m *mockK8sClient) GetAllNamespaces() ([]string, error) {
	if m.returnError {
		return nil, errors.New("mock error")
	}
	return m.namespaces, nil
}

func (m *mockK8sClient) ValidateNamespacesBatch(namespaces []string) (map[string]bool, error) {
	if m.validationError {
		return nil, errors.New("validation error")
	}
	if m.validationResults == nil {
		results := make(map[string]bool)
		for _, ns := range namespaces {
			results[ns] = true
		}
		return results, nil
	}
	return m.validationResults, nil
}

func (m *mockK8sClient) GetResourceTypes(includeNamespaced bool) ([]string, error) {
	if m.resourceTypeError {
		return nil, errors.New("resource type error")
	}
	return m.resourceTypes, nil
}

func (m *mockK8sClient) GetResourcesBatch(resourceTypes []string, namespace string) ([]map[string]interface{}, error) {
	if m.getBatchError {
		return nil, errors.New("get batch error")
	}
	return m.resources, nil
}

func (m *mockK8sClient) GetResources(resourceType, namespace string) ([]map[string]interface{}, error) {
	if m.returnError {
		return nil, errors.New("mock error")
	}
	return m.resources, nil
}

// Mock Reporter
type mockReporter struct {
	generateCalled bool
	returnError    bool
}

func (m *mockReporter) Generate(results map[string]domain.AnalysisResult, context, cluster string, startTime time.Time) error {
	m.generateCalled = true
	if m.returnError {
		return errors.New("reporter error")
	}
	return nil
}

func TestNewScannerService(t *testing.T) {
	cfg := &config.Config{}
	mockClient := &mockK8sClient{}

	scanner := NewScannerService(cfg, mockClient)

	if scanner == nil {
		t.Error("NewScannerService()가 nil을 반환했습니다")
	}
	// k8sClient가 nil이 아닌지만 확인 (인터페이스 타입이므로 직접 비교 불가)
	if scanner.k8sClient == nil {
		t.Error("k8sClient가 nil입니다")
	}
	if scanner.config != cfg {
		t.Error("config가 올바르게 설정되지 않았습니다")
	}
	if scanner.analyzer == nil {
		t.Error("analyzer가 초기화되지 않았습니다")
	}
}

func TestAddReporter(t *testing.T) {
	scanner := &ScannerService{
		reporters: []reporter.Reporter{},
	}

	mockReporter := &mockReporter{}
	scanner.AddReporter(mockReporter)

	if len(scanner.reporters) != 1 {
		t.Errorf("reporters 길이 = %v, want 1", len(scanner.reporters))
	}
}

func TestGetCurrentContext(t *testing.T) {
	mockClient := &mockK8sClient{
		currentContext: "test-context",
		currentCluster: "test-cluster",
	}
	scanner := &ScannerService{k8sClient: mockClient}

	context, cluster := scanner.GetCurrentContext()

	if context != "test-context" {
		t.Errorf("context = %v, want test-context", context)
	}
	if cluster != "test-cluster" {
		t.Errorf("cluster = %v, want test-cluster", cluster)
	}
}

func TestGetAllNamespaces(t *testing.T) {
	tests := []struct {
		name           string
		namespaces     []string
		exclusionRules []config.ExclusionRule
		returnError    bool
		want           []string
		wantErr        bool
	}{
		{
			name:       "제외 규칙 없음",
			namespaces: []string{"default", "test"},
			want:       []string{"default", "test"},
		},
		{
			name:       "네임스페이스 제외",
			namespaces: []string{"default", "kube-system", "test"},
			exclusionRules: []config.ExclusionRule{
				{Namespace: "kube-system", Kind: "*", Name: "*"},
			},
			want: []string{"default", "test"},
		},
		{
			name:        "에러 반환",
			returnError: true,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockK8sClient{
				namespaces:  tt.namespaces,
				returnError: tt.returnError,
			}
			scanner := &ScannerService{
				k8sClient: mockClient,
				config:    &config.Config{ExclusionRules: tt.exclusionRules},
			}

			got, err := scanner.GetAllNamespaces()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllNamespaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !equalSlices(got, tt.want) {
				t.Errorf("GetAllNamespaces() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateNamespaces(t *testing.T) {
	tests := []struct {
		name              string
		namespaces        []string
		validationResults map[string]bool
		validationError   bool
		want              []string
		wantErr           bool
	}{
		{
			name:       "모든 네임스페이스 유효",
			namespaces: []string{"default", "test"},
			want:       []string{"default", "test"},
		},
		{
			name:       "일부 네임스페이스 무효",
			namespaces: []string{"default", "invalid", "test"},
			validationResults: map[string]bool{
				"default": true,
				"invalid": false,
				"test":    true,
			},
			want: []string{"default", "test"},
		},
		{
			name:       "모든 네임스페이스 무효",
			namespaces: []string{"invalid1", "invalid2"},
			validationResults: map[string]bool{
				"invalid1": false,
				"invalid2": false,
			},
			wantErr: true,
		},
		{
			name:            "검증 에러",
			namespaces:      []string{"default"},
			validationError: true,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockK8sClient{
				validationResults: tt.validationResults,
				validationError:   tt.validationError,
			}
			scanner := &ScannerService{
				k8sClient: mockClient,
				config:    &config.Config{},
			}

			got, err := scanner.ValidateNamespaces(tt.namespaces)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNamespaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !equalSlices(got, tt.want) {
				t.Errorf("ValidateNamespaces() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptimizeConcurrency(t *testing.T) {
	tests := []struct {
		name            string
		namespaceCount  int
		requested       int
		wantConcurrency int
	}{
		{
			name:            "요청보다 적은 네임스페이스",
			namespaceCount:  5,
			requested:       10,
			wantConcurrency: 5,
		},
		{
			name:            "많은 네임스페이스",
			namespaceCount:  100,
			requested:       20,
			wantConcurrency: 10,
		},
		{
			name:            "절대 최대값 초과",
			namespaceCount:  50,
			requested:       20,
			wantConcurrency: 15,
		},
		{
			name:            "일반적인 경우",
			namespaceCount:  30,
			requested:       8,
			wantConcurrency: 8,
		},
	}

	scanner := &ScannerService{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			namespaces := make([]string, tt.namespaceCount)
			got := scanner.optimizeConcurrency(namespaces, tt.requested)

			if got != tt.wantConcurrency {
				t.Errorf("optimizeConcurrency() = %v, want %v", got, tt.wantConcurrency)
			}
		})
	}
}

func TestAnalyzeNamespaces(t *testing.T) {
	tests := []struct {
		name              string
		namespaces        []string
		resourceTypes     []string
		resources         []map[string]interface{}
		resourceTypeError bool
		wantErr           bool
	}{
		{
			name:          "정상 처리",
			namespaces:    []string{"default"},
			resourceTypes: []string{"Pod", "Service"},
			resources: []map[string]interface{}{
				{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "test-pod",
						"namespace": "default",
					},
				},
			},
			wantErr: false,
		},
		{
			name:              "리소스 타입 조회 실패",
			namespaces:        []string{"default"},
			resourceTypeError: true,
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockK8sClient{
				resourceTypes:     tt.resourceTypes,
				resources:         tt.resources,
				resourceTypeError: tt.resourceTypeError,
			}
			cfg := &config.Config{BatchSize: 5}
			scanner := NewScannerService(cfg, mockClient)

			results, err := scanner.AnalyzeNamespaces(tt.namespaces, 5)

			if (err != nil) != tt.wantErr {
				t.Errorf("AnalyzeNamespaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && results == nil {
				t.Error("AnalyzeNamespaces()가 nil 결과를 반환했습니다")
			}
		})
	}
}

func TestCalculateBatchSize(t *testing.T) {
	tests := []struct {
		name              string
		resourceTypeCount int
		configBatchSize   int
		want              int
	}{
		{
			name:              "설정된 배치 크기 사용",
			resourceTypeCount: 100,
			configBatchSize:   20,
			want:              20,
		},
		{
			name:              "최소 배치 크기",
			resourceTypeCount: 10,
			configBatchSize:   0,
			want:              5,
		},
		{
			name:              "최대 배치 크기",
			resourceTypeCount: 200,
			configBatchSize:   0,
			want:              15,
		},
		{
			name:              "일반적인 경우",
			resourceTypeCount: 50,
			configBatchSize:   0,
			want:              10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := &ScannerService{
				config: &config.Config{BatchSize: tt.configBatchSize},
			}

			got := scanner.calculateBatchSize(tt.resourceTypeCount)

			if got != tt.want {
				t.Errorf("calculateBatchSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateReports(t *testing.T) {
	tests := []struct {
		name          string
		reporterError bool
		wantErr       bool
	}{
		{
			name:    "정상 리포트 생성",
			wantErr: false,
		},
		{
			name:          "리포터 에러 처리",
			reporterError: true,
			wantErr:       false, // 경고만 출력하고 에러는 반환하지 않음
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReporter := &mockReporter{returnError: tt.reporterError}
			scanner := &ScannerService{
				reporters: []reporter.Reporter{mockReporter},
			}

			results := map[string]domain.AnalysisResult{
				"default": {},
			}

			err := scanner.GenerateReports(results, "test-context", "test-cluster", time.Now())

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateReports() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !mockReporter.generateCalled {
				t.Error("리포터의 Generate 메서드가 호출되지 않았습니다")
			}
		})
	}
}

func TestCreateResourceTypeBatches(t *testing.T) {
	tests := []struct {
		name          string
		resourceTypes []string
		batchSize     int
		wantBatches   int
	}{
		{
			name:          "균등 분할",
			resourceTypes: make([]string, 20),
			batchSize:     5,
			wantBatches:   4,
		},
		{
			name:          "나머지가 있는 분할",
			resourceTypes: make([]string, 23),
			batchSize:     5,
			wantBatches:   5,
		},
		{
			name:          "배치보다 작은 리소스",
			resourceTypes: make([]string, 3),
			batchSize:     5,
			wantBatches:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := &ScannerService{
				config: &config.Config{BatchSize: tt.batchSize},
			}

			batches := scanner.createResourceTypeBatches(tt.resourceTypes)

			if len(batches) != tt.wantBatches {
				t.Errorf("createResourceTypeBatches() 배치 수 = %v, want %v", len(batches), tt.wantBatches)
			}
		})
	}
}

// Helper function
func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
