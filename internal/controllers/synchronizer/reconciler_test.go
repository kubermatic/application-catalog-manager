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

package synchronizer

import (
	"testing"

	catalogv1alpha1 "k8c.io/application-catalog-manager/pkg/apis/applicationcatalog/v1alpha1"
	appskubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/apps.kubermatic/v1"
)

func TestMergeVersions(t *testing.T) {
	tests := []struct {
		name     string
		src      []appskubermaticv1.ApplicationVersion
		dst      []appskubermaticv1.ApplicationVersion
		expected []appskubermaticv1.ApplicationVersion
	}{
		{
			name: "empty src returns dst",
			src:  nil,
			dst: []appskubermaticv1.ApplicationVersion{
				{Version: "v1.0.0"},
			},
			expected: []appskubermaticv1.ApplicationVersion{
				{Version: "v1.0.0"},
			},
		},
		{
			name: "empty dst returns src",
			src: []appskubermaticv1.ApplicationVersion{
				{Version: "v1.0.0"},
			},
			dst: nil,
			expected: []appskubermaticv1.ApplicationVersion{
				{Version: "v1.0.0"},
			},
		},
		{
			name:     "both empty returns nil",
			src:      nil,
			dst:      nil,
			expected: nil,
		},
		{
			name: "existing versions are preserved when not in dst",
			src: []appskubermaticv1.ApplicationVersion{
				{Version: "v1.0.0"},
				{Version: "v2.0.0"},
			},
			dst: []appskubermaticv1.ApplicationVersion{
				{Version: "v2.0.0"},
			},
			expected: []appskubermaticv1.ApplicationVersion{
				{Version: "v1.0.0"},
				{Version: "v2.0.0"},
			},
		},
		{
			name: "new versions are appended",
			src: []appskubermaticv1.ApplicationVersion{
				{Version: "v1.0.0"},
			},
			dst: []appskubermaticv1.ApplicationVersion{
				{Version: "v1.0.0"},
				{Version: "v2.0.0"},
			},
			expected: []appskubermaticv1.ApplicationVersion{
				{Version: "v1.0.0"},
				{Version: "v2.0.0"},
			},
		},
		{
			name: "existing versions are updated with desired state",
			src: []appskubermaticv1.ApplicationVersion{
				{
					Version: "v1.0.0",
					Template: appskubermaticv1.ApplicationTemplate{
						Source: appskubermaticv1.ApplicationSource{
							Helm: &appskubermaticv1.HelmSource{
								URL:          "https://old-repo.example.com",
								ChartName:    "my-chart",
								ChartVersion: "1.0.0",
							},
						},
					},
				},
			},
			dst: []appskubermaticv1.ApplicationVersion{
				{
					Version: "v1.0.0",
					Template: appskubermaticv1.ApplicationTemplate{
						Source: appskubermaticv1.ApplicationSource{
							Helm: &appskubermaticv1.HelmSource{
								URL:          "https://new-repo.example.com",
								ChartName:    "my-chart",
								ChartVersion: "1.0.0",
							},
						},
					},
				},
			},
			expected: []appskubermaticv1.ApplicationVersion{
				{
					Version: "v1.0.0",
					Template: appskubermaticv1.ApplicationTemplate{
						Source: appskubermaticv1.ApplicationSource{
							Helm: &appskubermaticv1.HelmSource{
								URL:          "https://new-repo.example.com",
								ChartName:    "my-chart",
								ChartVersion: "1.0.0",
							},
						},
					},
				},
			},
		},
		{
			name: "complex scenario: preserve, update, and add",
			src: []appskubermaticv1.ApplicationVersion{
				{
					Version: "v1.0.0",
					Template: appskubermaticv1.ApplicationTemplate{
						Source: appskubermaticv1.ApplicationSource{
							Helm: &appskubermaticv1.HelmSource{URL: "https://old.example.com"},
						},
					},
				},
				{
					Version: "v2.0.0",
					Template: appskubermaticv1.ApplicationTemplate{
						Source: appskubermaticv1.ApplicationSource{
							Helm: &appskubermaticv1.HelmSource{URL: "https://old.example.com"},
						},
					},
				},
				{
					Version: "v3.0.0",
					Template: appskubermaticv1.ApplicationTemplate{
						Source: appskubermaticv1.ApplicationSource{
							Helm: &appskubermaticv1.HelmSource{URL: "https://old.example.com"},
						},
					},
				},
			},
			dst: []appskubermaticv1.ApplicationVersion{
				{
					Version: "v2.0.0",
					Template: appskubermaticv1.ApplicationTemplate{
						Source: appskubermaticv1.ApplicationSource{
							Helm: &appskubermaticv1.HelmSource{URL: "https://new.example.com"},
						},
					},
				},
				{
					Version: "v4.0.0",
					Template: appskubermaticv1.ApplicationTemplate{
						Source: appskubermaticv1.ApplicationSource{
							Helm: &appskubermaticv1.HelmSource{URL: "https://new.example.com"},
						},
					},
				},
			},
			expected: []appskubermaticv1.ApplicationVersion{
				{
					Version: "v1.0.0",
					Template: appskubermaticv1.ApplicationTemplate{
						Source: appskubermaticv1.ApplicationSource{
							Helm: &appskubermaticv1.HelmSource{URL: "https://old.example.com"},
						},
					},
				},
				{
					Version: "v2.0.0",
					Template: appskubermaticv1.ApplicationTemplate{
						Source: appskubermaticv1.ApplicationSource{
							Helm: &appskubermaticv1.HelmSource{URL: "https://new.example.com"},
						},
					},
				},
				{
					Version: "v3.0.0",
					Template: appskubermaticv1.ApplicationTemplate{
						Source: appskubermaticv1.ApplicationSource{
							Helm: &appskubermaticv1.HelmSource{URL: "https://old.example.com"},
						},
					},
				},
				{
					Version: "v4.0.0",
					Template: appskubermaticv1.ApplicationTemplate{
						Source: appskubermaticv1.ApplicationSource{
							Helm: &appskubermaticv1.HelmSource{URL: "https://new.example.com"},
						},
					},
				},
			},
		},
		{
			name: "idempotency: same input produces same output",
			src: []appskubermaticv1.ApplicationVersion{
				{Version: "v1.0.0"},
				{Version: "v2.0.0"},
			},
			dst: []appskubermaticv1.ApplicationVersion{
				{Version: "v1.0.0"},
				{Version: "v2.0.0"},
			},
			expected: []appskubermaticv1.ApplicationVersion{
				{Version: "v1.0.0"},
				{Version: "v2.0.0"},
			},
		},
		{
			name: "empty slice src (not nil) returns dst",
			src:  []appskubermaticv1.ApplicationVersion{},
			dst: []appskubermaticv1.ApplicationVersion{
				{Version: "v1.0.0"},
			},
			expected: []appskubermaticv1.ApplicationVersion{
				{Version: "v1.0.0"},
			},
		},
		{
			name: "empty slice dst (not nil) returns src",
			src: []appskubermaticv1.ApplicationVersion{
				{Version: "v1.0.0"},
			},
			dst: []appskubermaticv1.ApplicationVersion{},
			expected: []appskubermaticv1.ApplicationVersion{
				{Version: "v1.0.0"},
			},
		},
		{
			name: "duplicate versions in dst - last one wins",
			src: []appskubermaticv1.ApplicationVersion{
				{
					Version: "v1.0.0",
					Template: appskubermaticv1.ApplicationTemplate{
						Source: appskubermaticv1.ApplicationSource{
							Helm: &appskubermaticv1.HelmSource{URL: "https://original.example.com"},
						},
					},
				},
			},
			dst: []appskubermaticv1.ApplicationVersion{
				{
					Version: "v1.0.0",
					Template: appskubermaticv1.ApplicationTemplate{
						Source: appskubermaticv1.ApplicationSource{
							Helm: &appskubermaticv1.HelmSource{URL: "https://first.example.com"},
						},
					},
				},
				{
					Version: "v1.0.0",
					Template: appskubermaticv1.ApplicationTemplate{
						Source: appskubermaticv1.ApplicationSource{
							Helm: &appskubermaticv1.HelmSource{URL: "https://second.example.com"},
						},
					},
				},
			},
			expected: []appskubermaticv1.ApplicationVersion{
				{
					Version: "v1.0.0",
					Template: appskubermaticv1.ApplicationTemplate{
						Source: appskubermaticv1.ApplicationSource{
							Helm: &appskubermaticv1.HelmSource{URL: "https://second.example.com"},
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := mergeVersions(tc.src, tc.dst)

			if len(result) != len(tc.expected) {
				t.Fatalf("expected %d versions, got %d", len(tc.expected), len(result))
			}

			for i, expected := range tc.expected {
				if result[i].Version != expected.Version {
					t.Errorf("version[%d]: expected Version %q, got %q", i, expected.Version, result[i].Version)
				}
				if expected.Template.Source.Helm != nil {
					if result[i].Template.Source.Helm == nil {
						t.Errorf("version[%d]: expected Helm source, got nil", i)
						continue
					}
					if result[i].Template.Source.Helm.URL != expected.Template.Source.Helm.URL {
						t.Errorf("version[%d]: expected URL %q, got %q",
							i, expected.Template.Source.Helm.URL, result[i].Template.Source.Helm.URL)
					}
				}
			}
		})
	}
}

func TestMergeVersionsDoesNotModifyInput(t *testing.T) {
	src := []appskubermaticv1.ApplicationVersion{
		{Version: "v1.0.0"},
		{Version: "v2.0.0"},
	}
	dst := []appskubermaticv1.ApplicationVersion{
		{Version: "v2.0.0"},
		{Version: "v3.0.0"},
	}

	srcCopy := make([]appskubermaticv1.ApplicationVersion, len(src))
	copy(srcCopy, src)
	dstCopy := make([]appskubermaticv1.ApplicationVersion, len(dst))
	copy(dstCopy, dst)

	_ = mergeVersions(src, dst)

	for i := range src {
		if src[i].Version != srcCopy[i].Version {
			t.Errorf("src was modified: expected %v, got %v", srcCopy, src)
		}
	}
	for i := range dst {
		if dst[i].Version != dstCopy[i].Version {
			t.Errorf("dst was modified: expected %v, got %v", dstCopy, dst)
		}
	}
}

func TestMergeVersionsPreservesOrder(t *testing.T) {
	src := []appskubermaticv1.ApplicationVersion{
		{Version: "v1.0.0"},
		{Version: "v2.0.0"},
		{Version: "v3.0.0"},
	}
	dst := []appskubermaticv1.ApplicationVersion{
		{Version: "v3.0.0"},
		{Version: "v4.0.0"},
	}

	result := mergeVersions(src, dst)

	expectedOrder := []string{"v1.0.0", "v2.0.0", "v3.0.0", "v4.0.0"}
	if len(result) != len(expectedOrder) {
		t.Fatalf("expected %d versions, got %d", len(expectedOrder), len(result))
	}

	for i, expected := range expectedOrder {
		if result[i].Version != expected {
			t.Errorf("position %d: expected %q, got %q", i, expected, result[i].Version)
		}
	}
}

func TestReconcileChartNilBehavior(t *testing.T) {
	tests := []struct {
		name            string
		includeDefaults bool
		charts          []catalogv1alpha1.ChartConfig
		helmSpecNil     bool
		wantRequeue     bool
		description     string
	}{
		{
			name:            "nil charts with includeDefaults true requeues",
			includeDefaults: true,
			charts:          nil,
			helmSpecNil:     false,
			wantRequeue:     true,
			description:     "Webhook should have populated defaults but hasn't run yet",
		},
		{
			name:            "nil charts with includeDefaults false processes as empty",
			includeDefaults: false,
			charts:          nil,
			helmSpecNil:     false,
			wantRequeue:     false,
			description:     "Valid empty catalog, user doesn't want defaults",
		},
		{
			name:            "nil helm spec processes as empty",
			includeDefaults: false,
			charts:          nil,
			helmSpecNil:     true,
			wantRequeue:     false,
			description:     "No helm configuration, treat as empty catalog",
		},
		{
			name:            "empty slice with includeDefaults true processes",
			includeDefaults: true,
			charts:          []catalogv1alpha1.ChartConfig{},
			helmSpecNil:     false,
			wantRequeue:     false,
			description:     "Explicitly empty catalog, webhook already ran",
		},
		{
			name:            "empty slice with includeDefaults false processes",
			includeDefaults: false,
			charts:          []catalogv1alpha1.ChartConfig{},
			helmSpecNil:     false,
			wantRequeue:     false,
			description:     "Valid empty catalog with empty slice",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			catalog := &catalogv1alpha1.ApplicationCatalog{
				Spec: catalogv1alpha1.ApplicationCatalogSpec{},
			}

			if !tc.helmSpecNil {
				catalog.Spec.Helm = &catalogv1alpha1.HelmSpec{
					IncludeDefaults: tc.includeDefaults,
					Charts:          tc.charts,
				}
			}

			charts := catalog.GetHelmCharts()

			var shouldRequeue bool
			if charts == nil && catalog.Spec.Helm != nil && catalog.Spec.Helm.IncludeDefaults {
				shouldRequeue = true
			}

			if charts == nil && !tc.helmSpecNil && !catalog.Spec.Helm.IncludeDefaults {
				charts = []catalogv1alpha1.ChartConfig{}
			}

			if charts == nil && tc.helmSpecNil {
				charts = []catalogv1alpha1.ChartConfig{}
			}

			if shouldRequeue != tc.wantRequeue {
				t.Errorf("%s: expected requeue=%v, got requeue=%v\n%s",
					tc.name, tc.wantRequeue, shouldRequeue, tc.description)
			}

			if !tc.wantRequeue && charts == nil {
				t.Errorf("%s: expected charts to be converted to empty slice, got nil", tc.name)
			}
		})
	}
}

func TestReconcileNilToEmptySliceIdempotency(t *testing.T) {
	catalog := &catalogv1alpha1.ApplicationCatalog{
		Spec: catalogv1alpha1.ApplicationCatalogSpec{
			Helm: &catalogv1alpha1.HelmSpec{
				IncludeDefaults: false,
				Charts:          nil,
			},
		},
	}

	firstPassCharts := catalog.GetHelmCharts()
	if firstPassCharts != nil {
		t.Errorf("First pass: expected nil charts, got non-nil")
	}

	if firstPassCharts == nil && catalog.Spec.Helm != nil && !catalog.Spec.Helm.IncludeDefaults {
		firstPassCharts = []catalogv1alpha1.ChartConfig{}
	}

	if firstPassCharts == nil {
		t.Errorf("First pass: charts should be converted to empty slice")
	}

	if len(firstPassCharts) != 0 {
		t.Errorf("First pass: expected 0 charts, got %d", len(firstPassCharts))
	}

	secondPassCharts := firstPassCharts
	if secondPassCharts == nil {
		t.Errorf("Second pass: charts should not be nil")
	}

	if len(secondPassCharts) != 0 {
		t.Errorf("Second pass: expected 0 charts, got %d", len(secondPassCharts))
	}
}

func TestGetHelmChartsReturnsCorrectly(t *testing.T) {
	tests := []struct {
		name       string
		helmSpec   *catalogv1alpha1.HelmSpec
		wantNil    bool
		wantLength int
	}{
		{
			name:     "nil helm spec returns nil",
			helmSpec: nil,
			wantNil:  true,
		},
		{
			name: "nil charts returns nil",
			helmSpec: &catalogv1alpha1.HelmSpec{
				Charts: nil,
			},
			wantNil: true,
		},
		{
			name: "empty charts returns empty slice",
			helmSpec: &catalogv1alpha1.HelmSpec{
				Charts: []catalogv1alpha1.ChartConfig{},
			},
			wantNil:    false,
			wantLength: 0,
		},
		{
			name: "charts with items returns slice",
			helmSpec: &catalogv1alpha1.HelmSpec{
				Charts: []catalogv1alpha1.ChartConfig{
					{ChartName: "chart1"},
					{ChartName: "chart2"},
				},
			},
			wantNil:    false,
			wantLength: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			catalog := &catalogv1alpha1.ApplicationCatalog{
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: tc.helmSpec,
				},
			}

			result := catalog.GetHelmCharts()

			if tc.wantNil {
				if result != nil {
					t.Errorf("expected nil, got non-nil slice with length %d", len(result))
				}
			} else {
				if result == nil {
					t.Errorf("expected non-nil slice, got nil")
				} else if len(result) != tc.wantLength {
					t.Errorf("expected length %d, got %d", tc.wantLength, len(result))
				}
			}
		})
	}
}
