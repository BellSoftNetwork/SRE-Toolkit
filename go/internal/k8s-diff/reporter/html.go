package reporter

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/k8s-diff/domain"
	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/k8s-diff/utils"
)

// HTMLReporter HTML 리포터
type HTMLReporter struct {
	outputDir string
}

// NewHTMLReporter 새 HTML 리포터 생성
func NewHTMLReporter(outputDir string) *HTMLReporter {
	return &HTMLReporter{
		outputDir: outputDir,
	}
}

// Generate HTML 리포트 생성
func (r *HTMLReporter) Generate(results map[string]domain.ComparisonResult, sourceCluster, targetCluster domain.ClusterInfo) error {
	// 출력 디렉토리 생성
	if err := os.MkdirAll(r.outputDir, 0755); err != nil {
		return fmt.Errorf("출력 디렉토리 생성 실패: %w", err)
	}

	// 리포트 데이터 준비
	data := r.prepareReportData(results, sourceCluster, targetCluster)

	// HTML 파일 생성
	timestamp := time.Now().Format("20060102_150405")
	filename := filepath.Join(r.outputDir, fmt.Sprintf("%s.html", timestamp))

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("HTML 파일 생성 실패: %w", err)
	}
	defer file.Close()

	// 템플릿 실행
	tmpl := template.Must(template.New("report").Parse(htmlTemplate))
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("템플릿 실행 실패: %w", err)
	}

	fmt.Printf("\n✅ HTML 리포트 생성됨: %s\n", filename)
	return nil
}

// prepareReportData 리포트 데이터 준비
func (r *HTMLReporter) prepareReportData(results map[string]domain.ComparisonResult, sourceCluster, targetCluster domain.ClusterInfo) map[string]interface{} {
	totalSourceOnly, totalTargetOnly, totalModified := 0, 0, 0
	var namespaceDetails []map[string]interface{}

	// 네임스페이스 정렬
	namespaces := make([]string, 0, len(results))
	for ns := range results {
		namespaces = append(namespaces, ns)
	}
	sort.Strings(namespaces)

	// 데이터 수집
	for _, ns := range namespaces {
		result := results[ns]
		sourceOnly := len(result.OnlyInSource)
		targetOnly := len(result.OnlyInTarget)
		modified := len(result.ModifiedResources)

		totalSourceOnly += sourceOnly
		totalTargetOnly += targetOnly
		totalModified += modified

		if sourceOnly > 0 || targetOnly > 0 || modified > 0 {
			// 네임스페이스별 리소스 목록 생성
			var resources []map[string]interface{}

			// 소스에만 있는 리소스
			for _, res := range result.OnlyInSource {
				resources = append(resources, map[string]interface{}{
					"Type":         res.Kind,
					"Name":         res.Name,
					"APIVersion":   res.APIVersion,
					"Status":       "source-only",
					"StatusIcon":   "🔴",
					"StatusText":   "소스만",
					"CreationTime": formatTime(res.CreationTime),
				})
			}

			// 타겟에만 있는 리소스
			for _, res := range result.OnlyInTarget {
				resources = append(resources, map[string]interface{}{
					"Type":         res.Kind,
					"Name":         res.Name,
					"APIVersion":   res.APIVersion,
					"Status":       "target-only",
					"StatusIcon":   "🟢",
					"StatusText":   "타겟만",
					"CreationTime": formatTime(res.CreationTime),
				})
			}

			// 리소스 정렬
			sort.Slice(resources, func(i, j int) bool {
				if resources[i]["Type"].(string) != resources[j]["Type"].(string) {
					return resources[i]["Type"].(string) < resources[j]["Type"].(string)
				}
				return resources[i]["Name"].(string) < resources[j]["Name"].(string)
			})

			namespaceDetails = append(namespaceDetails, map[string]interface{}{
				"Name":       ns,
				"Resources":  resources,
				"SourceOnly": sourceOnly,
				"TargetOnly": targetOnly,
				"Modified":   modified,
			})
		}
	}

	return map[string]interface{}{
		"GeneratedAt":      time.Now().Format("2006-01-02 15:04:05"),
		"SourceCluster":    utils.ExtractClusterName(sourceCluster.Name),
		"TargetCluster":    utils.ExtractClusterName(targetCluster.Name),
		"TotalSourceOnly":  totalSourceOnly,
		"TotalTargetOnly":  totalTargetOnly,
		"TotalModified":    totalModified,
		"NamespaceDetails": namespaceDetails,
	}
}

// formatTime 시간 포맷팅
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

// HTML 템플릿
const htmlTemplate = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>K8s 클러스터 비교 리포트</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
            color: #333;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background-color: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1 {
            color: #2c3e50;
            border-bottom: 3px solid #3498db;
            padding-bottom: 10px;
        }
        h2 {
            color: #34495e;
            margin-top: 30px;
        }
        .info {
            background-color: #ecf0f1;
            padding: 15px;
            border-radius: 5px;
            margin-bottom: 20px;
        }
        .summary {
            display: flex;
            gap: 20px;
            margin: 20px 0;
        }
        .summary-card {
            flex: 1;
            background-color: #f8f9fa;
            padding: 20px;
            border-radius: 5px;
            text-align: center;
            border: 2px solid transparent;
        }
        .summary-card.source-only {
            border-color: #f39c12;
        }
        .summary-card.target-only {
            border-color: #27ae60;
        }
        .summary-card.modified {
            border-color: #3498db;
        }
        .summary-card h3 {
            margin: 0 0 10px 0;
            color: #555;
        }
        .summary-card .count {
            font-size: 36px;
            font-weight: bold;
        }
        .source-only .count {
            color: #f39c12;
        }
        .target-only .count {
            color: #27ae60;
        }
        .modified .count {
            color: #3498db;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 20px;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        th {
            background-color: #f8f9fa;
            font-weight: 600;
            color: #2c3e50;
        }
        tr:hover {
            background-color: #f8f9fa;
        }
        .tag {
            display: inline-block;
            padding: 2px 8px;
            border-radius: 3px;
            font-size: 12px;
            font-weight: 500;
        }
        .tag.namespace {
            background-color: #e3f2fd;
            color: #1976d2;
        }
        .tag.kind {
            background-color: #f3e5f5;
            color: #7b1fa2;
        }
        .tag.api-version {
            background-color: #e8f5e9;
            color: #388e3c;
        }
        .details {
            margin-top: 30px;
            max-height: 500px;
            overflow-y: auto;
        }
        .details table {
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔍 Kubernetes 클러스터 비교 리포트</h1>
        
        <div class="info">
            <p><strong>생성 시간:</strong> {{.GeneratedAt}}</p>
            <p><strong>소스 클러스터:</strong> {{.SourceCluster}}</p>
            <p><strong>타겟 클러스터:</strong> {{.TargetCluster}}</p>
        </div>

        <h2>전체 요약</h2>
        <div class="summary">
            <div class="summary-card source-only">
                <h3>소스에만 있는 리소스</h3>
                <div class="count">{{.TotalSourceOnly}}</div>
            </div>
            <div class="summary-card target-only">
                <h3>타겟에만 있는 리소스</h3>
                <div class="count">{{.TotalTargetOnly}}</div>
            </div>
            {{if .TotalModified}}
            <div class="summary-card modified">
                <h3>수정된 리소스</h3>
                <div class="count">{{.TotalModified}}</div>
            </div>
            {{end}}
        </div>

        <h2>네임스페이스별 리소스 비교</h2>
        
        {{range .NamespaceDetails}}
        <div style="margin-bottom: 40px;">
            <h3 style="background-color: #f8f9fa; padding: 10px; border-radius: 5px;">
                네임스페이스: {{.Name}} 
                <span style="font-size: 14px; color: #666; margin-left: 20px;">
                    (소스만: {{.SourceOnly}}, 타겟만: {{.TargetOnly}}{{if .Modified}}, 수정됨: {{.Modified}}{{end}})
                </span>
            </h3>
            
            <table>
                <thead>
                    <tr>
                        <th>타입</th>
                        <th>이름</th>
                        <th>API 버전</th>
                        <th style="width: 120px;">상태</th>
                        <th>생성 시간</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Resources}}
                    <tr>
                        <td><span class="tag kind">{{.Type}}</span></td>
                        <td>{{.Name}}</td>
                        <td><span class="tag api-version">{{.APIVersion}}</span></td>
                        <td>
                            <span style="font-size: 18px;">{{.StatusIcon}}</span>
                            <span style="margin-left: 5px;">{{.StatusText}}</span>
                        </td>
                        <td>{{.CreationTime}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>
        {{end}}
    </div>
</body>
</html>`
