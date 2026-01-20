/*
Copyright 2025 The Application Catalog Manager contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func newCatalogWithGlobalSettings(settings *RepositorySettings) *ApplicationCatalog {
	return &ApplicationCatalog{
		Spec: ApplicationCatalogSpec{
			Helm: &HelmSpec{
				RepositorySettings: settings,
			},
		},
	}
}

func newChartConfig(name string, settings *RepositorySettings) *ChartConfig {
	return &ChartConfig{
		ChartName:          name,
		RepositorySettings: settings,
	}
}

func newVersion(chartVersion, appVersion string, settings *RepositorySettings) *ChartVersion {
	return &ChartVersion{
		ChartVersion:       chartVersion,
		AppVersion:         appVersion,
		RepositorySettings: settings,
	}
}

func newCredentials(secretName string) *RepositoryCredentials {
	return &RepositoryCredentials{
		Username: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: secretName},
			Key:                  "username",
		},
		Password: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: secretName},
			Key:                  "password",
		},
	}
}

func TestResolveChartURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		catalog     *ApplicationCatalog
		chart       *ChartConfig
		version     *ChartVersion
		expectedURL string
	}{
		{
			name:        "default URL when nothing specified",
			catalog:     &ApplicationCatalog{},
			chart:       newChartConfig("test-chart", nil),
			version:     newVersion("1.0.0", "v1.0.0", nil),
			expectedURL: DefaultHelmRepository,
		},
		{
			name:        "nil helm spec uses default",
			catalog:     &ApplicationCatalog{Spec: ApplicationCatalogSpec{Helm: nil}},
			chart:       newChartConfig("test-chart", nil),
			version:     newVersion("1.0.0", "v1.0.0", nil),
			expectedURL: DefaultHelmRepository,
		},
		{
			name:        "nil global settings uses default",
			catalog:     newCatalogWithGlobalSettings(nil),
			chart:       newChartConfig("test-chart", nil),
			version:     newVersion("1.0.0", "v1.0.0", nil),
			expectedURL: DefaultHelmRepository,
		},
		{
			name:        "global URL overrides default",
			catalog:     newCatalogWithGlobalSettings(&RepositorySettings{BaseURL: "oci://global.registry.io/charts"}),
			chart:       newChartConfig("test-chart", nil),
			version:     newVersion("1.0.0", "v1.0.0", nil),
			expectedURL: "oci://global.registry.io/charts",
		},
		{
			name:        "chart URL overrides global",
			catalog:     newCatalogWithGlobalSettings(&RepositorySettings{BaseURL: "oci://global.registry.io/charts"}),
			chart:       newChartConfig("test-chart", &RepositorySettings{BaseURL: "oci://chart.registry.io/charts"}),
			version:     newVersion("1.0.0", "v1.0.0", nil),
			expectedURL: "oci://chart.registry.io/charts",
		},
		{
			name:        "version URL overrides chart and global",
			catalog:     newCatalogWithGlobalSettings(&RepositorySettings{BaseURL: "oci://global.registry.io/charts"}),
			chart:       newChartConfig("test-chart", &RepositorySettings{BaseURL: "oci://chart.registry.io/charts"}),
			version:     newVersion("1.0.0", "v1.0.0", &RepositorySettings{BaseURL: "oci://version.registry.io/charts"}),
			expectedURL: "oci://version.registry.io/charts",
		},
		{
			name:        "empty version URL falls back to chart",
			catalog:     newCatalogWithGlobalSettings(&RepositorySettings{BaseURL: "oci://global.registry.io/charts"}),
			chart:       newChartConfig("test-chart", &RepositorySettings{BaseURL: "oci://chart.registry.io/charts"}),
			version:     newVersion("1.0.0", "v1.0.0", &RepositorySettings{BaseURL: ""}),
			expectedURL: "oci://chart.registry.io/charts",
		},
		{
			name:        "empty chart URL falls back to global",
			catalog:     newCatalogWithGlobalSettings(&RepositorySettings{BaseURL: "oci://global.registry.io/charts"}),
			chart:       newChartConfig("test-chart", &RepositorySettings{BaseURL: ""}),
			version:     newVersion("1.0.0", "v1.0.0", nil),
			expectedURL: "oci://global.registry.io/charts",
		},
		{
			name:        "empty global URL falls back to default",
			catalog:     newCatalogWithGlobalSettings(&RepositorySettings{BaseURL: ""}),
			chart:       newChartConfig("test-chart", nil),
			version:     newVersion("1.0.0", "v1.0.0", nil),
			expectedURL: DefaultHelmRepository,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.catalog.ResolveChartURL(tt.chart, tt.version)
			if got != tt.expectedURL {
				t.Errorf("ResolveChartURL() = %q, want %q", got, tt.expectedURL)
			}
		})
	}
}

func TestResolveChartCredentials(t *testing.T) {
	t.Parallel()

	globalCreds := newCredentials("global-secret")
	chartCreds := newCredentials("chart-secret")
	versionCreds := newCredentials("version-secret")

	tests := []struct {
		name          string
		catalog       *ApplicationCatalog
		chart         *ChartConfig
		version       *ChartVersion
		expectedCreds *RepositoryCredentials
	}{
		{
			name:          "no credentials when using default URL",
			catalog:       &ApplicationCatalog{},
			chart:         newChartConfig("test-chart", nil),
			version:       newVersion("1.0.0", "v1.0.0", nil),
			expectedCreds: nil,
		},
		{
			name:          "nil helm spec returns nil credentials",
			catalog:       &ApplicationCatalog{Spec: ApplicationCatalogSpec{Helm: nil}},
			chart:         newChartConfig("test-chart", nil),
			version:       newVersion("1.0.0", "v1.0.0", nil),
			expectedCreds: nil,
		},
		{
			name: "global credentials when global URL set",
			catalog: newCatalogWithGlobalSettings(&RepositorySettings{
				BaseURL:     "oci://global.registry.io/charts",
				Credentials: globalCreds,
			}),
			chart:         newChartConfig("test-chart", nil),
			version:       newVersion("1.0.0", "v1.0.0", nil),
			expectedCreds: globalCreds,
		},
		{
			name: "chart credentials override global when chart URL set",
			catalog: newCatalogWithGlobalSettings(&RepositorySettings{
				BaseURL:     "oci://global.registry.io/charts",
				Credentials: globalCreds,
			}),
			chart: newChartConfig("test-chart", &RepositorySettings{
				BaseURL:     "oci://chart.registry.io/charts",
				Credentials: chartCreds,
			}),
			version:       newVersion("1.0.0", "v1.0.0", nil),
			expectedCreds: chartCreds,
		},
		{
			name: "version credentials override all when version URL set",
			catalog: newCatalogWithGlobalSettings(&RepositorySettings{
				BaseURL:     "oci://global.registry.io/charts",
				Credentials: globalCreds,
			}),
			chart: newChartConfig("test-chart", &RepositorySettings{
				BaseURL:     "oci://chart.registry.io/charts",
				Credentials: chartCreds,
			}),
			version: newVersion("1.0.0", "v1.0.0", &RepositorySettings{
				BaseURL:     "oci://version.registry.io/charts",
				Credentials: versionCreds,
			}),
			expectedCreds: versionCreds,
		},
		{
			name: "no credentials when version URL set without credentials",
			catalog: newCatalogWithGlobalSettings(&RepositorySettings{
				BaseURL:     "oci://global.registry.io/charts",
				Credentials: globalCreds,
			}),
			chart: newChartConfig("test-chart", &RepositorySettings{
				BaseURL:     "oci://chart.registry.io/charts",
				Credentials: chartCreds,
			}),
			version: newVersion("1.0.0", "v1.0.0", &RepositorySettings{
				BaseURL:     "oci://version.registry.io/charts",
				Credentials: nil, // No creds at version level - does NOT inherit
			}),
			expectedCreds: nil,
		},
		{
			name: "global credentials not used when global URL empty",
			catalog: newCatalogWithGlobalSettings(&RepositorySettings{
				BaseURL:     "", // empty URL
				Credentials: globalCreds,
			}),
			chart:         newChartConfig("test-chart", nil),
			version:       newVersion("1.0.0", "v1.0.0", nil),
			expectedCreds: nil,
		},
		{
			name: "chart credentials not used when chart URL empty - falls back to global",
			catalog: newCatalogWithGlobalSettings(&RepositorySettings{
				BaseURL:     "oci://global.registry.io/charts",
				Credentials: globalCreds,
			}),
			chart: newChartConfig("test-chart", &RepositorySettings{
				BaseURL:     "", // empty URL
				Credentials: chartCreds,
			}),
			version:       newVersion("1.0.0", "v1.0.0", nil),
			expectedCreds: globalCreds, // falls back to global
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.catalog.ResolveChartCredentials(tt.chart, tt.version)
			if got != tt.expectedCreds {
				t.Errorf("ResolveChartCredentials() = %v, want %v", got, tt.expectedCreds)
			}
		})
	}
}

func TestGetAppName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		chart    *ChartConfig
		expected string
	}{
		{
			name: "returns chartName when metadata is nil",
			chart: &ChartConfig{
				ChartName: "nginx",
				Metadata:  nil,
			},
			expected: "nginx",
		},
		{
			name: "returns chartName when appName is empty",
			chart: &ChartConfig{
				ChartName: "nginx",
				Metadata: &ChartMetadata{
					AppName:     "",
					DisplayName: "NGINX",
				},
			},
			expected: "nginx",
		},
		{
			name: "returns appName when set",
			chart: &ChartConfig{
				ChartName: "gpu-operator",
				Metadata: &ChartMetadata{
					AppName:     "nvidia-gpu-operator",
					DisplayName: "NVIDIA GPU Operator",
				},
			},
			expected: "nvidia-gpu-operator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.chart.GetAppName()
			if got != tt.expected {
				t.Errorf("GetAppName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestGetHelmCharts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		catalog  *ApplicationCatalog
		expected []ChartConfig
	}{
		{
			name: "returns nil when helm is nil",
			catalog: &ApplicationCatalog{
				Spec: ApplicationCatalogSpec{Helm: nil},
			},
			expected: nil,
		},
		{
			name: "returns charts when helm is set",
			catalog: &ApplicationCatalog{
				Spec: ApplicationCatalogSpec{
					Helm: &HelmSpec{
						Charts: []ChartConfig{
							{ChartName: "chart1"},
							{ChartName: "chart2"},
						},
					},
				},
			},
			expected: []ChartConfig{
				{ChartName: "chart1"},
				{ChartName: "chart2"},
			},
		},
		{
			name: "returns empty slice when charts is empty",
			catalog: &ApplicationCatalog{
				Spec: ApplicationCatalogSpec{
					Helm: &HelmSpec{
						Charts: []ChartConfig{},
					},
				},
			},
			expected: []ChartConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.catalog.GetHelmCharts()
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("GetHelmCharts() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetGlobalRepositorySettings(t *testing.T) {
	t.Parallel()

	globalSettings := &RepositorySettings{
		BaseURL: "oci://global.registry.io/charts",
	}

	tests := []struct {
		name     string
		catalog  *ApplicationCatalog
		expected *RepositorySettings
	}{
		{
			name:     "returns nil when helm is nil",
			catalog:  &ApplicationCatalog{Spec: ApplicationCatalogSpec{Helm: nil}},
			expected: nil,
		},
		{
			name:     "returns nil when repositorySettings is nil",
			catalog:  newCatalogWithGlobalSettings(nil),
			expected: nil,
		},
		{
			name:     "returns settings when set",
			catalog:  newCatalogWithGlobalSettings(globalSettings),
			expected: globalSettings,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.catalog.GetGlobalRepositorySettings()
			if got != tt.expected {
				t.Errorf("GetGlobalRepositorySettings() = %v, want %v", got, tt.expected)
			}
		})
	}
}
