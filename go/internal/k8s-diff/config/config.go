package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 애플리케이션 설정
type Config struct {
	ExclusionRules          []ExclusionRule `yaml:"exclusion_rules"`
	ResourceTypes           []string        `yaml:"resource_types"`
	ImportantResourceTypes  []string        `yaml:"important_resource_types"`
	SkipResourceTypes       []string        `yaml:"skip_resource_types"`
	BatchSize               int             `yaml:"batch_size"`
	MaxConcurrent           int             `yaml:"max_concurrent"`
	CompareResourceContents bool            `yaml:"compare_resource_contents"`
	StrictAPIVersion        bool            `yaml:"strict_api_version"`
}

// ExclusionRule 제외 규칙
type ExclusionRule struct {
	Namespace string `yaml:"namespace"`
	Kind      string `yaml:"kind"`
	Name      string `yaml:"name"`
}

// Match 규칙 매칭 여부 확인
func (r ExclusionRule) Match(namespace, kind, name string) bool {
	if !matchPattern(r.Namespace, namespace) {
		return false
	}
	if !matchPattern(r.Kind, kind) {
		return false
	}
	if !matchPattern(r.Name, name) {
		return false
	}
	return true
}

// matchPattern 와일드카드 패턴 매칭
func matchPattern(pattern, value string) bool {
	if pattern == "*" {
		return true
	}

	// 간단한 와일드카드 지원 (prefix-*)
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(value, prefix)
	}

	return pattern == value
}

// DefaultConfig 기본 설정 반환
func DefaultConfig() *Config {
	// 기본값은 최소한만 설정
	return &Config{
		ExclusionRules:          []ExclusionRule{},
		ResourceTypes:           []string{}, // 비어있으면 모든 리소스 타입 검색
		ImportantResourceTypes:  []string{},
		SkipResourceTypes:       []string{},
		BatchSize:               10,
		MaxConcurrent:           20,
		CompareResourceContents: false,
		StrictAPIVersion:        false, // 기본적으로 Kind만 비교
	}
}

// LoadConfigFromFile 파일에서 설정 로드
func LoadConfigFromFile(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("설정 파일 읽기 실패: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("YAML 파싱 실패: %w", err)
	}

	// 기본값 설정
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 10
	}
	if cfg.MaxConcurrent == 0 {
		cfg.MaxConcurrent = 20
	}

	return cfg, nil
}
