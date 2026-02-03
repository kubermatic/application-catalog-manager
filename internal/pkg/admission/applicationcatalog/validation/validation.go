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

// Package validation provides a validating admission webhook for ApplicationCatalog.
// It prevents conflicts when multiple catalogs attempt to manage the same ApplicationDefinition.
package validation

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"k8c.io/application-catalog-manager/internal/pkg/defaulting"
	catalogv1alpha1 "k8c.io/application-catalog-manager/pkg/apis/applicationcatalog/v1alpha1"
	appskubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/apps.kubermatic/v1"

	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	// WebhookPath is the HTTP path for this webhook.
	WebhookPath = "/validate-applicationcatalog-k8c-io-v1alpha1-applicationcatalog"
)

// AdmissionHandler handles validating admission requests for ApplicationCatalog.
type AdmissionHandler struct {
	log     *zap.SugaredLogger
	decoder admission.Decoder
	client  ctrlruntimeclient.Client
}

// NewAdmissionHandler creates a new AdmissionHandler.
func NewAdmissionHandler(log *zap.SugaredLogger, scheme *runtime.Scheme, client ctrlruntimeclient.Client) *AdmissionHandler {
	return &AdmissionHandler{
		log:     log,
		decoder: admission.NewDecoder(scheme),
		client:  client,
	}
}

// SetupWebhookWithManager registers the webhook with the manager.
func (h *AdmissionHandler) SetupWebhookWithManager(mgr ctrl.Manager) {
	mgr.GetWebhookServer().Register(WebhookPath, &webhook.Admission{Handler: h})
}

// Handle handles admission requests for ApplicationCatalog CR.
// It validates that the catalog doesn't conflict with existing ApplicationDefinitions
// managed by other catalogs.
func (h *AdmissionHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := h.log.With("uid", req.UID, "name", req.Name, "operation", req.Operation)

	switch req.Operation {
	case admissionv1.Create, admissionv1.Update:
		return h.handleValidation(ctx, log, req)
	case admissionv1.Delete:
		log.Debug("Allowing delete operation without validation")
		return admission.Allowed("delete operations do not require validation")
	default:
		log.Debugw("Allowing operation without validation", "operation", req.Operation)
		return admission.Allowed(fmt.Sprintf("%q operations do not require validation", req.Operation))
	}
}

// handleValidation validates the ApplicationCatalog for conflicts.
func (h *AdmissionHandler) handleValidation(ctx context.Context, log *zap.SugaredLogger, req admission.Request) admission.Response {
	catalog := &catalogv1alpha1.ApplicationCatalog{}
	if err := h.decoder.Decode(req, catalog); err != nil {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to decode request: %w", err))
	}

	// Validate include annotation before checking conflicts
	if catalog.Spec.Helm != nil && catalog.Spec.Helm.IncludeDefaults {
		annotation := catalog.Annotations["defaultcatalog.k8c.io/include"]
		if invalidNames := defaulting.ValidateIncludeAnnotation(annotation); len(invalidNames) > 0 {
			validNames := defaulting.GetDefaultChartNames()
			return admission.Denied(fmt.Sprintf("invalid chart names in annotation defaultcatalog.k8c.io/include: %v. Valid names are: %v", invalidNames, validNames))
		}
	}

	conflicts, err := h.detectConflicts(ctx, catalog)
	if err != nil {
		log.Errorw("Failed to detect conflicts", "error", err)
		return admission.Errored(http.StatusInternalServerError, fmt.Errorf("failed to validate catalog: %w", err))
	}

	if len(conflicts) > 0 {
		log.Warnw("Catalog conflicts detected", "conflicts", conflicts)
		return admission.Denied(formatConflictMessage(conflicts))
	}

	log.Debug("Validation passed, no conflicts detected")
	return admission.Allowed("no conflicts detected")
}

// ConflictInfo contains information about a detected conflict.
type ConflictInfo struct {
	AppDefName   string
	OwnerCatalog string
}

// detectConflicts checks for intra-catalog duplicates and external conflicts
// with ApplicationDefinitions managed by other catalogs.
func (h *AdmissionHandler) detectConflicts(ctx context.Context, catalog *catalogv1alpha1.ApplicationCatalog) ([]ConflictInfo, error) {
	charts := catalog.GetHelmCharts()
	if len(charts) == 0 {
		return nil, nil
	}

	var conflicts []ConflictInfo

	// Check for intra-catalog duplicates
	appNames := make(map[string]string, len(charts))
	for i := range charts {
		chart := &charts[i]
		appName := chart.GetAppName()

		if existingChart, exists := appNames[appName]; exists {
			conflicts = append(conflicts, ConflictInfo{
				AppDefName:   appName,
				OwnerCatalog: fmt.Sprintf("this catalog (duplicate: charts %q and %q resolve to same appName)", existingChart, chart.ChartName),
			})
			continue
		}
		appNames[appName] = chart.ChartName
	}

	// Check for external conflicts
	appDefList := &appskubermaticv1.ApplicationDefinitionList{}
	listOpts := &ctrlruntimeclient.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{
			catalogv1alpha1.LabelManagedByApplicationCatalog: "true",
		}),
	}

	if err := h.client.List(ctx, appDefList, listOpts); err != nil {
		return nil, fmt.Errorf("failed to list ApplicationDefinitions: %w", err)
	}

	existing := make(map[string]*appskubermaticv1.ApplicationDefinition)
	for i := range appDefList.Items {
		appDef := &appDefList.Items[i]
		if _, ok := appNames[appDef.Name]; ok {
			existing[appDef.Name] = appDef
		}
	}

	for appName := range appNames {
		appDef, found := existing[appName]
		if !found {
			continue
		}

		owner := appDef.Labels[catalogv1alpha1.LabelApplicationCatalogName]
		if owner == "" || owner == catalog.Name {
			continue
		}

		conflicts = append(conflicts, ConflictInfo{
			AppDefName:   appName,
			OwnerCatalog: owner,
		})
	}

	return conflicts, nil
}

// formatConflictMessage formats the conflict information into a user-friendly message.
func formatConflictMessage(conflicts []ConflictInfo) string {
	if len(conflicts) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("ApplicationCatalog conflicts detected:\n")

	for _, c := range conflicts {
		sb.WriteString(fmt.Sprintf("  - ApplicationDefinition %q is already managed by catalog %q\n", c.AppDefName, c.OwnerCatalog))
	}

	sb.WriteString("\nTo resolve this conflict, either:\n")
	sb.WriteString("  1. Remove the conflicting chart from this catalog\n")
	sb.WriteString("  2. Use a different appName in metadata.appName for the chart\n")
	sb.WriteString("  3. Delete the other catalog or remove the chart from it first")

	return sb.String()
}
