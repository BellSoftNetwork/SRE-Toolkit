package reporter

import (
	"bytes"
	"fmt"
	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/argus/domain"
	"html/template"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"
)

type HTMLReporter struct {
	outputDir string
}

func NewHTMLReporter(outputDir string) *HTMLReporter {
	return &HTMLReporter{
		outputDir: outputDir,
	}
}

func (r *HTMLReporter) Generate(results map[string]domain.AnalysisResult, context, cluster string, startTime time.Time) error {
	if err := os.MkdirAll(r.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	fontFamily := r.getSystemFontFamily()

	tmpl := template.Must(template.New("report").Funcs(template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"duration": func(start time.Time) string {
			return time.Since(start).Round(time.Second).String()
		},
		"fontFamily": func() string {
			return fontFamily
		},
		"getManagedStatus": func(result domain.AnalysisResult) string {
			if result.RootResources == 0 {
				return "➖"
			} else if result.ManualResources == 0 {
				return "✅"
			} else if result.ArgoCDManaged > 0 {
				return "⚠️"
			} else {
				return "❌"
			}
		},
	}).Parse(htmlTemplate))

	var sortedNamespaces []string
	for ns := range results {
		sortedNamespaces = append(sortedNamespaces, ns)
	}
	sort.Strings(sortedNamespaces)

	actionRequired := make(map[string]domain.AnalysisResult)
	for ns, result := range results {
		if result.ManualResources > 0 {
			actionRequired[ns] = result
		}
	}

	stats := r.calculateStatistics(results)

	data := struct {
		Context             string
		Cluster             string
		StartTime           time.Time
		TotalNamespaces     int
		ActionRequired      int
		AllResults          map[string]domain.AnalysisResult
		Results             map[string]domain.AnalysisResult
		SortedNamespaces    []string
		AllSortedNamespaces []string
		Stats               map[string]int
	}{
		Context:             context,
		Cluster:             cluster,
		StartTime:           startTime,
		TotalNamespaces:     len(results),
		ActionRequired:      len(actionRequired),
		AllResults:          results,
		Results:             actionRequired,
		SortedNamespaces:    []string{},
		AllSortedNamespaces: sortedNamespaces,
		Stats:               stats,
	}

	for ns := range actionRequired {
		data.SortedNamespaces = append(data.SortedNamespaces, ns)
	}
	sort.Strings(data.SortedNamespaces)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	filename := filepath.Join(r.outputDir, fmt.Sprintf("%s.html", startTime.Format("20060102_150405")))

	if err := os.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write HTML file: %w", err)
	}

	fmt.Printf("📄 HTML 리포트 생성: %s\n", filename)
	return nil
}

func (r *HTMLReporter) calculateStatistics(results map[string]domain.AnalysisResult) map[string]int {
	stats := map[string]int{
		"totalResources":              0,
		"totalRootResources":          0,
		"totalArgoCD":                 0,
		"totalManual":                 0,
		"totalExcluded":               0,
		"completelyManagedNamespaces": 0,
		"partiallyManagedNamespaces":  0,
		"unmanagedNamespaces":         0,
	}

	for _, result := range results {
		stats["totalResources"] += result.TotalResources
		stats["totalRootResources"] += result.RootResources
		stats["totalArgoCD"] += result.ArgoCDManaged
		stats["totalManual"] += result.ManualResources
		stats["totalExcluded"] += result.ExcludedDefaults

		if result.RootResources == 0 {
			stats["unmanagedNamespaces"]++
		} else if result.ManualResources == 0 {
			stats["completelyManagedNamespaces"]++
		} else if result.ArgoCDManaged > 0 {
			stats["partiallyManagedNamespaces"]++
		} else {
			stats["unmanagedNamespaces"]++
		}
	}

	return stats
}

func (r *HTMLReporter) getSystemFontFamily() string {
	switch runtime.GOOS {
	case "darwin":
		return "'Apple SD Gothic Neo', 'AppleGothic', sans-serif"
	case "windows":
		return "'Malgun Gothic', '맑은 고딕', 'Gulim', '굴림', sans-serif"
	default:
		return "'Noto Sans CJK KR', 'NanumGothic', '나눔고딕', 'UnDotum', '은 돋움', sans-serif"
	}
}

// HTML 템플릿
const htmlTemplate = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Argus 검사 결과</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: {{fontFamily}};
            background-color: #f5f5f5;
            color: #333;
            line-height: 1.6;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background-color: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #2c3e50;
            margin-bottom: 10px;
            font-size: 28px;
        }
        .header-info {
            color: #7f8c8d;
            margin-bottom: 30px;
            font-size: 14px;
        }
        .summary {
            background-color: #f8f9fa;
            padding: 20px;
            border-radius: 8px;
            margin-bottom: 30px;
            border-left: 4px solid #3498db;
        }
        .summary h2 {
            color: #2c3e50;
            font-size: 20px;
            margin-bottom: 15px;
        }
        .summary-stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            margin-top: 15px;
        }
        .stat-card {
            background-color: white;
            padding: 15px;
            border-radius: 6px;
            text-align: center;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }
        .stat-number {
            font-size: 32px;
            font-weight: bold;
            color: #3498db;
        }
        .stat-label {
            color: #7f8c8d;
            font-size: 14px;
            margin-top: 5px;
        }
        .namespace-section {
            margin-bottom: 30px;
        }
        .namespace-header {
            background-color: #e74c3c;
            color: white;
            padding: 15px 20px;
            border-radius: 8px 8px 0 0;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .namespace-name {
            font-size: 18px;
            font-weight: bold;
        }
        .resource-count {
            background-color: rgba(255,255,255,0.2);
            padding: 5px 15px;
            border-radius: 20px;
            font-size: 14px;
        }
        .resources-table {
            width: 100%;
            border-collapse: collapse;
            background-color: white;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }
        .resources-table th {
            background-color: #34495e;
            color: white;
            padding: 12px;
            text-align: left;
            font-weight: normal;
            font-size: 14px;
        }
        .resources-table td {
            padding: 12px;
            border-bottom: 1px solid #ecf0f1;
            font-size: 14px;
        }
        .resources-table tr:hover {
            background-color: #f8f9fa;
        }
        .resource-kind {
            font-weight: 500;
            color: #2c3e50;
        }
        .resource-name {
            color: #34495e;
            font-family: 'Monaco', 'Consolas', monospace;
            font-size: 13px;
        }
        .created-by {
            color: #7f8c8d;
            font-size: 13px;
        }
        .footer {
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #ecf0f1;
            text-align: center;
            color: #7f8c8d;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚨 Argus 검사 결과</h1>
        <div class="header-info">
            <div>컨텍스트: {{.Context}} | 클러스터: {{.Cluster}}</div>
            <div>검사 시작: {{formatTime .StartTime}} | 소요 시간: {{duration .StartTime}}</div>
        </div>

        <div class="summary">
            <h2>📊 전체 통계</h2>
            <div class="summary-stats">
                <div class="stat-card">
                    <div class="stat-number">{{.TotalNamespaces}}</div>
                    <div class="stat-label">검사한 네임스페이스</div>
                </div>
                <div class="stat-card">
                    <div class="stat-number">{{index .Stats "totalResources"}}</div>
                    <div class="stat-label">전체 리소스</div>
                </div>
                <div class="stat-card">
                    <div class="stat-number">{{index .Stats "totalRootResources"}}</div>
                    <div class="stat-label">최상위 리소스</div>
                </div>
                <div class="stat-card">
                    <div class="stat-number">{{index .Stats "totalArgoCD"}}</div>
                    <div class="stat-label">ArgoCD 관리</div>
                </div>
                <div class="stat-card">
                    <div class="stat-number">{{index .Stats "totalManual"}}</div>
                    <div class="stat-label">수동 생성</div>
                </div>
            </div>
            
            <h3 style="margin-top: 20px;">네임스페이스 관리 상태</h3>
            <div class="summary-stats">
                <div class="stat-card" style="background: linear-gradient(135deg, #27ae60, #2ecc71);">
                    <div class="stat-number">{{index .Stats "completelyManagedNamespaces"}}</div>
                    <div class="stat-label">✅ 완전 관리</div>
                </div>
                <div class="stat-card" style="background: linear-gradient(135deg, #f39c12, #f1c40f);">
                    <div class="stat-number">{{index .Stats "partiallyManagedNamespaces"}}</div>
                    <div class="stat-label">⚠️ 부분 관리</div>
                </div>
                <div class="stat-card" style="background: linear-gradient(135deg, #e74c3c, #c0392b);">
                    <div class="stat-number">{{index .Stats "unmanagedNamespaces"}}</div>
                    <div class="stat-label">❌ 미관리</div>
                </div>
            </div>
        </div>

        <h2 style="margin: 30px 0 20px;">📊 최종 결과 요약</h2>
        <table class="resources-table">
            <thead>
                <tr>
                    <th>네임스페이스</th>
                    <th style="text-align: center;">ArgoCD 관리</th>
                    <th style="text-align: right;">전체 리소스</th>
                    <th style="text-align: right;">최상위 리소스</th>
                    <th style="text-align: right;">ArgoCD 관리 중</th>
                    <th style="text-align: right;">수동 생성</th>
                    <th style="text-align: right;">기본 리소스</th>
                </tr>
            </thead>
            <tbody>
                {{range $ns := .AllSortedNamespaces}}
                {{$result := index $.AllResults $ns}}
                <tr>
                    <td>{{$ns}}</td>
                    <td style="text-align: center;">{{getManagedStatus $result}}</td>
                    <td style="text-align: right;">{{$result.TotalResources}}</td>
                    <td style="text-align: right;">{{$result.RootResources}}</td>
                    <td style="text-align: right;">{{$result.ArgoCDManaged}}</td>
                    <td style="text-align: right;">{{$result.ManualResources}}</td>
                    <td style="text-align: right;">{{$result.ExcludedDefaults}}</td>
                </tr>
                {{end}}
            </tbody>
            <tfoot>
                <tr style="font-weight: bold; background-color: #ecf0f1;">
                    <td>총계</td>
                    <td style="text-align: center;">-</td>
                    <td style="text-align: right;">{{index .Stats "totalResources"}}</td>
                    <td style="text-align: right;">{{index .Stats "totalRootResources"}}</td>
                    <td style="text-align: right;">{{index .Stats "totalArgoCD"}}</td>
                    <td style="text-align: right;">{{index .Stats "totalManual"}}</td>
                    <td style="text-align: right;">{{index .Stats "totalExcluded"}}</td>
                </tr>
            </tfoot>
        </table>

        {{if .ActionRequired}}
        <h2 style="margin-bottom: 20px; color: #e74c3c;">⚠️ 조치가 필요한 네임스페이스</h2>
        
        {{range $ns := .SortedNamespaces}}
        {{$result := index $.Results $ns}}
        <div class="namespace-section">
            <div class="namespace-header">
                <span class="namespace-name">{{$ns}}</span>
                <span class="resource-count">{{$result.ManualResources}}개 리소스</span>
            </div>
            <table class="resources-table">
                <thead>
                    <tr>
                        <th style="width: 20%;">리소스 타입</th>
                        <th style="width: 30%;">리소스 이름</th>
                        <th style="width: 20%;">API 버전</th>
                        <th style="width: 30%;">생성 정보</th>
                    </tr>
                </thead>
                <tbody>
                    {{range $resource := $result.ManualResourceList}}
                    <tr>
                        <td class="resource-kind">{{$resource.Identifier.Kind}}</td>
                        <td class="resource-name">{{$resource.Identifier.Name}}</td>
                        <td class="resource-kind">{{$resource.Identifier.APIVersion}}</td>
                        <td class="created-by">
                            수동으로 생성됨
                        </td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>
        {{end}}
        {{else}}
        <div style="text-align: center; padding: 40px; color: #27ae60;">
            <h2>✅ 모든 리소스가 ArgoCD로 관리되고 있습니다!</h2>
        </div>
        {{end}}

        <div class="footer">
            Generated by argus | {{formatTime .StartTime}}
        </div>
    </div>
</body>
</html>`
