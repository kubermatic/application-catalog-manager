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

// Package charts provides public access to the default application catalog
// charts used by the Application Catalog Manager.
//
// This package allows external consumers (like KKP's mirror-images command)
// to access the default Helm charts without requiring cluster access or
// Kubernetes client initialization.
package charts

import (
	"k8c.io/application-catalog-manager/internal/pkg/defaulting"
	catalogv1alpha1 "k8c.io/application-catalog-manager/pkg/apis/applicationcatalog/v1alpha1"
)

// GetDefaultCharts returns the default set of Helm charts that are
// included in every ApplicationCatalog when spec.helm.includeDefaults
// is true and no custom charts are specified.
//
// The returned charts are sorted by name and include all available
// versions for each chart. The chart definitions include:
// - Chart name and metadata (displayName, description, documentation URLs, logo)
// - Available chart versions with AppVersion mapping
// - Default Helm values for each chart
// - Repository settings (if overridden from default)
//
// This function does not require any cluster access or Kubernetes
// client initialization, making it suitable for offline tooling.
//
// Returns: A slice of ChartConfig structs representing all default charts.
func GetDefaultCharts() []catalogv1alpha1.ChartConfig {
	return defaulting.GetDefaultCharts()
}

// DefaultApplicationCatalog is a public re-export of the internal
// defaulting function. This allows external consumers to apply
// default charts to an ApplicationCatalog struct without requiring
// cluster access.
func DefaultApplicationCatalog(catalog *catalogv1alpha1.ApplicationCatalog) {
	defaulting.DefaultApplicationCatalog(catalog)
}
