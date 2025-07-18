package config

import "testing"

func TestExclusionRuleMatch(t *testing.T) {
	tests := []struct {
		name      string
		rule      ExclusionRule
		namespace string
		kind      string
		resName   string
		want      bool
	}{
		{
			name: "정확한 일치",
			rule: ExclusionRule{
				Namespace: "kube-system",
				Kind:      "ConfigMap",
				Name:      "kube-root-ca.crt",
			},
			namespace: "kube-system",
			kind:      "ConfigMap",
			resName:   "kube-root-ca.crt",
			want:      true,
		},
		{
			name: "네임스페이스 불일치",
			rule: ExclusionRule{
				Namespace: "kube-system",
				Kind:      "ConfigMap",
				Name:      "test",
			},
			namespace: "default",
			kind:      "ConfigMap",
			resName:   "test",
			want:      false,
		},
		{
			name: "와일드카드 네임스페이스",
			rule: ExclusionRule{
				Namespace: "*",
				Kind:      "Secret",
				Name:      "test",
			},
			namespace: "any-namespace",
			kind:      "Secret",
			resName:   "test",
			want:      true,
		},
		{
			name: "와일드카드 종류",
			rule: ExclusionRule{
				Namespace: "default",
				Kind:      "*",
				Name:      "test",
			},
			namespace: "default",
			kind:      "AnyKind",
			resName:   "test",
			want:      true,
		},
		{
			name: "와일드카드 이름",
			rule: ExclusionRule{
				Namespace: "default",
				Kind:      "ConfigMap",
				Name:      "*",
			},
			namespace: "default",
			kind:      "ConfigMap",
			resName:   "any-name",
			want:      true,
		},
		{
			name: "접두사 와일드카드",
			rule: ExclusionRule{
				Namespace: "default",
				Kind:      "Secret",
				Name:      "prefix-*",
			},
			namespace: "default",
			kind:      "Secret",
			resName:   "prefix-test",
			want:      true,
		},
		{
			name: "접미사 와일드카드",
			rule: ExclusionRule{
				Namespace: "default",
				Kind:      "ConfigMap",
				Name:      "*-suffix",
			},
			namespace: "default",
			kind:      "ConfigMap",
			resName:   "test-suffix",
			want:      true,
		},
		{
			name: "중간 와일드카드",
			rule: ExclusionRule{
				Namespace: "default",
				Kind:      "Service",
				Name:      "prefix*suffix",
			},
			namespace: "default",
			kind:      "Service",
			resName:   "prefix-middle-suffix",
			want:      true,
		},
		{
			name: "모든 필드 와일드카드",
			rule: ExclusionRule{
				Namespace: "*",
				Kind:      "*",
				Name:      "*",
			},
			namespace: "any",
			kind:      "any",
			resName:   "any",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rule.Match(tt.namespace, tt.kind, tt.resName); got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		value   string
		want    bool
	}{
		{
			name:    "정확한 일치",
			pattern: "exact",
			value:   "exact",
			want:    true,
		},
		{
			name:    "정확한 불일치",
			pattern: "exact",
			value:   "different",
			want:    false,
		},
		{
			name:    "모든 값 허용",
			pattern: "*",
			value:   "anything",
			want:    true,
		},
		{
			name:    "접두사 패턴 일치",
			pattern: "prefix*",
			value:   "prefix-test",
			want:    true,
		},
		{
			name:    "접두사 패턴 불일치",
			pattern: "prefix*",
			value:   "test-prefix",
			want:    false,
		},
		{
			name:    "접미사 패턴 일치",
			pattern: "*suffix",
			value:   "test-suffix",
			want:    true,
		},
		{
			name:    "접미사 패턴 불일치",
			pattern: "*suffix",
			value:   "suffix-test",
			want:    false,
		},
		{
			name:    "중간 와일드카드 일치",
			pattern: "start*end",
			value:   "start-middle-end",
			want:    true,
		},
		{
			name:    "중간 와일드카드 불일치 - 시작",
			pattern: "start*end",
			value:   "begin-middle-end",
			want:    false,
		},
		{
			name:    "중간 와일드카드 불일치 - 끝",
			pattern: "start*end",
			value:   "start-middle-finish",
			want:    false,
		},
		{
			name:    "빈 값 처리",
			pattern: "",
			value:   "",
			want:    true,
		},
		{
			name:    "빈 패턴과 값 불일치",
			pattern: "",
			value:   "something",
			want:    false,
		},
		{
			name:    "여러 와일드카드 - 첫 번째 두 개만 처리",
			pattern: "a*b*c",
			value:   "aXbYc",
			want:    false, // 두 개의 와일드카드만 처리하므로 false
		},
		{
			name:    "와일드카드 접두사와 정확한 일치",
			pattern: "test*",
			value:   "test",
			want:    true,
		},
		{
			name:    "와일드카드 접미사와 정확한 일치",
			pattern: "*test",
			value:   "test",
			want:    true,
		},
		{
			name:    "중간 와일드카드 빈 중간 부분",
			pattern: "ab*cd",
			value:   "abcd",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchPattern(tt.pattern, tt.value); got != tt.want {
				t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.pattern, tt.value, got, tt.want)
			}
		})
	}
}
