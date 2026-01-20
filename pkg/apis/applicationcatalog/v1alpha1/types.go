/*
Copyright 2026 The Application Catalog Manager contributors.

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

const (
	// ApplicationCatalogResourceName is the plural name of the ApplicationCatalog resource.
	ApplicationCatalogResourceName = "applicationcatalogs"

	// ApplicationCatalogKindName is the kind name of the ApplicationCatalog resource.
	ApplicationCatalogKindName = "ApplicationCatalog"
)

const (
	// LabelManagedByApplicationCatalog is applied to ApplicationDefinitions
	// created/managed by the ApplicationCatalog controller.
	LabelManagedByApplicationCatalog = "applicationcatalog.k8c.io/managed-by"

	// LabelApplicationCatalogName is applied to ApplicationDefinitions
	// to indicate which ApplicationCatalog generated them.
	LabelApplicationCatalogName = "applicationcatalog.k8c.io/catalog-name"
)

const (
	// AnnotationDefaultValuesGeneration tracks the generation of the ApplicationCatalog
	// when defaultValuesBlock was last synced. Used to preserve user customizations.
	AnnotationDefaultValuesGeneration = "applicationcatalog.k8c.io/default-values-generation"
)

const (
	// DefaultHelmRepository is the default OCI repository for Helm charts
	// when no repositorySettings.baseURL is specified.
	DefaultHelmRepository = "oci://quay.io/kubermatic-mirror/helm-charts"
)
