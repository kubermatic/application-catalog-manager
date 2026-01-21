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

package defaulting

import (
	"testing"

	catalogv1alpha1 "k8c.io/application-catalog-manager/pkg/apis/applicationcatalog/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDefaultApplicationCatalog(t *testing.T) {
	tests := []struct {
		name    string
		catalog *catalogv1alpha1.ApplicationCatalog
		verify  func(t *testing.T, catalog *catalogv1alpha1.ApplicationCatalog)
	}{
		{
			name: "includeDefaults=false, charts=nil",
			catalog: &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-catalog",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: nil,
					},
				},
			},
			verify: func(t *testing.T, catalog *catalogv1alpha1.ApplicationCatalog) {
				assertChartsNilOrEmpty(t, catalog)
			},
		},
		{
			name: "includeDefaults=false, charts=[]",
			catalog: &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-catalog",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{},
					},
				},
			},
			verify: func(t *testing.T, catalog *catalogv1alpha1.ApplicationCatalog) {
				// Empty array is preserved
				assertChartCount(t, catalog, 0)
			},
		},
		{
			name: "includeDefaults=false, charts has custom apps",
			catalog: &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "custom-catalog",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{ChartName: "my-custom-app"},
						},
					},
				},
			},
			verify: func(t *testing.T, catalog *catalogv1alpha1.ApplicationCatalog) {
				assertChartCount(t, catalog, 1)
				assertChartNames(t, catalog, []string{"my-custom-app"})
			},
		},
		{
			name: "includeDefaults=false, charts has default app (no override)",
			catalog: &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-catalog",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{ChartName: "ingress-nginx"},
						},
					},
				},
			},
			verify: func(t *testing.T, catalog *catalogv1alpha1.ApplicationCatalog) {
				// User's chart preserved even if it matches a default name
				assertChartCount(t, catalog, 1)
				assertChartNames(t, catalog, []string{"ingress-nginx"})
			},
		},
		{
			name: "includeDefaults=false, annotation set, charts=nil",
			catalog: &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-catalog",
					Annotations: map[string]string{
						"defaultcatalog.k8c.io/include": "ingress-nginx,cert-manager",
					},
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: nil,
					},
				},
			},
			verify: func(t *testing.T, catalog *catalogv1alpha1.ApplicationCatalog) {
				// Annotation ignored when includeDefaults=false
				assertChartsNilOrEmpty(t, catalog)
			},
		},

		// includeDefaults=true cases
		{
			name: "includeDefaults=true, no annotation, charts=nil",
			catalog: &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-catalog",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts:          nil,
						IncludeDefaults: true,
					},
				},
			},
			verify: func(t *testing.T, catalog *catalogv1alpha1.ApplicationCatalog) {
				// Auto-sync with all defaults
				assertChartsNotEmpty(t, catalog)
				assertChartCount(t, catalog, len(GetDefaultCharts()))
			},
		},
		{
			name: "includeDefaults=true, no annotation, charts=[]",
			catalog: &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-catalog",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts:          []catalogv1alpha1.ChartConfig{},
						IncludeDefaults: true,
					},
				},
			},
			verify: func(t *testing.T, catalog *catalogv1alpha1.ApplicationCatalog) {
				// Empty array + defaults = defaults only
				assertChartCount(t, catalog, len(GetDefaultCharts()))
			},
		},
		{
			name: "includeDefaults=true, no annotation, charts=[custom-only]",
			catalog: &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-catalog",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{ChartName: "my-custom-app"},
						},
						IncludeDefaults: true,
					},
				},
			},
			verify: func(t *testing.T, catalog *catalogv1alpha1.ApplicationCatalog) {
				// Custom app + all defaults
				assertChartNamesContains(t, catalog, []string{"my-custom-app"})
				assertChartCount(t, catalog, len(GetDefaultCharts())+1)
			},
		},
		{
			name: "includeDefaults=true, no annotation, charts=[default-only]",
			catalog: &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-catalog",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{ChartName: "ingress-nginx"},
						},
						IncludeDefaults: true,
					},
				},
			},
			verify: func(t *testing.T, catalog *catalogv1alpha1.ApplicationCatalog) {
				// Default app from user + all defaults = one ingress-nginx (from defaults)
				assertChartCount(t, catalog, len(GetDefaultCharts()))
			},
		},
		{
			name: "includeDefaults=true, no annotation, charts=[default-with-override]",
			catalog: &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-catalog",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{
								ChartName: "ingress-nginx",
								RepositorySettings: &catalogv1alpha1.RepositorySettings{
									BaseURL: "https://my-registry.com/charts",
								},
							},
						},
						IncludeDefaults: true,
					},
				},
			},
			verify: func(t *testing.T, catalog *catalogv1alpha1.ApplicationCatalog) {
				// User's override wins
				assertChartCount(t, catalog, len(GetDefaultCharts()))
				ingressNginx := findChart(t, catalog.Spec.Helm.Charts, "ingress-nginx")
				if ingressNginx.RepositorySettings == nil {
					t.Errorf("User's RepositorySettings override not preserved")
				} else if ingressNginx.RepositorySettings.BaseURL != "https://my-registry.com/charts" {
					t.Errorf("User's override not applied: got %s, want https://my-registry.com/charts",
						ingressNginx.RepositorySettings.BaseURL)
				}
			},
		},
		{
			name: "includeDefaults=true, no annotation, charts=[mixed]",
			catalog: &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-catalog",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{ChartName: "my-custom-app-1"},
							{ChartName: "ingress-nginx"},
						},
						IncludeDefaults: true,
					},
				},
			},
			verify: func(t *testing.T, catalog *catalogv1alpha1.ApplicationCatalog) {
				// Mix of custom and default apps
				assertChartNamesContains(t, catalog, []string{"my-custom-app-1", "ingress-nginx"})
			},
		},
		{
			name: "includeDefaults=true, annotation empty, charts=nil",
			catalog: &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-catalog",
					Annotations: map[string]string{
						"defaultcatalog.k8c.io/include": "",
					},
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts:          nil,
						IncludeDefaults: true,
					},
				},
			},
			verify: func(t *testing.T, catalog *catalogv1alpha1.ApplicationCatalog) {
				// Empty annotation = no filter (all defaults)
				assertChartsNotEmpty(t, catalog)
				assertChartCount(t, catalog, len(GetDefaultCharts()))
			},
		},
		{
			name: "includeDefaults=true, annotation=nginx,cert-manager, charts=nil",
			catalog: &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-catalog",
					Annotations: map[string]string{
						"defaultcatalog.k8c.io/include": "ingress-nginx,cert-manager",
					},
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts:          nil,
						IncludeDefaults: true,
					},
				},
			},
			verify: func(t *testing.T, catalog *catalogv1alpha1.ApplicationCatalog) {
				assertChartCount(t, catalog, 2)
				assertChartNames(t, catalog, []string{"cert-manager", "ingress-nginx"})
			},
		},
		{
			name: "includeDefaults=true, annotation=nginx,cert-manager, charts=[]",
			catalog: &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-catalog",
					Annotations: map[string]string{
						"defaultcatalog.k8c.io/include": "ingress-nginx,cert-manager",
					},
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts:          []catalogv1alpha1.ChartConfig{},
						IncludeDefaults: true,
					},
				},
			},
			verify: func(t *testing.T, catalog *catalogv1alpha1.ApplicationCatalog) {
				// Empty array + filtered defaults
				assertChartCount(t, catalog, 2)
				assertChartNames(t, catalog, []string{"cert-manager", "ingress-nginx"})
			},
		},
		{
			name: "includeDefaults=true, annotation=nginx,cert-manager, charts=[custom-only]",
			catalog: &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-catalog",
					Annotations: map[string]string{
						"defaultcatalog.k8c.io/include": "ingress-nginx,cert-manager",
					},
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{ChartName: "my-custom-app"},
						},
						IncludeDefaults: true,
					},
				},
			},
			verify: func(t *testing.T, catalog *catalogv1alpha1.ApplicationCatalog) {
				// Custom app + filtered defaults
				assertChartCount(t, catalog, 3)
				assertChartNames(t, catalog, []string{"cert-manager", "ingress-nginx", "my-custom-app"})
			},
		},
		{
			name: "includeDefaults=true, annotation=nginx,cert-manager, charts=[nginx-with-override]",
			catalog: &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-catalog",
					Annotations: map[string]string{
						"defaultcatalog.k8c.io/include": "ingress-nginx,cert-manager",
					},
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{
								ChartName: "ingress-nginx",
								RepositorySettings: &catalogv1alpha1.RepositorySettings{
									BaseURL: "https://my-registry.com/charts",
								},
							},
						},
						IncludeDefaults: true,
					},
				},
			},
			verify: func(t *testing.T, catalog *catalogv1alpha1.ApplicationCatalog) {
				assertChartCount(t, catalog, 2)
				assertChartNames(t, catalog, []string{"cert-manager", "ingress-nginx"})

				ingressNginx := findChart(t, catalog.Spec.Helm.Charts, "ingress-nginx")
				if ingressNginx.RepositorySettings == nil {
					t.Errorf("User's RepositorySettings override not preserved")
				} else if ingressNginx.RepositorySettings.BaseURL != "https://my-registry.com/charts" {
					t.Errorf("User's override not applied: got %s, want https://my-registry.com/charts",
						ingressNginx.RepositorySettings.BaseURL)
				}
			},
		},
		{
			name: "includeDefaults=true, annotation=non-existent, charts=nil",
			catalog: &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-catalog",
					Annotations: map[string]string{
						"defaultcatalog.k8c.io/include": "non-existent-app",
					},
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts:          nil,
						IncludeDefaults: true,
					},
				},
			},
			verify: func(t *testing.T, catalog *catalogv1alpha1.ApplicationCatalog) {
				// Non-existent app in annotation = empty result
				assertChartCount(t, catalog, 0)
			},
		},
		{
			name: "includeDefaults=true, annotation=nginx,non-existent,cert-manager, charts=nil",
			catalog: &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-catalog",
					Annotations: map[string]string{
						"defaultcatalog.k8c.io/include": "ingress-nginx,non-existent,cert-manager",
					},
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts:          nil,
						IncludeDefaults: true,
					},
				},
			},
			verify: func(t *testing.T, catalog *catalogv1alpha1.ApplicationCatalog) {
				// Non-existent app ignored, other apps included
				assertChartCount(t, catalog, 2)
				assertChartNames(t, catalog, []string{"cert-manager", "ingress-nginx"})
			},
		},
		{
			name: "includeDefaults=true, annotation with whitespace, charts=nil",
			catalog: &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-catalog",
					Annotations: map[string]string{
						"defaultcatalog.k8c.io/include": "ingress-nginx , cert-manager , argo-cd",
					},
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts:          nil,
						IncludeDefaults: true,
					},
				},
			},
			verify: func(t *testing.T, catalog *catalogv1alpha1.ApplicationCatalog) {
				assertChartCount(t, catalog, 3)
				assertChartNames(t, catalog, []string{"argo-cd", "cert-manager", "ingress-nginx"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DefaultApplicationCatalog(tt.catalog)
			tt.verify(t, tt.catalog)
		})
	}
}

// assertChartsNilOrEmpty verifies that no defaults were injected.
func assertChartsNilOrEmpty(t *testing.T, catalog *catalogv1alpha1.ApplicationCatalog) {
	t.Helper()

	if catalog.Spec.Helm == nil || catalog.Spec.Helm.Charts == nil {
		return
	}
	if len(catalog.Spec.Helm.Charts) > 0 {
		t.Errorf("Expected no charts to be injected, but got %d charts", len(catalog.Spec.Helm.Charts))
	}
}

// assertChartsNotEmpty verifies that charts were injected.
func assertChartsNotEmpty(t *testing.T, catalog *catalogv1alpha1.ApplicationCatalog) {
	t.Helper()

	if catalog.Spec.Helm == nil || catalog.Spec.Helm.Charts == nil {
		t.Fatal("Expected charts to be injected, but got nil")
	}
	if len(catalog.Spec.Helm.Charts) == 0 {
		t.Errorf("Expected at least one chart, got 0")
	}
}

// assertChartCount verifies the exact number of charts.
func assertChartCount(t *testing.T, catalog *catalogv1alpha1.ApplicationCatalog, want int) {
	t.Helper()

	if catalog.Spec.Helm == nil || catalog.Spec.Helm.Charts == nil {
		t.Fatalf("Expected charts to be present, but got nil")
	}
	if got := len(catalog.Spec.Helm.Charts); got != want {
		t.Errorf("Chart count mismatch: got %d, want %d", got, want)
	}
}

// assertChartNames verifies the exact set of chart names (order matters).
func assertChartNames(t *testing.T, catalog *catalogv1alpha1.ApplicationCatalog, want []string) {
	t.Helper()

	if catalog.Spec.Helm == nil || catalog.Spec.Helm.Charts == nil {
		t.Fatalf("Expected charts to be present, but got nil")
	}

	got := make([]string, len(catalog.Spec.Helm.Charts))
	for i, chart := range catalog.Spec.Helm.Charts {
		got[i] = chart.ChartName
	}

	if !equalStringSlices(got, want) {
		t.Errorf("Chart names mismatch:\ngot  %v\nwant %v", got, want)
	}
}

// assertChartNamesContains verifies that all specified chart names are present.
func assertChartNamesContains(t *testing.T, catalog *catalogv1alpha1.ApplicationCatalog, required []string) {
	t.Helper()

	if catalog.Spec.Helm == nil || catalog.Spec.Helm.Charts == nil {
		t.Fatalf("Expected charts to be present, but got nil")
	}

	chartNames := make(map[string]bool, len(catalog.Spec.Helm.Charts))
	for _, chart := range catalog.Spec.Helm.Charts {
		chartNames[chart.ChartName] = true
	}

	for _, name := range required {
		if !chartNames[name] {
			t.Errorf("Expected chart %q to be present, but it was not found", name)
		}
	}
}

// equalStringSlices compares two string slices for equality.
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// findChart returns the chart with the given name or fails the test if not found.
func findChart(t *testing.T, charts []catalogv1alpha1.ChartConfig, name string) *catalogv1alpha1.ChartConfig {
	t.Helper()

	for i := range charts {
		if charts[i].ChartName == name {
			return &charts[i]
		}
	}
	t.Fatalf("Chart %q not found", name)
	return nil
}
