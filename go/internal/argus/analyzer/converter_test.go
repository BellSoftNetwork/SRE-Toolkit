package analyzer

import (
	"reflect"
	"testing"

	"gitlab.bellsoft.net/devops/sre-toolkit/go/internal/argus/config"
	"gitlab.bellsoft.net/devops/sre-toolkit/go/internal/argus/domain"
)

func TestMapToResource(t *testing.T) {
	tests := []struct {
		name          string
		obj           map[string]interface{}
		namespace     string
		expectedNil   bool
		expectedCheck func(*testing.T, *domain.KubernetesResource)
	}{
		{
			name: "올바른 리소스 변환",
			obj: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name":              "test-deployment",
					"namespace":         "test-ns",
					"creationTimestamp": "2024-01-01T00:00:00Z",
					"labels": map[string]interface{}{
						"app": "test",
					},
					"annotations": map[string]interface{}{
						"test.io/version": "v1",
					},
					"ownerReferences": []interface{}{
						map[string]interface{}{
							"kind": "ReplicaSet",
							"name": "test-rs",
						},
					},
				},
			},
			namespace:   "default",
			expectedNil: false,
			expectedCheck: func(t *testing.T, r *domain.KubernetesResource) {
				if r.Identifier.APIVersion != "apps/v1" {
					t.Errorf("APIVersion = %v, want apps/v1", r.Identifier.APIVersion)
				}
				if r.Identifier.Kind != "Deployment" {
					t.Errorf("Kind = %v, want Deployment", r.Identifier.Kind)
				}
				if r.Identifier.Name != "test-deployment" {
					t.Errorf("Name = %v, want test-deployment", r.Identifier.Name)
				}
				if r.Identifier.Namespace != "test-ns" {
					t.Errorf("Namespace = %v, want test-ns", r.Identifier.Namespace)
				}
				if r.CreatedAt != "2024-01-01T00:00:00Z" {
					t.Errorf("CreatedAt = %v, want 2024-01-01T00:00:00Z", r.CreatedAt)
				}
				if r.Labels["app"] != "test" {
					t.Errorf("Labels[app] = %v, want test", r.Labels["app"])
				}
				if r.Annotations["test.io/version"] != "v1" {
					t.Errorf("Annotations[test.io/version] = %v, want v1", r.Annotations["test.io/version"])
				}
				if len(r.OwnerReferences) != 1 {
					t.Errorf("len(OwnerReferences) = %v, want 1", len(r.OwnerReferences))
				}
			},
		},
		{
			name: "네임스페이스가 없는 리소스 - 기본값 사용",
			obj: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "test-config",
				},
			},
			namespace:   "default-ns",
			expectedNil: false,
			expectedCheck: func(t *testing.T, r *domain.KubernetesResource) {
				if r.Identifier.Namespace != "default-ns" {
					t.Errorf("Namespace = %v, want default-ns", r.Identifier.Namespace)
				}
			},
		},
		{
			name: "메타데이터가 없는 경우",
			obj: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
			},
			namespace:   "default",
			expectedNil: true,
		},
		{
			name: "메타데이터가 잘못된 타입인 경우",
			obj: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata":   "invalid",
			},
			namespace:   "default",
			expectedNil: true,
		},
		{
			name: "빈 라벨과 어노테이션",
			obj: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name":        "test-service",
					"labels":      map[string]interface{}{},
					"annotations": map[string]interface{}{},
				},
			},
			namespace:   "default",
			expectedNil: false,
			expectedCheck: func(t *testing.T, r *domain.KubernetesResource) {
				if len(r.Labels) != 0 {
					t.Errorf("len(Labels) = %v, want 0", len(r.Labels))
				}
				if len(r.Annotations) != 0 {
					t.Errorf("len(Annotations) = %v, want 0", len(r.Annotations))
				}
			},
		},
		{
			name: "ownerReferences가 없는 경우",
			obj: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]interface{}{
					"name": "test-pod",
				},
			},
			namespace:   "default",
			expectedNil: false,
			expectedCheck: func(t *testing.T, r *domain.KubernetesResource) {
				if r.OwnerReferences != nil {
					t.Errorf("OwnerReferences = %v, want nil", r.OwnerReferences)
				}
			},
		},
	}

	cfg := &config.Config{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapToResource(tt.obj, tt.namespace, cfg)

			if tt.expectedNil {
				if got != nil {
					t.Errorf("MapToResource() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Errorf("MapToResource() = nil, want non-nil")
				return
			}

			if got.Config != cfg {
				t.Errorf("Config pointer mismatch")
			}

			if tt.expectedCheck != nil {
				tt.expectedCheck(t, got)
			}
		})
	}
}

func TestGetString(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]interface{}
		key  string
		want string
	}{
		{
			name: "문자열 값 존재",
			m:    map[string]interface{}{"key": "value"},
			key:  "key",
			want: "value",
		},
		{
			name: "키가 없는 경우",
			m:    map[string]interface{}{},
			key:  "missing",
			want: "",
		},
		{
			name: "값이 문자열이 아닌 경우",
			m:    map[string]interface{}{"key": 123},
			key:  "key",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getString(tt.m, tt.key); got != tt.want {
				t.Errorf("getString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStringOr(t *testing.T) {
	tests := []struct {
		name         string
		m            map[string]interface{}
		key          string
		defaultValue string
		want         string
	}{
		{
			name:         "값이 존재하는 경우",
			m:            map[string]interface{}{"key": "value"},
			key:          "key",
			defaultValue: "default",
			want:         "value",
		},
		{
			name:         "키가 없는 경우",
			m:            map[string]interface{}{},
			key:          "missing",
			defaultValue: "default",
			want:         "default",
		},
		{
			name:         "빈 문자열인 경우",
			m:            map[string]interface{}{"key": ""},
			key:          "key",
			defaultValue: "default",
			want:         "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getStringOr(tt.m, tt.key, tt.defaultValue); got != tt.want {
				t.Errorf("getStringOr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStringMap(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]interface{}
		key  string
		want map[string]string
	}{
		{
			name: "올바른 맵",
			m: map[string]interface{}{
				"labels": map[string]interface{}{
					"app":     "test",
					"version": "v1",
				},
			},
			key: "labels",
			want: map[string]string{
				"app":     "test",
				"version": "v1",
			},
		},
		{
			name: "빈 맵",
			m: map[string]interface{}{
				"labels": map[string]interface{}{},
			},
			key:  "labels",
			want: map[string]string{},
		},
		{
			name: "키가 없는 경우",
			m:    map[string]interface{}{},
			key:  "labels",
			want: map[string]string{},
		},
		{
			name: "문자열이 아닌 값이 포함된 경우",
			m: map[string]interface{}{
				"labels": map[string]interface{}{
					"app":     "test",
					"version": 123,
					"enabled": true,
				},
			},
			key: "labels",
			want: map[string]string{
				"app": "test",
			},
		},
		{
			name: "잘못된 타입",
			m: map[string]interface{}{
				"labels": "not-a-map",
			},
			key:  "labels",
			want: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getStringMap(tt.m, tt.key)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getStringMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSlice(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]interface{}
		key  string
		want []interface{}
	}{
		{
			name: "슬라이스 존재",
			m: map[string]interface{}{
				"items": []interface{}{"a", "b", "c"},
			},
			key:  "items",
			want: []interface{}{"a", "b", "c"},
		},
		{
			name: "빈 슬라이스",
			m: map[string]interface{}{
				"items": []interface{}{},
			},
			key:  "items",
			want: []interface{}{},
		},
		{
			name: "키가 없는 경우",
			m:    map[string]interface{}{},
			key:  "items",
			want: nil,
		},
		{
			name: "잘못된 타입",
			m: map[string]interface{}{
				"items": "not-a-slice",
			},
			key:  "items",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSlice(tt.m, tt.key)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
