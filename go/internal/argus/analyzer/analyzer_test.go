package analyzer

import (
	"regexp"
	"testing"

	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/argus/config"
	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/argus/domain"
)

func TestNewAnalyzer(t *testing.T) {
	cfg := &config.Config{}
	analyzer := NewAnalyzer(cfg)

	if analyzer == nil {
		t.Error("NewAnalyzer()가 nil을 반환했습니다")
	}

	if analyzer.config != cfg {
		t.Error("Analyzer의 config이 올바르게 설정되지 않았습니다")
	}
}

func TestAnalyzeResources(t *testing.T) {
	tests := []struct {
		name      string
		resources []domain.KubernetesResource
		config    *config.Config
		want      domain.AnalysisResult
	}{
		{
			name:      "빈 리소스 목록",
			resources: []domain.KubernetesResource{},
			config:    &config.Config{},
			want: domain.AnalysisResult{
				TotalResources:     0,
				ManualResourceList: []domain.KubernetesResource{},
				ArgoCDResourceList: []domain.KubernetesResource{},
			},
		},
		{
			name: "루트 리소스가 아닌 경우 제외",
			resources: []domain.KubernetesResource{
				{
					Identifier: domain.ResourceIdentifier{
						Kind: "Pod",
						Name: "test-pod",
					},
					OwnerReferences: []interface{}{"some-owner"},
				},
			},
			config: &config.Config{},
			want: domain.AnalysisResult{
				TotalResources:     1,
				RootResources:      0,
				ManualResourceList: []domain.KubernetesResource{},
				ArgoCDResourceList: []domain.KubernetesResource{},
			},
		},
		{
			name: "ArgoCD 관리 리소스",
			resources: []domain.KubernetesResource{
				{
					Identifier: domain.ResourceIdentifier{
						Kind: "Deployment",
						Name: "test-deployment",
					},
					Labels: map[string]string{
						"argocd.argoproj.io/instance": "test-app",
					},
					OwnerReferences: []interface{}{},
				},
			},
			config: &config.Config{},
			want: domain.AnalysisResult{
				TotalResources:     1,
				RootResources:      1,
				ArgoCDManaged:      1,
				ManualResourceList: []domain.KubernetesResource{},
				ArgoCDResourceList: []domain.KubernetesResource{
					{
						Identifier: domain.ResourceIdentifier{
							Kind: "Deployment",
							Name: "test-deployment",
						},
						Labels: map[string]string{
							"argocd.argoproj.io/instance": "test-app",
						},
						OwnerReferences: []interface{}{},
					},
				},
			},
		},
		{
			name: "수동 관리 리소스",
			resources: []domain.KubernetesResource{
				{
					Identifier: domain.ResourceIdentifier{
						Kind: "ConfigMap",
						Name: "test-config",
					},
					Labels:          map[string]string{},
					OwnerReferences: []interface{}{},
				},
			},
			config: &config.Config{},
			want: domain.AnalysisResult{
				TotalResources:  1,
				RootResources:   1,
				ManualResources: 1,
				ManualResourceList: []domain.KubernetesResource{
					{
						Identifier: domain.ResourceIdentifier{
							Kind: "ConfigMap",
							Name: "test-config",
						},
						Labels:          map[string]string{},
						OwnerReferences: []interface{}{},
					},
				},
				ArgoCDResourceList: []domain.KubernetesResource{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewAnalyzer(tt.config)
			got := analyzer.AnalyzeResources(tt.resources)

			if got.TotalResources != tt.want.TotalResources {
				t.Errorf("TotalResources = %v, want %v", got.TotalResources, tt.want.TotalResources)
			}
			if got.RootResources != tt.want.RootResources {
				t.Errorf("RootResources = %v, want %v", got.RootResources, tt.want.RootResources)
			}
			if got.ArgoCDManaged != tt.want.ArgoCDManaged {
				t.Errorf("ArgoCDManaged = %v, want %v", got.ArgoCDManaged, tt.want.ArgoCDManaged)
			}
			if got.ManualResources != tt.want.ManualResources {
				t.Errorf("ManualResources = %v, want %v", got.ManualResources, tt.want.ManualResources)
			}
		})
	}
}

func TestShouldExcludeResource(t *testing.T) {
	tests := []struct {
		name     string
		resource *domain.KubernetesResource
		config   *config.Config
		want     bool
	}{
		{
			name: "제외 규칙과 일치하는 리소스",
			resource: &domain.KubernetesResource{
				Identifier: domain.ResourceIdentifier{
					Namespace: "kube-system",
					Kind:      "ConfigMap",
					Name:      "test-config",
				},
			},
			config: &config.Config{
				ExclusionRules: []config.ExclusionRule{
					{
						Namespace: "kube-system",
						Kind:      "ConfigMap",
						Name:      "*",
					},
				},
			},
			want: true,
		},
		{
			name: "제외된 시크릿 패턴",
			resource: &domain.KubernetesResource{
				Identifier: domain.ResourceIdentifier{
					Kind: "Secret",
					Name: "sh.helm.release.v1.test.v1",
				},
			},
			config: &config.Config{
				SecretPatterns: []*regexp.Regexp{
					regexp.MustCompile(`^sh\.helm\.release\.v1\.`),
				},
			},
			want: true,
		},
		{
			name: "cert-manager 어노테이션이 있는 시크릿",
			resource: &domain.KubernetesResource{
				Identifier: domain.ResourceIdentifier{
					Kind: "Secret",
					Name: "test-cert",
				},
				Annotations: map[string]string{
					"cert-manager.io/certificate-name": "test",
				},
			},
			config: &config.Config{
				CertManagerAnnotations: map[string]bool{
					"cert-manager.io/certificate-name": true,
				},
			},
			want: true,
		},
		{
			name: "Rancher 관리 리소스",
			resource: &domain.KubernetesResource{
				Identifier: domain.ResourceIdentifier{
					Kind: "ClusterRole",
					Name: "cattle-admin",
				},
			},
			config: &config.Config{
				RancherManagedPatterns: map[string][]*regexp.Regexp{
					"ClusterRole": {
						regexp.MustCompile(`^cattle-`),
					},
				},
			},
			want: true,
		},
		{
			name: "Rancher RoleBinding - rb- prefix",
			resource: &domain.KubernetesResource{
				Identifier: domain.ResourceIdentifier{
					Kind: "RoleBinding",
					Name: "rb-clusteradmin",
				},
			},
			config: &config.Config{
				RancherManagedPatterns: map[string][]*regexp.Regexp{
					"RoleBinding": {
						regexp.MustCompile(`^rb-`),
					},
				},
			},
			want: true,
		},
		{
			name: "자동 관리 어노테이션이 있는 리소스",
			resource: &domain.KubernetesResource{
				Annotations: map[string]string{
					"meta.helm.sh/release-name": "test",
				},
			},
			config: &config.Config{
				AutoManagedAnnotations: map[string]bool{
					"meta.helm.sh/release-name": true,
				},
			},
			want: true,
		},
		{
			name: "StatefulSet PVC",
			resource: &domain.KubernetesResource{
				Identifier: domain.ResourceIdentifier{
					Kind: "PersistentVolumeClaim",
					Name: "data-postgres-0",
				},
			},
			config: &config.Config{
				StatefulSetPVCPattern: regexp.MustCompile(`-\d+$`),
			},
			want: true,
		},
		{
			name: "제외되지 않는 일반 리소스",
			resource: &domain.KubernetesResource{
				Identifier: domain.ResourceIdentifier{
					Kind: "Deployment",
					Name: "test-app",
				},
			},
			config: &config.Config{},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := &Analyzer{config: tt.config}
			if got := analyzer.shouldExcludeResource(tt.resource); got != tt.want {
				t.Errorf("shouldExcludeResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsStatefulSetPVC(t *testing.T) {
	tests := []struct {
		name     string
		resource *domain.KubernetesResource
		config   *config.Config
		want     bool
	}{
		{
			name: "StatefulSet PVC - 패턴 일치",
			resource: &domain.KubernetesResource{
				Identifier: domain.ResourceIdentifier{
					Kind: "PersistentVolumeClaim",
					Name: "data-mongodb-0",
				},
			},
			config: &config.Config{
				StatefulSetPVCPattern: regexp.MustCompile(`-\d+$`),
			},
			want: true,
		},
		{
			name: "StatefulSet PVC - 라벨로 식별",
			resource: &domain.KubernetesResource{
				Identifier: domain.ResourceIdentifier{
					Kind: "PersistentVolumeClaim",
					Name: "test-pvc",
				},
				Labels: map[string]string{
					"app.kubernetes.io/instance":  "test",
					"app.kubernetes.io/component": "database",
				},
			},
			config: &config.Config{
				StatefulSetPVCPattern: regexp.MustCompile(`^$`), // 빈 패턴
			},
			want: true,
		},
		{
			name: "일반 PVC",
			resource: &domain.KubernetesResource{
				Identifier: domain.ResourceIdentifier{
					Kind: "PersistentVolumeClaim",
					Name: "test-pvc",
				},
			},
			config: &config.Config{
				StatefulSetPVCPattern: regexp.MustCompile(`-\d+$`),
			},
			want: false,
		},
		{
			name: "PVC가 아닌 리소스",
			resource: &domain.KubernetesResource{
				Identifier: domain.ResourceIdentifier{
					Kind: "ConfigMap",
					Name: "data-0",
				},
			},
			config: &config.Config{
				StatefulSetPVCPattern: regexp.MustCompile(`-\d+$`),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := &Analyzer{config: tt.config}
			if got := analyzer.isStatefulSetPVC(tt.resource); got != tt.want {
				t.Errorf("isStatefulSetPVC() = %v, want %v", got, tt.want)
			}
		})
	}
}
