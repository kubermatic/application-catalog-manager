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

// Package mutation provides a mutating admission webhook for ApplicationCatalog.
// It injects default charts when spec.helm.charts is nil.
package mutation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"k8c.io/application-catalog-manager/internal/pkg/defaulting"
	catalogv1alpha1 "k8c.io/application-catalog-manager/pkg/apis/applicationcatalog/v1alpha1"

	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	// WebhookPath is the HTTP path for this webhook.
	WebhookPath = "/mutate-applicationcatalog-k8c-io-v1alpha1-applicationcatalog"
)

// AdmissionHandler handles mutating admission requests for ApplicationCatalog.
type AdmissionHandler struct {
	log     *zap.SugaredLogger
	decoder admission.Decoder
}

// NewAdmissionHandler creates a new AdmissionHandler.
func NewAdmissionHandler(log *zap.SugaredLogger, scheme *runtime.Scheme) *AdmissionHandler {
	return &AdmissionHandler{
		log:     log,
		decoder: admission.NewDecoder(scheme),
	}
}

// SetupWebhookWithManager registers the webhook with the manager.
func (h *AdmissionHandler) SetupWebhookWithManager(mgr ctrl.Manager) {
	mgr.GetWebhookServer().Register(WebhookPath, &webhook.Admission{Handler: h})
}

// Handle handles admission requests for ApplicationCatalog CR.
// It applies the defaulting logic on Create and Update operations.
func (h *AdmissionHandler) Handle(_ context.Context, req admission.Request) admission.Response {
	log := h.log.With("uid", req.UID, "name", req.Name, "operation", req.Operation)

	switch req.Operation {
	case admissionv1.Create, admissionv1.Update:
		return h.handleMutation(log, req)
	default:
		log.Debugw("Allowing operation without mutation", "operation", req.Operation)
		return admission.Allowed(fmt.Sprintf("%q operations do not require mutation", req.Operation))
	}
}

// handleMutation applies defaulting to the ApplicationCatalog and returns a patch response.
func (h *AdmissionHandler) handleMutation(log *zap.SugaredLogger, req admission.Request) admission.Response {
	catalog := &catalogv1alpha1.ApplicationCatalog{}
	if err := h.decoder.Decode(req, catalog); err != nil {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to decode request: %w", err))
	}

	defaulting.DefaultApplicationCatalog(catalog)

	chartsWereNil := catalog.Spec.Helm == nil || catalog.Spec.Helm.Charts == nil
	if chartsWereNil {
		log.Infow("Injected default charts", "chartCount", len(catalog.Spec.Helm.Charts))
	} else {
		log.Debug("Charts already specified, no defaults injected")
	}

	mutatedData, err := json.Marshal(catalog)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, fmt.Errorf("failed to marshal mutated object: %w", err))
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, mutatedData)
}
