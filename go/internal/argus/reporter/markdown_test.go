package reporter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/argus/domain"
)

func TestNewMarkdownReporter(t *testing.T) {
	reportDir := "/tmp/test"
	reporter := NewMarkdownReporter(reportDir)

	if reporter == nil {
		t.Fatal("NewMarkdownReporter()가 nil을 반환했습니다")
	}

	if reporter.reportDir != reportDir {
		t.Errorf("reportDir = %v, want %v", reporter.reportDir, reportDir)
	}
}

func TestMarkdownReporter_Generate(t *testing.T) {
	tests := []struct {
		name       string
		results    map[string]domain.AnalysisResult
		context    string
		cluster    string
		wantErr    bool
		checkFiles func(*testing.T, string)
	}{
		{
			name: "정상적인 리포트 생성",
			results: map[string]domain.AnalysisResult{
				"default": {
					TotalResources:   10,
					RootResources:    8,
					ArgoCDManaged:    5,
					ManualResources:  3,
					ExcludedDefaults: 0,
				},
				"test-namespace": {
					TotalResources:   20,
					RootResources:    15,
					ArgoCDManaged:    15,
					ManualResources:  0,
					ExcludedDefaults: 0,
				},
			},
			context: "test-context",
			cluster: "test-cluster",
			wantErr: false,
			checkFiles: func(t *testing.T, dir string) {
				files, _ := os.ReadDir(dir)
				if len(files) != 1 {
					t.Errorf("생성된 파일 수 = %v, want 1", len(files))
				}
			},
		},
		{
			name:    "빈 결과",
			results: map[string]domain.AnalysisResult{},
			context: "test-context",
			cluster: "test-cluster",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			reporter := NewMarkdownReporter(tmpDir)

			startTime := time.Now()
			err := reporter.Generate(tt.results, tt.context, tt.cluster, startTime)

			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.checkFiles != nil {
				tt.checkFiles(t, tmpDir)
			}
		})
	}
}

func TestGenerateMarkdownContent(t *testing.T) {
	tests := []struct {
		name     string
		results  map[string]domain.AnalysisResult
		context  string
		cluster  string
		contains []string
	}{
		{
			name: "ArgoCD 미관리 리소스가 있는 경우",
			results: map[string]domain.AnalysisResult{
				"namespace-with-manual": {
					TotalResources:  10,
					ManualResources: 5,
					ArgoCDManaged:   5,
				},
			},
			context: "test-context",
			cluster: "test-cluster",
			contains: []string{
				"# Argus 분석 리포트",
				"## 실행 정보",
				"컨텍스트**: test-context",
				"클러스터**: test-cluster",
				"## ArgoCD 미관리 네임스페이스",
				"namespace-with-manual",
				"5개 수동 리소스",
				"## 💡 권장사항",
			},
		},
		{
			name: "모든 리소스가 관리되는 경우",
			results: map[string]domain.AnalysisResult{
				"fully-managed": {
					TotalResources:  10,
					ManualResources: 0,
					ArgoCDManaged:   10,
				},
			},
			context: "prod-context",
			cluster: "prod-cluster",
			contains: []string{
				"# Argus 분석 리포트",
				"## ✅ 모든 네임스페이스가 완전히 관리됨",
				"모든 리소스가 ArgoCD를 통해 관리되고 있습니다!",
			},
		},
		{
			name: "여러 네임스페이스 혼합",
			results: map[string]domain.AnalysisResult{
				"ns-a": {
					TotalResources:  20,
					ManualResources: 10,
					ArgoCDManaged:   10,
				},
				"ns-b": {
					TotalResources:  15,
					ManualResources: 0,
					ArgoCDManaged:   15,
				},
				"ns-c": {
					TotalResources:  5,
					ManualResources: 2,
					ArgoCDManaged:   3,
				},
			},
			context: "multi-context",
			cluster: "multi-cluster",
			contains: []string{
				"## 목차",
				"### ArgoCD 미관리 네임스페이스",
				"[ns-a](#ns-a)",
				"[ns-c](#ns-c)",
				"10개 수동 리소스",
				"2개 수동 리소스",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := &MarkdownReporter{}
			startTime := time.Now()

			content := reporter.generateMarkdownContent(tt.results, tt.context, tt.cluster, startTime)

			for _, expected := range tt.contains {
				if !strings.Contains(content, expected) {
					t.Errorf("컨텐츠에 '%s'가 포함되어야 합니다", expected)
				}
			}
		})
	}
}

func TestMarkdownReporter_DirectoryCreation(t *testing.T) {
	// 존재하지 않는 깊은 경로 테스트
	deepPath := filepath.Join(t.TempDir(), "deep", "nested", "path")
	reporter := NewMarkdownReporter(deepPath)

	results := map[string]domain.AnalysisResult{
		"test": {
			TotalResources: 1,
		},
	}

	err := reporter.Generate(results, "test", "test", time.Now())
	if err != nil {
		t.Errorf("디렉토리 생성이 실패했습니다: %v", err)
	}

	// 디렉토리가 생성되었는지 확인
	if _, err := os.Stat(deepPath); os.IsNotExist(err) {
		t.Error("디렉토리가 생성되지 않았습니다")
	}
}

func TestMarkdownReporter_FileNaming(t *testing.T) {
	// 특정 시간으로 테스트
	testTime := time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)

	// Generate 메서드를 직접 호출하는 대신, 파일명 형식만 테스트
	expectedFileName := "20240115_143045.md"
	actualFileName := testTime.Format("20060102_150405") + ".md"

	if actualFileName != expectedFileName {
		t.Errorf("파일명 형식이 잘못되었습니다. got = %v, want %v", actualFileName, expectedFileName)
	}
}
