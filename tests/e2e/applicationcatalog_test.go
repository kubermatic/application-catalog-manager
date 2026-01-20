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

package e2e_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	catalogv1alpha1 "k8c.io/application-catalog-manager/pkg/apis/applicationcatalog/v1alpha1"
	appskubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/apps.kubermatic/v1"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

type applicationCatalogSuite struct {
	suite
}

func (s *applicationCatalogSuite) setupTestCase(ctx context.Context, config *envconf.Config) error {
	if err := s.withClient(config.Client()); err != nil {
		return err
	}

	if err := catalogv1alpha1.AddToScheme(s.client.Scheme()); err != nil {
		return err
	}

	if err := s.cleanupAllApplicationCatalogs(ctx); err != nil {
		return err
	}

	if err := s.cleanupAllApplicationDefinitions(ctx); err != nil {
		return err
	}

	return nil
}

func (s *applicationCatalogSuite) cleanupAllApplicationCatalogs(ctx context.Context) error {
	if s.client == nil {
		return errClientNotInitialized
	}

	return waitFor(ctx, func(ctx context.Context) (bool, error) {
		catalogs := catalogv1alpha1.ApplicationCatalogList{}
		if err := s.client.List(ctx, &catalogs); err != nil {
			return false, err
		}

		for _, catalog := range catalogs.Items {
			if err := s.client.Delete(ctx, &catalog); err != nil && !apierrors.IsNotFound(err) {
				return false, nil
			}
		}

		if err := s.client.List(ctx, &catalogs); err != nil {
			return false, err
		}

		return len(catalogs.Items) == 0, nil
	})
}

func (s *applicationCatalogSuite) createApplicationCatalog(ctx context.Context, catalog *catalogv1alpha1.ApplicationCatalog) error {
	return waitFor(ctx, func(ctx context.Context) (bool, error) {
		if err := s.client.Create(ctx, catalog); err != nil {
			return false, nil
		}
		return true, nil
	})
}

func (s *applicationCatalogSuite) getApplicationCatalog(ctx context.Context, name string) (*catalogv1alpha1.ApplicationCatalog, error) {
	catalog := &catalogv1alpha1.ApplicationCatalog{}
	if err := s.client.Get(ctx, client.ObjectKey{Name: name}, catalog); err != nil {
		return nil, err
	}
	return catalog, nil
}

func (s *applicationCatalogSuite) updateApplicationCatalog(ctx context.Context, catalog *catalogv1alpha1.ApplicationCatalog) error {
	return s.client.Update(ctx, catalog)
}

func (s *applicationCatalogSuite) deleteApplicationCatalog(ctx context.Context, name string) error {
	catalog := &catalogv1alpha1.ApplicationCatalog{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return s.client.Delete(ctx, catalog)
}

func (s *applicationCatalogSuite) getApplicationDefinition(ctx context.Context, name string) (*appskubermaticv1.ApplicationDefinition, error) {
	appDef := &appskubermaticv1.ApplicationDefinition{}
	if err := s.client.Get(ctx, client.ObjectKey{Name: name}, appDef); err != nil {
		return nil, err
	}
	return appDef, nil
}

func (s *applicationCatalogSuite) updateApplicationDefinition(ctx context.Context, appDef *appskubermaticv1.ApplicationDefinition) error {
	return s.client.Update(ctx, appDef)
}

func (s *applicationCatalogSuite) listApplicationDefinitions(ctx context.Context) (*appskubermaticv1.ApplicationDefinitionList, error) {
	list := &appskubermaticv1.ApplicationDefinitionList{}
	if err := s.client.List(ctx, list); err != nil {
		return nil, err
	}
	return list, nil
}

func (s *applicationCatalogSuite) cleanup(ctx context.Context) error {
	catalogErr := s.cleanupAllApplicationCatalogs(ctx)
	appDefErr := s.cleanupAllApplicationDefinitions(ctx)

	if catalogErr != nil {
		return catalogErr
	}
	return appDefErr
}

func TestWebhookNilChartsInjectsDefaults(t *testing.T) {
	var s applicationCatalogSuite
	f := features.New("WebhookNilChartsInjectsDefaults")

	f.Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		err := s.setupTestCase(ctx, cfg)
		require.NoError(t, err, "failed to setup test case")
		return ctx
	}).Assess("Webhook should inject default charts when spec.helm.charts is nil",
		func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			const catalogName = "test-nil-charts"
			catalog := &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: catalogName,
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: nil,
				},
			}

			err := s.createApplicationCatalog(ctx, catalog)
			require.NoError(t, err, "failed to create ApplicationCatalog")

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				created, err := s.getApplicationCatalog(ctx, catalogName)
				if err != nil {
					return false, nil
				}

				if created.Spec.Helm == nil || created.Spec.Helm.Charts == nil {
					t.Log("waiting for webhook to inject defaults")
					return false, nil
				}

				if len(created.Spec.Helm.Charts) == 0 {
					t.Log("expected default charts to be injected")
					return false, nil
				}

				t.Logf("Webhook injected %d default charts", len(created.Spec.Helm.Charts))
				return true, nil
			})
			require.NoError(t, err, "webhook should inject default charts")

			return ctx
		},
	).Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		require.NoError(t, s.cleanup(ctx))
		return ctx
	})

	testEnv.Test(t, f.Feature())
}

func TestWebhookEmptyArrayNoDefaults(t *testing.T) {
	var s applicationCatalogSuite
	f := features.New("WebhookEmptyArrayNoDefaults")

	f.Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		err := s.setupTestCase(ctx, cfg)
		require.NoError(t, err, "failed to setup test case")
		return ctx
	}).Assess("Webhook should preserve empty array (user wants no apps)",
		func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			const catalogName = "test-empty-charts"
			catalog := &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: catalogName,
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{}, // explicitly empty
					},
				},
			}

			err := s.createApplicationCatalog(ctx, catalog)
			require.NoError(t, err, "failed to create ApplicationCatalog")

			created, err := s.getApplicationCatalog(ctx, catalogName)
			require.NoError(t, err, "failed to get ApplicationCatalog")

			require.NotNil(t, created.Spec.Helm, "helm spec should not be nil")
			require.NotNil(t, created.Spec.Helm.Charts, "charts should not be nil")
			require.Empty(t, created.Spec.Helm.Charts, "charts should remain empty")

			t.Log("Empty array preserved correctly - no defaults injected")
			return ctx
		},
	).Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		require.NoError(t, s.cleanup(ctx))
		return ctx
	})

	testEnv.Test(t, f.Feature())
}

func TestWebhookCustomChartsPreserved(t *testing.T) {
	var s applicationCatalogSuite
	f := features.New("WebhookCustomChartsPreserved")

	f.Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		err := s.setupTestCase(ctx, cfg)
		require.NoError(t, err, "failed to setup test case")
		return ctx
	}).Assess("Webhook should preserve custom charts without injecting defaults",
		func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			const (
				catalogName     = "test-custom-charts"
				customChartName = "custom-nginx"
			)
			catalog := &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: catalogName,
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{
								ChartName: customChartName,
								ChartVersions: []catalogv1alpha1.ChartVersion{
									{ChartVersion: "1.0.0", AppVersion: "1.0.0"},
								},
							},
						},
					},
				},
			}

			err := s.createApplicationCatalog(ctx, catalog)
			require.NoError(t, err, "failed to create ApplicationCatalog")

			created, err := s.getApplicationCatalog(ctx, catalogName)
			require.NoError(t, err, "failed to get ApplicationCatalog")

			require.Len(t, created.Spec.Helm.Charts, 1, "should have exactly 1 chart")
			require.Equal(t, customChartName, created.Spec.Helm.Charts[0].ChartName)
			return ctx
		},
	).Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		require.NoError(t, s.cleanup(ctx))
		return ctx
	})

	testEnv.Test(t, f.Feature())
}

func TestSyncCreatesApplicationDefinitions(t *testing.T) {
	var s applicationCatalogSuite
	f := features.New("SyncCreatesApplicationDefinitions")

	f.Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		err := s.setupTestCase(ctx, cfg)
		require.NoError(t, err, "failed to setup test case")
		return ctx
	}).Assess("Controller should create ApplicationDefinitions for each chart",
		func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			const (
				appOneName = "app-one"
				appTwoName = "app-two"
			)
			catalog := &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-sync-creates",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{
								ChartName: appOneName,
								Metadata: &catalogv1alpha1.ChartMetadata{
									DisplayName: "App One",
									Description: "First test app",
								},
								ChartVersions: []catalogv1alpha1.ChartVersion{
									{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
								},
							},
							{
								ChartName: appTwoName,
								Metadata: &catalogv1alpha1.ChartMetadata{
									DisplayName: "App Two",
									Description: "Second test app",
								},
								ChartVersions: []catalogv1alpha1.ChartVersion{
									{ChartVersion: "2.0.0", AppVersion: "v2.0.0"},
								},
							},
						},
					},
				},
			}

			err := s.createApplicationCatalog(ctx, catalog)
			require.NoError(t, err, "failed to create ApplicationCatalog")

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDefs, err := s.listApplicationDefinitions(ctx)
				if err != nil {
					return false, nil
				}

				if len(appDefs.Items) < 2 {
					t.Logf("waiting for 2 ApplicationDefinitions, got %d", len(appDefs.Items))
					return false, nil
				}

				foundApps := make(map[string]bool)
				for _, app := range appDefs.Items {
					foundApps[app.Name] = true
				}

				if !foundApps[appOneName] || !foundApps[appTwoName] {
					t.Log("expected apps not found")
					return false, nil
				}

				return true, nil
			})
			require.NoError(t, err, "ApplicationDefinitions should be created")

			t.Log("ApplicationDefinitions created successfully")
			return ctx
		},
	).Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		require.NoError(t, s.cleanup(ctx))
		return ctx
	})

	testEnv.Test(t, f.Feature())
}

func TestSyncSetsCorrectLabels(t *testing.T) {
	var s applicationCatalogSuite
	f := features.New("SyncSetsCorrectLabels")

	f.Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		err := s.setupTestCase(ctx, cfg)
		require.NoError(t, err, "failed to setup test case")
		return ctx
	}).Assess("Controller should set managed-by and catalog-name labels",
		func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			catalogName := "test-labels-catalog"
			catalog := &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: catalogName,
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{
								ChartName: "labeled-app",
								ChartVersions: []catalogv1alpha1.ChartVersion{
									{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
								},
							},
						},
					},
				},
			}

			err := s.createApplicationCatalog(ctx, catalog)
			require.NoError(t, err, "failed to create ApplicationCatalog")

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDef, err := s.getApplicationDefinition(ctx, "labeled-app")
				if err != nil {
					return false, nil
				}

				labels := appDef.Labels
				if labels == nil {
					t.Log("labels not set yet")
					return false, nil
				}

				managedBy, ok := labels[catalogv1alpha1.LabelManagedByApplicationCatalog]
				if !ok || managedBy != "true" {
					t.Logf("expected managed-by label to be 'true', got %q", managedBy)
					return false, nil
				}

				catalogLabel, ok := labels[catalogv1alpha1.LabelApplicationCatalogName]
				if !ok || catalogLabel != catalogName {
					t.Logf("expected catalog-name label to be %q, got %q", catalogName, catalogLabel)
					return false, nil
				}

				return true, nil
			})
			require.NoError(t, err, "labels should be set correctly")

			t.Log("Labels set correctly")
			return ctx
		},
	).Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		require.NoError(t, s.cleanup(ctx))
		return ctx
	})

	testEnv.Test(t, f.Feature())
}

func TestSyncAppNameFromMetadata(t *testing.T) {
	var s applicationCatalogSuite
	f := features.New("SyncAppNameFromMetadata")

	f.Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		err := s.setupTestCase(ctx, cfg)
		require.NoError(t, err, "failed to setup test case")
		return ctx
	}).Assess("Controller should use metadata.appName for ApplicationDefinition name",
		func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			catalog := &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-appname",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{
								ChartName: "gpu-operator",
								Metadata: &catalogv1alpha1.ChartMetadata{
									AppName:     "nvidia-gpu-operator", // custom app name
									DisplayName: "NVIDIA GPU Operator",
								},
								ChartVersions: []catalogv1alpha1.ChartVersion{
									{ChartVersion: "v25.0.0", AppVersion: "v25.0.0"},
								},
							},
						},
					},
				},
			}

			err := s.createApplicationCatalog(ctx, catalog)
			require.NoError(t, err, "failed to create ApplicationCatalog")

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDef, err := s.getApplicationDefinition(ctx, "nvidia-gpu-operator")
				if err != nil {
					t.Log("waiting for nvidia-gpu-operator ApplicationDefinition")
					return false, nil
				}

				if appDef.Name != "nvidia-gpu-operator" {
					t.Logf("expected name nvidia-gpu-operator, got %s", appDef.Name)
					return false, nil
				}

				return true, nil
			})
			require.NoError(t, err, "should use metadata.appName for ApplicationDefinition name")

			_, err = s.getApplicationDefinition(ctx, "gpu-operator")
			require.True(t, apierrors.IsNotFound(err), "gpu-operator should not exist")

			t.Log("metadata.appName used correctly")
			return ctx
		},
	).Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		require.NoError(t, s.cleanup(ctx))
		return ctx
	})

	testEnv.Test(t, f.Feature())
}

func TestSyncMultipleVersions(t *testing.T) {
	var s applicationCatalogSuite
	f := features.New("SyncMultipleVersions")

	f.Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		err := s.setupTestCase(ctx, cfg)
		require.NoError(t, err, "failed to setup test case")
		return ctx
	}).Assess("Controller should include all versions in ApplicationDefinition",
		func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			catalog := &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-versions",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{
								ChartName: "multi-version-app",
								ChartVersions: []catalogv1alpha1.ChartVersion{
									{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
									{ChartVersion: "1.1.0", AppVersion: "v1.1.0"},
									{ChartVersion: "2.0.0", AppVersion: "v2.0.0"},
								},
							},
						},
					},
				},
			}

			err := s.createApplicationCatalog(ctx, catalog)
			require.NoError(t, err, "failed to create ApplicationCatalog")

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDef, err := s.getApplicationDefinition(ctx, "multi-version-app")
				if err != nil {
					return false, nil
				}

				if len(appDef.Spec.Versions) != 3 {
					t.Logf("expected 3 versions, got %d", len(appDef.Spec.Versions))
					return false, nil
				}

				return true, nil
			})
			require.NoError(t, err, "should have all versions")

			appDef, _ := s.getApplicationDefinition(ctx, "multi-version-app")
			versions := make(map[string]bool)
			for _, v := range appDef.Spec.Versions {
				versions[v.Version] = true
			}

			require.True(t, versions["v1.0.0"], "v1.0.0 should exist")
			require.True(t, versions["v1.1.0"], "v1.1.0 should exist")
			require.True(t, versions["v2.0.0"], "v2.0.0 should exist")

			t.Log("Multiple versions handled correctly")
			return ctx
		},
	).Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		require.NoError(t, s.cleanup(ctx))
		return ctx
	})

	testEnv.Test(t, f.Feature())
}

func TestURLResolution(t *testing.T) {
	testCases := []struct {
		name          string
		description   string
		catalogName   string
		appName       string
		globalURL     string
		chartURL      string
		chartVersions []catalogv1alpha1.ChartVersion
		expectedURLs  map[string]string // appVersion -> expected URL
	}{
		{
			name:        "default-only",
			description: "Should use default repository URL when none specified",
			catalogName: "test-url-default",
			appName:     "default-url-app",
			chartVersions: []catalogv1alpha1.ChartVersion{
				{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
			},
			expectedURLs: map[string]string{
				"v1.0.0": catalogv1alpha1.DefaultHelmRepository,
			},
		},
		{
			name:        "global-override",
			description: "Global repositorySettings should override default",
			catalogName: "test-url-global",
			appName:     "global-url-app",
			globalURL:   "oci://custom.registry.io/charts",
			chartVersions: []catalogv1alpha1.ChartVersion{
				{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
			},
			expectedURLs: map[string]string{
				"v1.0.0": "oci://custom.registry.io/charts",
			},
		},
		{
			name:        "chart-overrides-global",
			description: "Chart-level URL should override global URL",
			catalogName: "test-url-chart-override",
			appName:     "chart-override-app",
			globalURL:   "oci://global.registry.io/charts",
			chartURL:    "oci://chart.registry.io/special",
			chartVersions: []catalogv1alpha1.ChartVersion{
				{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
			},
			expectedURLs: map[string]string{
				"v1.0.0": "oci://chart.registry.io/special",
			},
		},
		{
			name:        "version-overrides-all",
			description: "Version-level URL should override chart and global URLs",
			catalogName: "test-url-version-override",
			appName:     "version-override-app",
			globalURL:   "oci://global.registry.io/charts",
			chartURL:    "oci://chart.registry.io/special",
			chartVersions: []catalogv1alpha1.ChartVersion{
				{
					ChartVersion: "1.0.0",
					AppVersion:   "v1.0.0",
					RepositorySettings: &catalogv1alpha1.RepositorySettings{
						BaseURL: "https://legacy.charts.io",
					},
				},
				{
					ChartVersion: "2.0.0",
					AppVersion:   "v2.0.0",
					// No override - should use chart URL
				},
			},
			expectedURLs: map[string]string{
				"v1.0.0": "https://legacy.charts.io",
				"v2.0.0": "oci://chart.registry.io/special",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var s applicationCatalogSuite
			f := features.New("URLResolution-" + tc.name)

			f.Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err := s.setupTestCase(ctx, cfg)
				require.NoError(t, err, "failed to setup test case")
				return ctx
			}).Assess(tc.description, func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				helmSpec := &catalogv1alpha1.HelmSpec{}

				if tc.globalURL != "" {
					helmSpec.RepositorySettings = &catalogv1alpha1.RepositorySettings{
						BaseURL: tc.globalURL,
					}
				}

				chartConfig := catalogv1alpha1.ChartConfig{
					ChartName:     tc.appName,
					ChartVersions: tc.chartVersions,
				}

				if tc.chartURL != "" {
					chartConfig.RepositorySettings = &catalogv1alpha1.RepositorySettings{
						BaseURL: tc.chartURL,
					}
				}

				helmSpec.Charts = []catalogv1alpha1.ChartConfig{chartConfig}

				catalog := &catalogv1alpha1.ApplicationCatalog{
					ObjectMeta: metav1.ObjectMeta{
						Name: tc.catalogName,
					},
					Spec: catalogv1alpha1.ApplicationCatalogSpec{
						Helm: helmSpec,
					},
				}

				err := s.createApplicationCatalog(ctx, catalog)
				require.NoError(t, err, "failed to create ApplicationCatalog")

				err = waitFor(ctx, func(ctx context.Context) (bool, error) {
					appDef, err := s.getApplicationDefinition(ctx, tc.appName)
					if err != nil {
						return false, nil
					}

					if len(appDef.Spec.Versions) < len(tc.expectedURLs) {
						return false, nil
					}

					actualURLs := make(map[string]string)
					for _, v := range appDef.Spec.Versions {
						actualURLs[v.Version] = v.Template.Source.Helm.URL
					}

					for appVersion, expectedURL := range tc.expectedURLs {
						if actualURLs[appVersion] != expectedURL {
							t.Logf("%s: expected %q, got %q", appVersion, expectedURL, actualURLs[appVersion])
							return false, nil
						}
					}

					return true, nil
				})
				require.NoError(t, err, "URL resolution failed")

				t.Logf("URL resolution test %q passed", tc.name)
				return ctx
			}).Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				require.NoError(t, s.cleanup(ctx))
				return ctx
			})

			testEnv.Test(t, f.Feature())
		})
	}
}

func TestDefaultsPreservesUserCustomization(t *testing.T) {
	var s applicationCatalogSuite
	f := features.New("DefaultsPreservesUserCustomization")

	f.Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		err := s.setupTestCase(ctx, cfg)
		require.NoError(t, err, "failed to setup test case")
		return ctx
	}).Assess("Controller should preserve user's defaultValuesBlock on update",
		func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			catalog := &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-preserve-values",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{
								ChartName:          "preserve-values-app",
								DefaultValuesBlock: "original: value",
								ChartVersions: []catalogv1alpha1.ChartVersion{
									{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
								},
							},
						},
					},
				},
			}

			err := s.createApplicationCatalog(ctx, catalog)
			require.NoError(t, err, "failed to create ApplicationCatalog")

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				_, err := s.getApplicationDefinition(ctx, "preserve-values-app")
				return err == nil, nil
			})
			require.NoError(t, err, "ApplicationDefinition should be created")

			userValues := "customized: by-user\nreplicas: 5"
			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDef, err := s.getApplicationDefinition(ctx, "preserve-values-app")
				if err != nil {
					return false, nil
				}

				appDef.Spec.DefaultValuesBlock = userValues
				if err := s.updateApplicationDefinition(ctx, appDef); err != nil {
					t.Logf("failed to update: %v", err)
					return false, nil
				}

				return true, nil
			})
			require.NoError(t, err, "should update defaultValuesBlock")

			catalogUpdated, _ := s.getApplicationCatalog(ctx, "test-preserve-values")
			catalogUpdated.Spec.Helm.Charts[0].Metadata = &catalogv1alpha1.ChartMetadata{
				DisplayName: "Updated Display Name",
			}
			err = s.updateApplicationCatalog(ctx, catalogUpdated)
			require.NoError(t, err, "should update catalog")

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDef, err := s.getApplicationDefinition(ctx, "preserve-values-app")
				if err != nil {
					return false, nil
				}

				if appDef.Spec.DisplayName != "Updated Display Name" {
					t.Log("waiting for spec update")
					return false, nil
				}

				if appDef.Spec.DefaultValuesBlock != userValues {
					t.Logf("expected user values %q, got %q", userValues, appDef.Spec.DefaultValuesBlock)
					return false, nil
				}

				return true, nil
			})
			require.NoError(t, err, "user's defaultValuesBlock should be preserved")

			t.Log("User customization preserved correctly")
			return ctx
		},
	).Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		require.NoError(t, s.cleanup(ctx))
		return ctx
	})

	testEnv.Test(t, f.Feature())
}

func TestDefaultsUpdatesIfEmpty(t *testing.T) {
	var s applicationCatalogSuite
	f := features.New("DefaultsUpdatesIfEmpty")

	f.Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		err := s.setupTestCase(ctx, cfg)
		require.NoError(t, err, "failed to setup test case")
		return ctx
	}).Assess("Controller can set defaultValuesBlock when empty",
		func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			catalog := &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-empty-values",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{
								ChartName: "empty-values-app",
								ChartVersions: []catalogv1alpha1.ChartVersion{
									{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
								},
							},
						},
					},
				},
			}

			err := s.createApplicationCatalog(ctx, catalog)
			require.NoError(t, err, "failed to create ApplicationCatalog")

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDef, err := s.getApplicationDefinition(ctx, "empty-values-app")
				if err != nil {
					return false, nil
				}

				if appDef.Spec.DefaultValuesBlock != "" {
					t.Logf("expected empty defaultValuesBlock, got %q", appDef.Spec.DefaultValuesBlock)
				}

				return true, nil
			})
			require.NoError(t, err, "ApplicationDefinition should be created")

			catalogUpdated, _ := s.getApplicationCatalog(ctx, "test-empty-values")
			catalogUpdated.Spec.Helm.Charts[0].DefaultValuesBlock = "new: value"
			err = s.updateApplicationCatalog(ctx, catalogUpdated)
			require.NoError(t, err, "should update catalog")

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDef, err := s.getApplicationDefinition(ctx, "empty-values-app")
				if err != nil {
					return false, nil
				}

				if appDef.Spec.DefaultValuesBlock != "new: value" {
					t.Log("waiting for defaultValuesBlock update")
					return false, nil
				}

				return true, nil
			})
			require.NoError(t, err, "defaultValuesBlock should be set")

			t.Log("Empty defaultValuesBlock updated correctly")
			return ctx
		},
	).Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		require.NoError(t, s.cleanup(ctx))
		return ctx
	})

	testEnv.Test(t, f.Feature())
}

func TestOrphanRemoveChartRemovesLabels(t *testing.T) {
	var s applicationCatalogSuite
	f := features.New("OrphanRemoveChartRemovesLabels")

	f.Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		err := s.setupTestCase(ctx, cfg)
		require.NoError(t, err, "failed to setup test case")
		return ctx
	}).Assess("Removing chart should remove managed labels but preserve AppDef",
		func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			catalogName := "test-orphan-labels"
			catalog := &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: catalogName,
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{
								ChartName: "orphan-app-one",
								ChartVersions: []catalogv1alpha1.ChartVersion{
									{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
								},
							},
							{
								ChartName: "orphan-app-two",
								ChartVersions: []catalogv1alpha1.ChartVersion{
									{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
								},
							},
						},
					},
				},
			}

			err := s.createApplicationCatalog(ctx, catalog)
			require.NoError(t, err, "failed to create ApplicationCatalog")

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDefs, err := s.listApplicationDefinitions(ctx)
				if err != nil {
					return false, nil
				}
				return len(appDefs.Items) >= 2, nil
			})
			require.NoError(t, err, "both ApplicationDefinitions should be created")

			catalogUpdated, _ := s.getApplicationCatalog(ctx, catalogName)
			catalogUpdated.Spec.Helm.Charts = []catalogv1alpha1.ChartConfig{
				{
					ChartName: "orphan-app-one", // Keep only app-one
					ChartVersions: []catalogv1alpha1.ChartVersion{
						{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
					},
				},
			}
			err = s.updateApplicationCatalog(ctx, catalogUpdated)
			require.NoError(t, err, "should update catalog")

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDef, err := s.getApplicationDefinition(ctx, "orphan-app-two")
				if err != nil {
					t.Logf("orphan-app-two not found: %v", err)
					return false, fmt.Errorf("ApplicationDefinition should still exist")
				}

				if appDef.Labels == nil {
					return true, nil
				}

				if _, exists := appDef.Labels[catalogv1alpha1.LabelManagedByApplicationCatalog]; exists {
					t.Log("waiting for managed-by label to be removed")
					return false, nil
				}

				if _, exists := appDef.Labels[catalogv1alpha1.LabelApplicationCatalogName]; exists {
					t.Log("waiting for catalog-name label to be removed")
					return false, nil
				}

				return true, nil
			})
			require.NoError(t, err, "managed labels should be removed")

			appOne, err := s.getApplicationDefinition(ctx, "orphan-app-one")
			require.NoError(t, err)
			require.Equal(t, "true", appOne.Labels[catalogv1alpha1.LabelManagedByApplicationCatalog])
			require.Equal(t, catalogName, appOne.Labels[catalogv1alpha1.LabelApplicationCatalogName])

			t.Log("Orphan management (label removal) works correctly")
			return ctx
		},
	).Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		require.NoError(t, s.cleanup(ctx))
		return ctx
	})

	testEnv.Test(t, f.Feature())
}

func TestOrphanDeleteCatalogUnmanagesAll(t *testing.T) {
	var s applicationCatalogSuite
	f := features.New("OrphanDeleteCatalogUnmanagesAll")

	f.Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		err := s.setupTestCase(ctx, cfg)
		require.NoError(t, err, "failed to setup test case")
		return ctx
	}).Assess("Deleting catalog should unmanage all ApplicationDefinitions",
		func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			catalogName := "test-delete-catalog"
			catalog := &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: catalogName,
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{
								ChartName: "delete-test-app",
								ChartVersions: []catalogv1alpha1.ChartVersion{
									{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
								},
							},
						},
					},
				},
			}

			err := s.createApplicationCatalog(ctx, catalog)
			require.NoError(t, err, "failed to create ApplicationCatalog")

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				_, err := s.getApplicationDefinition(ctx, "delete-test-app")
				return err == nil, nil
			})
			require.NoError(t, err, "ApplicationDefinition should be created")

			err = s.deleteApplicationCatalog(ctx, catalogName)
			require.NoError(t, err, "should delete catalog")

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDef, err := s.getApplicationDefinition(ctx, "delete-test-app")
				if apierrors.IsNotFound(err) {
					return false, fmt.Errorf("ApplicationDefinition should NOT be deleted")
				}
				if err != nil {
					return false, nil
				}

				if appDef.Labels == nil {
					return true, nil
				}

				if _, exists := appDef.Labels[catalogv1alpha1.LabelManagedByApplicationCatalog]; exists {
					t.Log("waiting for managed-by label to be removed")
					return false, nil
				}

				return true, nil
			})
			require.NoError(t, err, "ApplicationDefinition should be unmanaged (not deleted)")

			t.Log("Catalog deletion correctly unmanages ApplicationDefinitions")
			return ctx
		},
	).Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		require.NoError(t, s.cleanup(ctx))
		return ctx
	})

	testEnv.Test(t, f.Feature())
}

func TestOrphanReaddChartRemanagesAppDef(t *testing.T) {
	var s applicationCatalogSuite
	f := features.New("OrphanReaddChartRemanagesAppDef")

	f.Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		err := s.setupTestCase(ctx, cfg)
		require.NoError(t, err, "failed to setup test case")
		return ctx
	}).Assess("Re-adding chart should re-add managed labels",
		func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			catalogName := "test-readd-chart"
			catalog := &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: catalogName,
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{
								ChartName: "readd-app",
								ChartVersions: []catalogv1alpha1.ChartVersion{
									{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
								},
							},
						},
					},
				},
			}

			err := s.createApplicationCatalog(ctx, catalog)
			require.NoError(t, err, "failed to create ApplicationCatalog")

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDef, err := s.getApplicationDefinition(ctx, "readd-app")
				if err != nil {
					return false, nil
				}
				return appDef.Labels[catalogv1alpha1.LabelManagedByApplicationCatalog] == "true", nil
			})
			require.NoError(t, err, "ApplicationDefinition should be managed")

			catalogUpdated, _ := s.getApplicationCatalog(ctx, catalogName)
			catalogUpdated.Spec.Helm.Charts = []catalogv1alpha1.ChartConfig{} // Empty
			err = s.updateApplicationCatalog(ctx, catalogUpdated)
			require.NoError(t, err, "should update catalog to empty")

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDef, err := s.getApplicationDefinition(ctx, "readd-app")
				if err != nil {
					return false, nil
				}
				if appDef.Labels == nil {
					return true, nil
				}
				_, exists := appDef.Labels[catalogv1alpha1.LabelManagedByApplicationCatalog]
				return !exists, nil
			})
			require.NoError(t, err, "labels should be removed")

			catalogUpdated, _ = s.getApplicationCatalog(ctx, catalogName)
			catalogUpdated.Spec.Helm.Charts = []catalogv1alpha1.ChartConfig{
				{
					ChartName: "readd-app",
					ChartVersions: []catalogv1alpha1.ChartVersion{
						{ChartVersion: "2.0.0", AppVersion: "v2.0.0"}, // New version
					},
				},
			}
			err = s.updateApplicationCatalog(ctx, catalogUpdated)
			require.NoError(t, err, "should re-add chart")

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDef, err := s.getApplicationDefinition(ctx, "readd-app")
				if err != nil {
					return false, nil
				}

				if appDef.Labels == nil {
					return false, nil
				}

				managedBy, exists := appDef.Labels[catalogv1alpha1.LabelManagedByApplicationCatalog]
				if !exists || managedBy != "true" {
					t.Log("waiting for managed-by label to be re-added")
					return false, nil
				}

				catalogLabel, exists := appDef.Labels[catalogv1alpha1.LabelApplicationCatalogName]
				if !exists || catalogLabel != catalogName {
					t.Log("waiting for catalog-name label to be re-added")
					return false, nil
				}

				return true, nil
			})
			require.NoError(t, err, "labels should be re-added")

			t.Log("Re-adoption (re-managing orphaned AppDef) works correctly")
			return ctx
		},
	).Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		require.NoError(t, s.cleanup(ctx))
		return ctx
	})

	testEnv.Test(t, f.Feature())
}

func TestMultiCatalogSeparateAppDefs(t *testing.T) {
	var s applicationCatalogSuite
	f := features.New("MultiCatalogSeparateAppDefs")

	f.Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		err := s.setupTestCase(ctx, cfg)
		require.NoError(t, err, "failed to setup test case")
		return ctx
	}).Assess("Multiple catalogs should create separate ApplicationDefinitions",
		func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			catalog1 := &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "catalog-one",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{
								ChartName: "app-from-catalog-one",
								ChartVersions: []catalogv1alpha1.ChartVersion{
									{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
								},
							},
						},
					},
				},
			}

			catalog2 := &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "catalog-two",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{
								ChartName: "app-from-catalog-two",
								ChartVersions: []catalogv1alpha1.ChartVersion{
									{ChartVersion: "2.0.0", AppVersion: "v2.0.0"},
								},
							},
						},
					},
				},
			}

			err := s.createApplicationCatalog(ctx, catalog1)
			require.NoError(t, err, "failed to create catalog-one")

			err = s.createApplicationCatalog(ctx, catalog2)
			require.NoError(t, err, "failed to create catalog-two")

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDefs, err := s.listApplicationDefinitions(ctx)
				if err != nil {
					return false, nil
				}

				if len(appDefs.Items) < 2 {
					t.Logf("waiting for 2 apps, got %d", len(appDefs.Items))
					return false, nil
				}

				foundApps := make(map[string]string) // name -> catalog
				for _, app := range appDefs.Items {
					catalog := app.Labels[catalogv1alpha1.LabelApplicationCatalogName]
					foundApps[app.Name] = catalog
				}

				if foundApps["app-from-catalog-one"] != "catalog-one" {
					return false, nil
				}
				if foundApps["app-from-catalog-two"] != "catalog-two" {
					return false, nil
				}

				return true, nil
			})
			require.NoError(t, err, "both catalogs should create their own ApplicationDefinitions")

			t.Log("Multiple catalogs create separate ApplicationDefinitions correctly")
			return ctx
		},
	).Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		require.NoError(t, s.cleanup(ctx))
		return ctx
	})

	testEnv.Test(t, f.Feature())
}

func TestMultiCatalogDeleteOneKeepsOther(t *testing.T) {
	var s applicationCatalogSuite
	f := features.New("MultiCatalogDeleteOneKeepsOther")

	f.Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		err := s.setupTestCase(ctx, cfg)
		require.NoError(t, err, "failed to setup test case")
		return ctx
	}).Assess("Deleting one catalog should not affect other catalog's apps",
		func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			catalog1 := &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "delete-test-one",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{
								ChartName: "app-delete-one",
								ChartVersions: []catalogv1alpha1.ChartVersion{
									{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
								},
							},
						},
					},
				},
			}

			catalog2 := &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "delete-test-two",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{
								ChartName: "app-delete-two",
								ChartVersions: []catalogv1alpha1.ChartVersion{
									{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
								},
							},
						},
					},
				},
			}

			err := s.createApplicationCatalog(ctx, catalog1)
			require.NoError(t, err)

			err = s.createApplicationCatalog(ctx, catalog2)
			require.NoError(t, err)

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDefs, _ := s.listApplicationDefinitions(ctx)
				return len(appDefs.Items) >= 2, nil
			})
			require.NoError(t, err)

			err = s.deleteApplicationCatalog(ctx, "delete-test-one")
			require.NoError(t, err)

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDef, err := s.getApplicationDefinition(ctx, "app-delete-two")
				if err != nil {
					return false, nil
				}

				managedBy := appDef.Labels[catalogv1alpha1.LabelManagedByApplicationCatalog]
				catalogName := appDef.Labels[catalogv1alpha1.LabelApplicationCatalogName]

				if managedBy != "true" || catalogName != "delete-test-two" {
					t.Log("catalog-two's app should still be managed")
					return false, nil
				}

				return true, nil
			})
			require.NoError(t, err, "other catalog's apps should be unaffected")

			t.Log("Deleting one catalog doesn't affect other catalog's apps")
			return ctx
		},
	).Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		require.NoError(t, s.cleanup(ctx))
		return ctx
	})

	testEnv.Test(t, f.Feature())
}

func TestCredsGlobalCredentials(t *testing.T) {
	var s applicationCatalogSuite
	f := features.New("CredsGlobalCredentials")

	f.Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		err := s.setupTestCase(ctx, cfg)
		require.NoError(t, err, "failed to setup test case")

		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-registry-creds",
				Namespace: "kubermatic",
			},
			Data: map[string][]byte{
				"username": []byte("test-user"),
				"password": []byte("test-pass"),
			},
		}
		_ = s.client.Create(ctx, secret)

		return ctx
	}).Assess("Global credentials should be used when global URL is set",
		func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			catalog := &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-global-creds",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						RepositorySettings: &catalogv1alpha1.RepositorySettings{
							BaseURL: "oci://private.registry.io/charts",
							Credentials: &catalogv1alpha1.RepositoryCredentials{
								Username: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "test-registry-creds",
									},
									Key: "username",
								},
								Password: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "test-registry-creds",
									},
									Key: "password",
								},
							},
						},
						Charts: []catalogv1alpha1.ChartConfig{
							{
								ChartName: "creds-app",
								ChartVersions: []catalogv1alpha1.ChartVersion{
									{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
								},
							},
						},
					},
				},
			}

			err := s.createApplicationCatalog(ctx, catalog)
			require.NoError(t, err, "failed to create ApplicationCatalog")

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDef, err := s.getApplicationDefinition(ctx, "creds-app")
				if err != nil {
					return false, nil
				}

				if len(appDef.Spec.Versions) == 0 {
					return false, nil
				}

				creds := appDef.Spec.Versions[0].Template.Source.Helm.Credentials
				if creds == nil {
					t.Log("waiting for credentials to be set")
					return false, nil
				}

				if creds.Username == nil || creds.Username.Name != "test-registry-creds" {
					t.Log("username credential not set correctly")
					return false, nil
				}

				if creds.Password == nil || creds.Password.Name != "test-registry-creds" {
					t.Log("password credential not set correctly")
					return false, nil
				}

				return true, nil
			})
			require.NoError(t, err, "global credentials should be used")

			t.Log("Global credentials resolved correctly")
			return ctx
		},
	).Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		require.NoError(t, s.cleanup(ctx))
		return ctx
	})

	testEnv.Test(t, f.Feature())
}

func TestCredsNoCredentialsWithDefaultURL(t *testing.T) {
	var s applicationCatalogSuite
	f := features.New("CredsNoCredentialsWithDefaultURL")

	f.Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		err := s.setupTestCase(ctx, cfg)
		require.NoError(t, err, "failed to setup test case")
		return ctx
	}).Assess("No credentials when using default URL",
		func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			catalog := &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-no-creds",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{
								ChartName: "no-creds-app",
								ChartVersions: []catalogv1alpha1.ChartVersion{
									{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
								},
							},
						},
					},
				},
			}

			err := s.createApplicationCatalog(ctx, catalog)
			require.NoError(t, err, "failed to create ApplicationCatalog")

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDef, err := s.getApplicationDefinition(ctx, "no-creds-app")
				if err != nil {
					return false, nil
				}

				if len(appDef.Spec.Versions) == 0 {
					return false, nil
				}

				creds := appDef.Spec.Versions[0].Template.Source.Helm.Credentials
				if creds != nil {
					t.Logf("expected no credentials, got %+v", creds)
					return false, nil
				}

				return true, nil
			})
			require.NoError(t, err, "should have no credentials with default URL")

			t.Log("No credentials with default URL - correct")
			return ctx
		},
	).Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		require.NoError(t, s.cleanup(ctx))
		return ctx
	})

	testEnv.Test(t, f.Feature())
}

func TestIntegrationFullLifecycle(t *testing.T) {
	var s applicationCatalogSuite
	f := features.New("IntegrationFullLifecycle")

	f.Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		err := s.setupTestCase(ctx, cfg)
		require.NoError(t, err, "failed to setup test case")
		return ctx
	}).Assess("Full lifecycle: create, update, customize, orphan, delete",
		func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			catalogName := "lifecycle-test"

			t.Log("Step 1: Creating catalog with 2 charts")
			catalog := &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: catalogName,
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{
								ChartName:          "lifecycle-app-one",
								DefaultValuesBlock: "replicas: 1",
								ChartVersions: []catalogv1alpha1.ChartVersion{
									{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
								},
							},
							{
								ChartName: "lifecycle-app-two",
								ChartVersions: []catalogv1alpha1.ChartVersion{
									{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
								},
							},
						},
					},
				},
			}

			err := s.createApplicationCatalog(ctx, catalog)
			require.NoError(t, err)

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDefs, _ := s.listApplicationDefinitions(ctx)
				return len(appDefs.Items) >= 2, nil
			})
			require.NoError(t, err, "2 ApplicationDefinitions should be created")
			t.Log("Step 1 complete: 2 ApplicationDefinitions created")

			t.Log("Step 2: Adding third chart")
			catalogUpdated, _ := s.getApplicationCatalog(ctx, catalogName)
			catalogUpdated.Spec.Helm.Charts = append(catalogUpdated.Spec.Helm.Charts, catalogv1alpha1.ChartConfig{
				ChartName: "lifecycle-app-three",
				ChartVersions: []catalogv1alpha1.ChartVersion{
					{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
				},
			})
			err = s.updateApplicationCatalog(ctx, catalogUpdated)
			require.NoError(t, err)

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDefs, _ := s.listApplicationDefinitions(ctx)
				return len(appDefs.Items) == 3, nil
			})
			require.NoError(t, err, "3 ApplicationDefinitions should exist")
			t.Log("Step 2 complete: 3 ApplicationDefinitions exist")

			t.Log("Step 3: User customizing defaultValuesBlock")
			userValues := "customized: true\nreplicas: 5"
			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDef, err := s.getApplicationDefinition(ctx, "lifecycle-app-one")
				if err != nil {
					return false, nil
				}
				appDef.Spec.DefaultValuesBlock = userValues
				return s.updateApplicationDefinition(ctx, appDef) == nil, nil
			})
			require.NoError(t, err)
			t.Log("Step 3 complete: User customization applied")

			t.Log("Step 4: Removing chart (orphan)")
			catalogUpdated, _ = s.getApplicationCatalog(ctx, catalogName)
			catalogUpdated.Spec.Helm.Charts = catalogUpdated.Spec.Helm.Charts[:2]
			err = s.updateApplicationCatalog(ctx, catalogUpdated)
			require.NoError(t, err)

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDef, err := s.getApplicationDefinition(ctx, "lifecycle-app-three")
				if err != nil {
					return false, nil
				}
				if appDef.Labels == nil {
					return true, nil
				}
				_, exists := appDef.Labels[catalogv1alpha1.LabelManagedByApplicationCatalog]
				return !exists, nil
			})
			require.NoError(t, err, "lifecycle-app-three should be orphaned")
			t.Log("Step 4 complete: Third app orphaned (labels removed)")

			t.Log("Step 5: Verifying user customization preserved")
			appDef, err := s.getApplicationDefinition(ctx, "lifecycle-app-one")
			require.NoError(t, err)
			require.Equal(t, userValues, appDef.Spec.DefaultValuesBlock, "user values should be preserved")
			t.Log("Step 5 complete: User customization preserved")

			t.Log("Step 6: Deleting catalog")
			err = s.deleteApplicationCatalog(ctx, catalogName)
			require.NoError(t, err)

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDef, err := s.getApplicationDefinition(ctx, "lifecycle-app-one")
				if apierrors.IsNotFound(err) {
					return false, fmt.Errorf("AppDef should NOT be deleted")
				}
				if err != nil {
					return false, nil
				}
				if appDef.Labels == nil {
					return true, nil
				}
				_, exists := appDef.Labels[catalogv1alpha1.LabelManagedByApplicationCatalog]
				return !exists, nil
			})
			require.NoError(t, err, "All apps should be unmanaged")
			t.Log("Step 6 complete: All apps unmanaged (not deleted)")

			t.Log("Full lifecycle test completed successfully!")
			return ctx
		},
	).Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		require.NoError(t, s.cleanup(ctx))
		return ctx
	})

	testEnv.Test(t, f.Feature())
}

func TestValidationWebhookRejectsConflictingCatalog(t *testing.T) {
	var s applicationCatalogSuite
	f := features.New("ValidationWebhookRejectsConflictingCatalog")

	f.Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		err := s.setupTestCase(ctx, cfg)
		require.NoError(t, err, "failed to setup test case")
		return ctx
	}).Assess("Webhook should reject catalog that conflicts with existing ApplicationDefinition",
		func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			catalog1 := &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "conflict-catalog-one",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{
								ChartName: "conflict-nginx",
								ChartVersions: []catalogv1alpha1.ChartVersion{
									{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
								},
							},
						},
					},
				},
			}

			err := s.createApplicationCatalog(ctx, catalog1)
			require.NoError(t, err, "failed to create catalog-one")

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDef, err := s.getApplicationDefinition(ctx, "conflict-nginx")
				if err != nil {
					return false, nil
				}
				return appDef.Labels[catalogv1alpha1.LabelManagedByApplicationCatalog] == "true" &&
					appDef.Labels[catalogv1alpha1.LabelApplicationCatalogName] == "conflict-catalog-one", nil
			})
			require.NoError(t, err, "conflict-nginx ApplicationDefinition should be created")
			t.Log("Step 1: catalog-one created with conflict-nginx ApplicationDefinition")

			catalog2 := &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "conflict-catalog-two",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{
								ChartName: "conflict-nginx", // Same app name - conflict!
								ChartVersions: []catalogv1alpha1.ChartVersion{
									{ChartVersion: "2.0.0", AppVersion: "v2.0.0"},
								},
							},
						},
					},
				},
			}

			err = s.client.Create(ctx, catalog2)
			require.Error(t, err, "creating conflicting catalog should be rejected by webhook")
			t.Logf("Step 2: Conflicting catalog correctly rejected: %v", err)

			appDef, err := s.getApplicationDefinition(ctx, "conflict-nginx")
			require.NoError(t, err, "original ApplicationDefinition should still exist")
			require.Equal(t, "conflict-catalog-one", appDef.Labels[catalogv1alpha1.LabelApplicationCatalogName],
				"ApplicationDefinition should still be owned by catalog-one")
			t.Log("Step 3: Original ApplicationDefinition unchanged")

			return ctx
		},
	).Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		require.NoError(t, s.cleanup(ctx))
		return ctx
	})

	testEnv.Test(t, f.Feature())
}

func TestValidationWebhookAllowsSameCatalogUpdate(t *testing.T) {
	var s applicationCatalogSuite
	f := features.New("ValidationWebhookAllowsSameCatalogUpdate")

	f.Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		err := s.setupTestCase(ctx, cfg)
		require.NoError(t, err, "failed to setup test case")
		return ctx
	}).Assess("Webhook should allow updating a catalog with its own apps",
		func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			catalog := &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "self-update-catalog",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{
								ChartName: "self-update-app",
								ChartVersions: []catalogv1alpha1.ChartVersion{
									{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
								},
							},
						},
					},
				},
			}

			err := s.createApplicationCatalog(ctx, catalog)
			require.NoError(t, err, "failed to create catalog")

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				_, err := s.getApplicationDefinition(ctx, "self-update-app")
				return err == nil, nil
			})
			require.NoError(t, err, "ApplicationDefinition should be created")

			updated, err := s.getApplicationCatalog(ctx, "self-update-catalog")
			require.NoError(t, err)

			updated.Spec.Helm.Charts[0].ChartVersions = append(
				updated.Spec.Helm.Charts[0].ChartVersions,
				catalogv1alpha1.ChartVersion{ChartVersion: "2.0.0", AppVersion: "v2.0.0"},
			)

			err = s.updateApplicationCatalog(ctx, updated)
			require.NoError(t, err, "updating same catalog should be allowed")

			t.Log("Catalog self-update allowed successfully")
			return ctx
		},
	).Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		require.NoError(t, s.cleanup(ctx))
		return ctx
	})

	testEnv.Test(t, f.Feature())
}

func TestValidationWebhookAllowsTakeoverOfUnmanagedAppDef(t *testing.T) {
	var s applicationCatalogSuite
	f := features.New("ValidationWebhookAllowsTakeoverOfUnmanagedAppDef")

	f.Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		err := s.setupTestCase(ctx, cfg)
		require.NoError(t, err, "failed to setup test case")
		return ctx
	}).Assess("Webhook should allow catalog to take over unmanaged ApplicationDefinition",
		func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			unmanagedAppDef := &appskubermaticv1.ApplicationDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: "unmanaged-takeover-app",
				},
				Spec: appskubermaticv1.ApplicationDefinitionSpec{
					Method: appskubermaticv1.HelmTemplateMethod,
					Versions: []appskubermaticv1.ApplicationVersion{
						{
							Version: "v0.0.1",
							Template: appskubermaticv1.ApplicationTemplate{
								Source: appskubermaticv1.ApplicationSource{
									Helm: &appskubermaticv1.HelmSource{
										URL:          "https://example.com/charts",
										ChartName:    "unmanaged-takeover-app",
										ChartVersion: "0.0.1",
									},
								},
							},
						},
					},
				},
			}

			err := s.client.Create(ctx, unmanagedAppDef)
			require.NoError(t, err, "failed to create unmanaged ApplicationDefinition")
			t.Log("Step 1: Created unmanaged ApplicationDefinition")

			catalog := &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "takeover-catalog",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{
								ChartName: "unmanaged-takeover-app",
								ChartVersions: []catalogv1alpha1.ChartVersion{
									{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
								},
							},
						},
					},
				},
			}

			err = s.createApplicationCatalog(ctx, catalog)
			require.NoError(t, err, "should be able to create catalog for unmanaged AppDef")
			t.Log("Step 2: Catalog created successfully - takeover allowed")

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDef, err := s.getApplicationDefinition(ctx, "unmanaged-takeover-app")
				if err != nil {
					return false, nil
				}
				return appDef.Labels[catalogv1alpha1.LabelApplicationCatalogName] == "takeover-catalog", nil
			})
			require.NoError(t, err, "AppDef should now be managed by the catalog")
			t.Log("Step 3: ApplicationDefinition is now managed by takeover-catalog")

			return ctx
		},
	).Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		require.NoError(t, s.cleanup(ctx))
		return ctx
	})

	testEnv.Test(t, f.Feature())
}

func TestValidationWebhookAllowsAfterOriginalCatalogDeleted(t *testing.T) {
	var s applicationCatalogSuite
	f := features.New("ValidationWebhookAllowsAfterOriginalCatalogDeleted")

	f.Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		err := s.setupTestCase(ctx, cfg)
		require.NoError(t, err, "failed to setup test case")
		return ctx
	}).Assess("After original catalog is deleted, another catalog should be able to take over",
		func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			catalog1 := &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "original-catalog",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{
								ChartName: "handover-app",
								ChartVersions: []catalogv1alpha1.ChartVersion{
									{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
								},
							},
						},
					},
				},
			}

			err := s.createApplicationCatalog(ctx, catalog1)
			require.NoError(t, err, "failed to create original catalog")

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDef, err := s.getApplicationDefinition(ctx, "handover-app")
				if err != nil {
					return false, nil
				}
				return appDef.Labels[catalogv1alpha1.LabelApplicationCatalogName] == "original-catalog", nil
			})
			require.NoError(t, err)
			t.Log("Step 1: Original catalog created")

			err = s.deleteApplicationCatalog(ctx, "original-catalog")
			require.NoError(t, err)

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDef, err := s.getApplicationDefinition(ctx, "handover-app")
				if err != nil {
					return false, nil
				}
				if appDef.Labels == nil {
					return true, nil
				}
				_, exists := appDef.Labels[catalogv1alpha1.LabelApplicationCatalogName]
				return !exists, nil
			})
			require.NoError(t, err, "AppDef should be unmanaged after catalog deletion")
			t.Log("Step 2: Original catalog deleted, AppDef unmanaged")

			catalog2 := &catalogv1alpha1.ApplicationCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name: "successor-catalog",
				},
				Spec: catalogv1alpha1.ApplicationCatalogSpec{
					Helm: &catalogv1alpha1.HelmSpec{
						Charts: []catalogv1alpha1.ChartConfig{
							{
								ChartName: "handover-app",
								ChartVersions: []catalogv1alpha1.ChartVersion{
									{ChartVersion: "2.0.0", AppVersion: "v2.0.0"},
								},
							},
						},
					},
				},
			}

			err = s.createApplicationCatalog(ctx, catalog2)
			require.NoError(t, err, "should be able to create successor catalog")
			t.Log("Step 3: Successor catalog created successfully")

			err = waitFor(ctx, func(ctx context.Context) (bool, error) {
				appDef, err := s.getApplicationDefinition(ctx, "handover-app")
				if err != nil {
					return false, nil
				}
				return appDef.Labels[catalogv1alpha1.LabelApplicationCatalogName] == "successor-catalog", nil
			})
			require.NoError(t, err, "AppDef should now be managed by successor catalog")

			return ctx
		},
	).Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		require.NoError(t, s.cleanup(ctx))
		return ctx
	})

	testEnv.Test(t, f.Feature())
}
