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
	corev1 "k8s.io/api/core/v1"
)

// RepositorySettings defines the connection settings for a Helm chart repository.
type RepositorySettings struct {
	// BaseURL is the base URL of the Helm chart repository.
	// Supports http, https, and oci schemes.
	// Examples:
	//   - oci://quay.io/kubermatic-mirror/helm-charts
	//   - https://charts.example.com
	BaseURL string `json:"baseURL,omitempty"`

	// Credentials contains authentication information for the repository.
	// Only used when BaseURL is specified.
	//
	// +optional
	Credentials *RepositoryCredentials `json:"credentials,omitempty"`
}

// RepositoryCredentials defines authentication credentials for a Helm repository.
type RepositoryCredentials struct {
	// Username is a reference to a secret key containing the username.
	//
	// +optional
	Username *corev1.SecretKeySelector `json:"username,omitempty"`

	// Password is a reference to a secret key containing the password.
	//
	// +optional
	Password *corev1.SecretKeySelector `json:"password,omitempty"`

	// RegistryConfigFile is a reference to a secret key containing
	// a Docker config.json for OCI registry authentication.
	//
	// +optional
	RegistryConfigFile *corev1.SecretKeySelector `json:"registryConfigFile,omitempty"`
}

// ChartMetadata contains display information for an application.
type ChartMetadata struct {
	// AppName is the name used for the ApplicationDefinition metadata.name.
	// If not specified, chartName is used.
	//
	// +optional
	AppName string `json:"appName,omitempty"`

	// DisplayName is a human-readable name for the application.
	// +kubebuilder:validation:MinLength=1
	DisplayName string `json:"displayName"`

	// Description provides a brief description of the application.
	//
	// +optional
	Description string `json:"description,omitempty"`

	// DocumentationURL is a link to the application's documentation.
	//
	// +optional
	DocumentationURL string `json:"documentationURL,omitempty"`

	// SourceURL is a link to the application's source code repository.
	//
	// +optional
	SourceURL string `json:"sourceURL,omitempty"`

	// Logo is a base64-encoded image for the application logo.
	//
	// +optional
	Logo string `json:"logo,omitempty"`

	// LogoFormat specifies the format of the logo image.
	// +kubebuilder:validation:Enum=svg+xml;png
	LogoFormat string `json:"logoFormat,omitempty"`
}

// ChartVersion defines a specific version of a Helm chart.
type ChartVersion struct {
	// ChartVersion is the semantic version of the Helm chart (e.g., "4.7.1", "v1.16.0").
	// This corresponds to the chart version in Chart.yaml.
	//
	// +kubebuilder:validation:MinLength=1
	ChartVersion string `json:"chartVersion"`

	// AppVersion is the version of the application contained in the chart.
	// This maps to ApplicationDefinition.spec.versions[].version.
	//
	// +kubebuilder:validation:MinLength=1
	AppVersion string `json:"appVersion"`

	// RepositorySettings allows overriding the repository URL for this specific version.
	// Takes precedence over chart-level and global repository settings.
	//
	// +optional
	RepositorySettings *RepositorySettings `json:"repositorySettings,omitempty"`
}

// ChartConfig defines the configuration for a single Helm chart
// that will be converted to an ApplicationDefinition.
type ChartConfig struct {
	// ChartName is the name of the Helm chart in the repository.
	// This is used as the chart name when pulling from the repository.
	//
	// +kubebuilder:validation:MinLength=1
	ChartName string `json:"chartName"`

	// Metadata contains display information for the application.
	// If metadata.appName is not specified, chartName is used for ApplicationDefinition name.
	Metadata *ChartMetadata `json:"metadata,omitempty"`

	// RepositorySettings allows overriding the repository URL for this chart.
	// Takes precedence over global repository settings.
	//
	// +optional
	RepositorySettings *RepositorySettings `json:"repositorySettings,omitempty"`

	// DefaultValuesBlock contains the default Helm values for this application.
	// This is a YAML string that preserves comments.
	//
	// +optional
	DefaultValuesBlock string `json:"defaultValuesBlock,omitempty"`

	// ChartVersions lists the available versions of this chart.
	//
	// +kubebuilder:validation:MinItems=1
	ChartVersions []ChartVersion `json:"chartVersions"`
}

// GetAppName returns the application name to use for the ApplicationDefinition.
// Returns metadata.appName if set, otherwise falls back to chartName.
func (c *ChartConfig) GetAppName() string {
	if c.Metadata != nil && c.Metadata.AppName != "" {
		return c.Metadata.AppName
	}
	return c.ChartName
}
