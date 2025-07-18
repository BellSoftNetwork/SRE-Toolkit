package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ArgoCD        ArgoCDConfig        `yaml:"argocd"`
	Exclusions    map[string][]string `yaml:"exclusions"`
	AutoManaged   AutoManagedConfig   `yaml:"auto_managed"`
	Patterns      PatternsConfig      `yaml:"patterns"`
	ResourceTypes ResourceTypesConfig `yaml:"resource_types"`
	Performance   PerformanceConfig   `yaml:"performance"`

	ExclusionRules         []ExclusionRule
	SecretPatterns         []*regexp.Regexp
	RancherManagedPatterns map[string][]*regexp.Regexp
	AutoManagedAnnotations map[string]bool
	CertManagerAnnotations map[string]bool
	StatefulSetPVCPattern  *regexp.Regexp
	SkipResourceTypes      map[string]bool
	ImportantResourceTypes []string
	BatchSize              int
}

type ArgoCDConfig struct {
	ManagedLabels   []string `yaml:"managed_labels"`
	SyncAnnotations []string `yaml:"sync_annotations"`
}

type AutoManagedConfig struct {
	Annotations            []string `yaml:"annotations"`
	CertManagerAnnotations []string `yaml:"cert_manager_annotations"`
}

type PatternsConfig struct {
	SecretPatterns []string            `yaml:"secret_patterns"`
	RancherManaged map[string][]string `yaml:"rancher_managed"`
	StatefulSetPVC string              `yaml:"statefulset_pvc"`
}

type ResourceTypesConfig struct {
	Skip      []string `yaml:"skip"`
	Important []string `yaml:"important"`
}

type PerformanceConfig struct {
	DefaultMaxConcurrent int `yaml:"default_max_concurrent"`
	FastScanConcurrent   int `yaml:"fast_scan_concurrent"`
	BatchSize            int `yaml:"batch_size"`
}

var DefaultMaxConcurrent = 10

func NewDefaultConfig() *Config {
	return &Config{}
}

func LoadConfigFromFile(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	var exclusionRules []ExclusionRule
	for _, patterns := range cfg.Exclusions {
		for _, pattern := range patterns {
			parts := strings.Split(pattern, "/")
			if len(parts) != 3 {
				continue
			}
			exclusionRules = append(exclusionRules, ExclusionRule{
				Namespace: parts[0],
				Kind:      parts[1],
				Name:      parts[2],
				Pattern:   pattern,
			})
		}
	}
	cfg.ExclusionRules = exclusionRules

	for _, pattern := range cfg.Patterns.SecretPatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid secret pattern %s: %w", pattern, err)
		}
		cfg.SecretPatterns = append(cfg.SecretPatterns, re)
	}

	cfg.RancherManagedPatterns = make(map[string][]*regexp.Regexp)
	for resourceType, patterns := range cfg.Patterns.RancherManaged {
		for _, pattern := range patterns {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid rancher pattern %s: %w", pattern, err)
			}
			cfg.RancherManagedPatterns[resourceType] = append(cfg.RancherManagedPatterns[resourceType], re)
		}
	}

	if cfg.Patterns.StatefulSetPVC != "" {
		re, err := regexp.Compile(cfg.Patterns.StatefulSetPVC)
		if err != nil {
			return nil, fmt.Errorf("invalid StatefulSet PVC pattern: %w", err)
		}
		cfg.StatefulSetPVCPattern = re
	}

	cfg.AutoManagedAnnotations = make(map[string]bool)
	for _, ann := range cfg.AutoManaged.Annotations {
		cfg.AutoManagedAnnotations[ann] = true
	}

	cfg.CertManagerAnnotations = make(map[string]bool)
	for _, ann := range cfg.AutoManaged.CertManagerAnnotations {
		cfg.CertManagerAnnotations[ann] = true
	}

	cfg.SkipResourceTypes = make(map[string]bool)
	for _, rt := range cfg.ResourceTypes.Skip {
		cfg.SkipResourceTypes[rt] = true
	}

	cfg.ImportantResourceTypes = cfg.ResourceTypes.Important
	cfg.BatchSize = cfg.Performance.BatchSize
	DefaultMaxConcurrent = cfg.Performance.DefaultMaxConcurrent

	return &cfg, nil
}

func (c *Config) GetManagedLabels() []string {
	return c.ArgoCD.ManagedLabels
}

func (c *Config) GetSyncAnnotations() []string {
	return c.ArgoCD.SyncAnnotations
}
