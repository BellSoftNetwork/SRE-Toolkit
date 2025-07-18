package domain

import (
	"testing"

	"gitlab.bellsoft.net/devops/sre-workbench/go/internal/argus/config"
)

func TestKubernetesResource_IsRootResource(t *testing.T) {
	tests := []struct {
		name            string
		ownerReferences []interface{}
		want            bool
	}{
		{
			name:            "OwnerReference가 없는 경우",
			ownerReferences: []interface{}{},
			want:            true,
		},
		{
			name:            "nil OwnerReference",
			ownerReferences: nil,
			want:            true,
		},
		{
			name:            "OwnerReference가 있는 경우",
			ownerReferences: []interface{}{"some-owner"},
			want:            false,
		},
		{
			name:            "여러 OwnerReference가 있는 경우",
			ownerReferences: []interface{}{"owner1", "owner2"},
			want:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &KubernetesResource{
				OwnerReferences: tt.ownerReferences,
			}
			if got := r.IsRootResource(); got != tt.want {
				t.Errorf("IsRootResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKubernetesResource_IsArgoCDManaged(t *testing.T) {
	tests := []struct {
		name        string
		labels      map[string]string
		annotations map[string]string
		config      *config.Config
		want        bool
	}{
		{
			name: "ArgoCD 인스턴스 라벨이 있는 경우",
			labels: map[string]string{
				"argocd.argoproj.io/instance": "my-app",
			},
			want: true,
		},
		{
			name: "ArgoCD 트래킹 어노테이션이 있는 경우",
			annotations: map[string]string{
				"argocd.argoproj.io/tracking-id": "abc123",
			},
			want: true,
		},
		{
			name: "Config의 관리 라벨이 있는 경우",
			labels: map[string]string{
				"custom.io/managed": "true",
			},
			config: &config.Config{
				ArgoCD: config.ArgoCDConfig{
					ManagedLabels: []string{"custom.io/managed"},
				},
			},
			want: true,
		},
		{
			name: "Config의 동기화 어노테이션이 있는 경우",
			annotations: map[string]string{
				"custom.io/sync": "enabled",
			},
			config: &config.Config{
				ArgoCD: config.ArgoCDConfig{
					SyncAnnotations: []string{"custom.io/sync"},
				},
			},
			want: true,
		},
		{
			name: "라벨과 어노테이션 모두 있는 경우",
			labels: map[string]string{
				"argocd.argoproj.io/instance": "my-app",
			},
			annotations: map[string]string{
				"argocd.argoproj.io/tracking-id": "abc123",
			},
			want: true,
		},
		{
			name:        "관리되지 않는 리소스",
			labels:      map[string]string{"app": "test"},
			annotations: map[string]string{"version": "1.0"},
			config:      &config.Config{},
			want:        false,
		},
		{
			name: "nil maps",
			want: false,
		},
		{
			name:        "빈 라벨과 어노테이션",
			labels:      map[string]string{},
			annotations: map[string]string{},
			config:      &config.Config{},
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &KubernetesResource{
				Labels:      tt.labels,
				Annotations: tt.annotations,
				Config:      tt.config,
			}
			if got := r.IsArgoCDManaged(); got != tt.want {
				t.Errorf("IsArgoCDManaged() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceIdentifier(t *testing.T) {
	// ResourceIdentifier 구조체의 필드가 올바르게 설정되는지 테스트
	identifier := ResourceIdentifier{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
		Name:       "test-deployment",
		Namespace:  "default",
	}

	if identifier.APIVersion != "apps/v1" {
		t.Errorf("APIVersion = %v, want apps/v1", identifier.APIVersion)
	}
	if identifier.Kind != "Deployment" {
		t.Errorf("Kind = %v, want Deployment", identifier.Kind)
	}
	if identifier.Name != "test-deployment" {
		t.Errorf("Name = %v, want test-deployment", identifier.Name)
	}
	if identifier.Namespace != "default" {
		t.Errorf("Namespace = %v, want default", identifier.Namespace)
	}
}

func TestAnalysisResult(t *testing.T) {
	// AnalysisResult 초기화 테스트
	result := AnalysisResult{
		TotalResources:     100,
		RootResources:      80,
		ArgoCDManaged:      70,
		ManualResources:    10,
		ExcludedDefaults:   20,
		ManualResourceList: []KubernetesResource{},
		ArgoCDResourceList: []KubernetesResource{},
	}

	if result.TotalResources != 100 {
		t.Errorf("TotalResources = %v, want 100", result.TotalResources)
	}
	if result.RootResources != 80 {
		t.Errorf("RootResources = %v, want 80", result.RootResources)
	}
	if result.ArgoCDManaged != 70 {
		t.Errorf("ArgoCDManaged = %v, want 70", result.ArgoCDManaged)
	}
	if result.ManualResources != 10 {
		t.Errorf("ManualResources = %v, want 10", result.ManualResources)
	}
	if result.ExcludedDefaults != 20 {
		t.Errorf("ExcludedDefaults = %v, want 20", result.ExcludedDefaults)
	}
}

func TestNamespaceAnalysis(t *testing.T) {
	// NamespaceAnalysis 구조체 테스트
	analysis := NamespaceAnalysis{
		Namespace: "test-namespace",
		Result: AnalysisResult{
			TotalResources:  50,
			ArgoCDManaged:   40,
			ManualResources: 10,
		},
		Error: nil,
	}

	if analysis.Namespace != "test-namespace" {
		t.Errorf("Namespace = %v, want test-namespace", analysis.Namespace)
	}
	if analysis.Result.TotalResources != 50 {
		t.Errorf("Result.TotalResources = %v, want 50", analysis.Result.TotalResources)
	}
	if analysis.Error != nil {
		t.Errorf("Error = %v, want nil", analysis.Error)
	}
}
