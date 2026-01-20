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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HelmSpec defines the Helm-specific configuration for the application catalog.
type HelmSpec struct {
	// RepositorySettings defines the default repository settings for all charts.
	// Individual charts can override these settings.
	// By default, the controller will use the `DefaultHelmRepository`.
	//
	// +optional
	RepositorySettings *RepositorySettings `json:"repositorySettings,omitempty"`

	// Charts is the list of Helm charts to include in this catalog.
	// Each chart will be converted to an ApplicationDefinition.
	//
	// If nil/omitted, the webhook will inject default charts.
	// If empty (`[]`), no ApplicationDefinitions will be created.
	// If specified, only the listed charts will be created.
	//
	// +optional
	Charts []ChartConfig `json:"charts"`
}

// ApplicationCatalogSpec defines the desired state of ApplicationCatalog.
type ApplicationCatalogSpec struct {
	// Helm contains Helm chart configuration for this catalog.
	Helm *HelmSpec `json:"helm,omitempty"`
}

// ApplicationCatalogStatus defines the observed state of ApplicationCatalog.
type ApplicationCatalogStatus struct {
	// ObservedGeneration is the most recent generation observed by the controller.
	//
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=appcat
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name="Age",type="date"

// ApplicationCatalog is the Schema for the applicationcatalogs API.
// It defines a collection of Helm charts that will be converted to ApplicationDefinitions.
type ApplicationCatalog struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationCatalogSpec   `json:"spec,omitempty"`
	Status ApplicationCatalogStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApplicationCatalogList contains a list of ApplicationCatalog.
type ApplicationCatalogList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ApplicationCatalog `json:"items"`
}

// GetHelmCharts returns the list of Helm charts, or nil if not configured.
func (ac *ApplicationCatalog) GetHelmCharts() []ChartConfig {
	if ac.Spec.Helm == nil {
		return nil
	}
	return ac.Spec.Helm.Charts
}

// GetGlobalRepositorySettings returns the global repository settings, or nil if not configured.
func (ac *ApplicationCatalog) GetGlobalRepositorySettings() *RepositorySettings {
	if ac.Spec.Helm == nil {
		return nil
	}
	return ac.Spec.Helm.RepositorySettings
}

// ResolveChartURL resolves the repository URL for a specific chart version.
// It follows the precedence: version-level > chart-level > global > default.
func (ac *ApplicationCatalog) ResolveChartURL(chart *ChartConfig, version *ChartVersion) string {
	if version.RepositorySettings != nil && version.RepositorySettings.BaseURL != "" {
		return version.RepositorySettings.BaseURL
	}

	if chart.RepositorySettings != nil && chart.RepositorySettings.BaseURL != "" {
		return chart.RepositorySettings.BaseURL
	}

	if global := ac.GetGlobalRepositorySettings(); global != nil && global.BaseURL != "" {
		return global.BaseURL
	}

	return DefaultHelmRepository
}

// ResolveChartCredentials resolves the credentials for a specific chart version.
// Credentials are only returned if a baseURL is specified at the same level.
func (ac *ApplicationCatalog) ResolveChartCredentials(chart *ChartConfig, version *ChartVersion) *RepositoryCredentials {
	if version.RepositorySettings != nil && version.RepositorySettings.BaseURL != "" {
		return version.RepositorySettings.Credentials
	}

	if chart.RepositorySettings != nil && chart.RepositorySettings.BaseURL != "" {
		return chart.RepositorySettings.Credentials
	}

	if global := ac.GetGlobalRepositorySettings(); global != nil && global.BaseURL != "" {
		return global.Credentials
	}

	return nil
}
