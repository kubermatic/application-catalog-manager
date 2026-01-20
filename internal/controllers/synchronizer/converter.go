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
	catalogv1alpha1 "k8c.io/application-catalog-manager/pkg/apis/applicationcatalog/v1alpha1"
	appskubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/apps.kubermatic/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// convertChartToApplicationDefinition converts a ChartConfig from an ApplicationCatalog
// into an ApplicationDefinition. This creates a new ApplicationDefinition with all
// fields populated from the catalog.
// The caller is responsible for preserving user customizations (like defaultValuesBlock)
// when updating existing resources.
func convertChartToApplicationDefinition(
	catalog *catalogv1alpha1.ApplicationCatalog,
	chart *catalogv1alpha1.ChartConfig,
) *appskubermaticv1.ApplicationDefinition {
	appName := chart.GetAppName()

	appDef := &appskubermaticv1.ApplicationDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: appName,
			Labels: map[string]string{
				catalogv1alpha1.LabelManagedByApplicationCatalog: "true",
				catalogv1alpha1.LabelApplicationCatalogName:      catalog.Name,
			},
		},
		Spec: appskubermaticv1.ApplicationDefinitionSpec{
			Method:             appskubermaticv1.HelmTemplateMethod,
			DefaultValuesBlock: chart.DefaultValuesBlock,
			Versions:           convertVersions(catalog, chart),
		},
	}

	if chart.Metadata != nil {
		appDef.Spec.DisplayName = chart.Metadata.DisplayName
		appDef.Spec.Description = chart.Metadata.Description
		appDef.Spec.DocumentationURL = chart.Metadata.DocumentationURL
		appDef.Spec.SourceURL = chart.Metadata.SourceURL
		appDef.Spec.Logo = chart.Metadata.Logo
		appDef.Spec.LogoFormat = chart.Metadata.LogoFormat
	}

	return appDef
}

// convertVersions converts ChartVersions from a ChartConfig into ApplicationVersions.
// It resolves the repository URL for each version using the catalog's precedence rules:
// version-level > chart-level > global > default.
func convertVersions(
	catalog *catalogv1alpha1.ApplicationCatalog,
	chart *catalogv1alpha1.ChartConfig,
) []appskubermaticv1.ApplicationVersion {
	versions := make([]appskubermaticv1.ApplicationVersion, 0, len(chart.ChartVersions))

	for i := range chart.ChartVersions {
		chartVersion := &chart.ChartVersions[i]

		url := catalog.ResolveChartURL(chart, chartVersion)

		version := appskubermaticv1.ApplicationVersion{
			Version: chartVersion.AppVersion,
			Template: appskubermaticv1.ApplicationTemplate{
				Source: appskubermaticv1.ApplicationSource{
					Helm: &appskubermaticv1.HelmSource{
						ChartName:    chart.ChartName,
						ChartVersion: chartVersion.ChartVersion,
						URL:          url,
					},
				},
			},
		}

		creds := catalog.ResolveChartCredentials(chart, chartVersion)
		if creds != nil {
			version.Template.Source.Helm.Credentials = convertCredentials(creds)
		}

		versions = append(versions, version)
	}

	return versions
}

// convertCredentials converts RepositoryCredentials from ApplicationCatalog format
// to HelmCredentials format used by ApplicationDefinition.
func convertCredentials(creds *catalogv1alpha1.RepositoryCredentials) *appskubermaticv1.HelmCredentials {
	if creds == nil {
		return nil
	}

	return &appskubermaticv1.HelmCredentials{
		Username:           creds.Username,
		Password:           creds.Password,
		RegistryConfigFile: creds.RegistryConfigFile,
	}
}
