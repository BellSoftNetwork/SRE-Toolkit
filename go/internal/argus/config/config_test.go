package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()
	if cfg == nil {
		t.Fatal("NewDefaultConfig()가 nil을 반환했습니다")
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		check   func(*testing.T, *Config)
	}{
		{
			name: "올바른 설정 파일",
			content: `
argocd:
  managed_labels:
    - "argocd.argoproj.io/instance"
  sync_annotations:
    - "argocd.argoproj.io/tracking-id"
exclusions:
  default_patterns:
    - "kube-system/ConfigMap/*"
auto_managed:
  annotations:
    - "meta.helm.sh/release-name"
  cert_manager_annotations:
    - "cert-manager.io/certificate-name"
patterns:
  secret_patterns:
    - '^sh\.helm\.release\.v1\.'
  rancher_managed:
    ClusterRole:
      - "^cattle-"
  statefulset_pvc: '-\d+$'
resource_types:
  skip:
    - "Event"
  important:
    - "Deployment"
performance:
  default_max_concurrent: 20
  batch_size: 100
`,
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				if len(cfg.ArgoCD.ManagedLabels) != 1 {
					t.Errorf("ManagedLabels 길이 = %v, want 1", len(cfg.ArgoCD.ManagedLabels))
				}
				if len(cfg.ExclusionRules) != 1 {
					t.Errorf("ExclusionRules 길이 = %v, want 1", len(cfg.ExclusionRules))
				}
				if len(cfg.SecretPatterns) != 1 {
					t.Errorf("SecretPatterns 길이 = %v, want 1", len(cfg.SecretPatterns))
				}
				if cfg.BatchSize != 100 {
					t.Errorf("BatchSize = %v, want 100", cfg.BatchSize)
				}
				if !cfg.SkipResourceTypes["Event"] {
					t.Error("Event가 SkipResourceTypes에 없습니다")
				}
			},
		},
		{
			name: "잘못된 정규식 패턴",
			content: `
patterns:
  secret_patterns:
    - "["
`,
			wantErr: true,
		},
		{
			name:    "빈 설정 파일",
			content: ``,
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				if cfg == nil {
					t.Fatal("Config가 nil입니다")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")

			if err := os.WriteFile(configFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("테스트 파일 생성 실패: %v", err)
			}

			cfg, err := LoadConfigFromFile(configFile)

			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfigFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

func TestLoadConfigFromFile_FileNotFound(t *testing.T) {
	_, err := LoadConfigFromFile("/non/existent/file.yaml")
	if err == nil {
		t.Error("존재하지 않는 파일에 대해 에러가 발생해야 합니다")
	}
}

func TestExclusionRuleParsing(t *testing.T) {
	content := `
exclusions:
  patterns:
    - "ns1/Kind1/name1"
    - "ns2/Kind2/*"
    - "*/Kind3/prefix*"
    - "invalid-pattern"
`
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	cfg, err := LoadConfigFromFile(configFile)
	if err != nil {
		t.Fatalf("LoadConfigFromFile() error = %v", err)
	}

	if len(cfg.ExclusionRules) != 3 {
		t.Errorf("ExclusionRules 길이 = %v, want 3", len(cfg.ExclusionRules))
	}

	expectedRules := []struct {
		namespace, kind, name string
	}{
		{"ns1", "Kind1", "name1"},
		{"ns2", "Kind2", "*"},
		{"*", "Kind3", "prefix*"},
	}

	for i, expected := range expectedRules {
		if i >= len(cfg.ExclusionRules) {
			break
		}
		rule := cfg.ExclusionRules[i]
		if rule.Namespace != expected.namespace {
			t.Errorf("ExclusionRules[%d].Namespace = %v, want %v", i, rule.Namespace, expected.namespace)
		}
		if rule.Kind != expected.kind {
			t.Errorf("ExclusionRules[%d].Kind = %v, want %v", i, rule.Kind, expected.kind)
		}
		if rule.Name != expected.name {
			t.Errorf("ExclusionRules[%d].Name = %v, want %v", i, rule.Name, expected.name)
		}
	}
}

func TestGetManagedLabels(t *testing.T) {
	cfg := &Config{
		ArgoCD: ArgoCDConfig{
			ManagedLabels: []string{"label1", "label2"},
		},
	}

	labels := cfg.GetManagedLabels()
	if len(labels) != 2 {
		t.Errorf("GetManagedLabels() 길이 = %v, want 2", len(labels))
	}
}

func TestGetSyncAnnotations(t *testing.T) {
	cfg := &Config{
		ArgoCD: ArgoCDConfig{
			SyncAnnotations: []string{"ann1", "ann2"},
		},
	}

	annotations := cfg.GetSyncAnnotations()
	if len(annotations) != 2 {
		t.Errorf("GetSyncAnnotations() 길이 = %v, want 2", len(annotations))
	}
}

func TestPatternCompilation(t *testing.T) {
	tests := []struct {
		name    string
		content string
		check   func(*testing.T, *Config)
	}{
		{
			name: "시크릿 패턴 컴파일",
			content: `
patterns:
  secret_patterns:
    - "^prefix-"
    - "-suffix$"
`,
			check: func(t *testing.T, cfg *Config) {
				if len(cfg.SecretPatterns) != 2 {
					t.Fatalf("SecretPatterns 길이 = %v, want 2", len(cfg.SecretPatterns))
				}
				if !cfg.SecretPatterns[0].MatchString("prefix-test") {
					t.Error("첫 번째 패턴이 'prefix-test'와 매치되어야 합니다")
				}
				if !cfg.SecretPatterns[1].MatchString("test-suffix") {
					t.Error("두 번째 패턴이 'test-suffix'와 매치되어야 합니다")
				}
			},
		},
		{
			name: "Rancher 관리 패턴 컴파일",
			content: `
patterns:
  rancher_managed:
    ClusterRole:
      - "^cattle-"
      - "^rancher-"
    ServiceAccount:
      - "^default$"
`,
			check: func(t *testing.T, cfg *Config) {
				if len(cfg.RancherManagedPatterns) != 2 {
					t.Fatalf("RancherManagedPatterns 길이 = %v, want 2", len(cfg.RancherManagedPatterns))
				}
				if len(cfg.RancherManagedPatterns["ClusterRole"]) != 2 {
					t.Errorf("ClusterRole 패턴 수 = %v, want 2", len(cfg.RancherManagedPatterns["ClusterRole"]))
				}
				if len(cfg.RancherManagedPatterns["ServiceAccount"]) != 1 {
					t.Errorf("ServiceAccount 패턴 수 = %v, want 1", len(cfg.RancherManagedPatterns["ServiceAccount"]))
				}
			},
		},
		{
			name: "StatefulSet PVC 패턴 컴파일",
			content: `
patterns:
  statefulset_pvc: '-\d+$'
`,
			check: func(t *testing.T, cfg *Config) {
				if cfg.StatefulSetPVCPattern == nil {
					t.Fatal("StatefulSetPVCPattern이 nil입니다")
				}
				if !cfg.StatefulSetPVCPattern.MatchString("data-postgres-0") {
					t.Error("패턴이 'data-postgres-0'와 매치되어야 합니다")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")

			if err := os.WriteFile(configFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("테스트 파일 생성 실패: %v", err)
			}

			cfg, err := LoadConfigFromFile(configFile)
			if err != nil {
				t.Fatalf("LoadConfigFromFile() error = %v", err)
			}

			tt.check(t, cfg)
		})
	}
}

func TestAnnotationMaps(t *testing.T) {
	content := `
auto_managed:
  annotations:
    - "helm.sh/release"
    - "meta.helm.sh/release-name"
  cert_manager_annotations:
    - "cert-manager.io/certificate-name"
    - "cert-manager.io/issuer"
`
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	cfg, err := LoadConfigFromFile(configFile)
	if err != nil {
		t.Fatalf("LoadConfigFromFile() error = %v", err)
	}

	if len(cfg.AutoManagedAnnotations) != 2 {
		t.Errorf("AutoManagedAnnotations 길이 = %v, want 2", len(cfg.AutoManagedAnnotations))
	}
	if !cfg.AutoManagedAnnotations["helm.sh/release"] {
		t.Error("helm.sh/release가 AutoManagedAnnotations에 없습니다")
	}

	if len(cfg.CertManagerAnnotations) != 2 {
		t.Errorf("CertManagerAnnotations 길이 = %v, want 2", len(cfg.CertManagerAnnotations))
	}
	if !cfg.CertManagerAnnotations["cert-manager.io/issuer"] {
		t.Error("cert-manager.io/issuer가 CertManagerAnnotations에 없습니다")
	}
}
