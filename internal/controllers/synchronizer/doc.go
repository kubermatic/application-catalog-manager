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

// Package synchronizer implements a controller that watches ApplicationCatalog CRs
// and synchronizes them to ApplicationDefinition CRs.
//
// The controller converts each ChartConfig in an ApplicationCatalog into an
// ApplicationDefinition, applying URL resolution with the following precedence:
// version-level > chart-level > global > default.
//
// Key features:
// - Automatic cleanup of orphaned ApplicationDefinitions when charts are removed
// - Labels for tracking ownership (catalog-name, managed-by)
// - Annotation preservation (custom annotations are preserved during updates)
// - Management annotations for tracking sync state
// - DefaultValuesBlock preservation (user customizations are respected)
package synchronizer
