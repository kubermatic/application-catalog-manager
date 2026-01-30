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
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"go.uber.org/zap"

	"k8c.io/application-catalog-manager/internal/pkg/kubernetes"
	catalogv1alpha1 "k8c.io/application-catalog-manager/pkg/apis/applicationcatalog/v1alpha1"
	appskubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/apps.kubermatic/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var errRequeueAfter10Secs = fmt.Errorf("requeue after 10 seconds")

func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	l := r.logger.With("catalog", req.Name)
	l.Info("Reconciling ApplicationCatalog")

	err := r.reconcile(ctx, l, req)
	if err != nil {
		if errors.Is(err, errRequeueAfter10Secs) {
			l.Debug("requeuing after 10 seconds")
			return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
		}
		return reconcile.Result{}, err
	}

	if r.cfg.ReconciliationInterval > 0 {
		return reconcile.Result{RequeueAfter: r.cfg.ReconciliationInterval}, nil
	}

	return reconcile.Result{}, nil
}

func (r *Reconciler) reconcile(ctx context.Context, l *zap.SugaredLogger, req reconcile.Request) error {
	catalog := &catalogv1alpha1.ApplicationCatalog{}
	if err := r.Get(ctx, req.NamespacedName, catalog); err != nil {
		if apierrors.IsNotFound(err) {
			l.Info("ApplicationCatalog not found, unmanaging orphaned ApplicationDefinitions")

			return r.handleDeletion(ctx, req.Name)
		}

		return fmt.Errorf("failed to get ApplicationCatalog: %w", err)
	}

	charts := catalog.GetHelmCharts()

	// If charts is nil but includeDefaults is true, the webhook should have
	// populated defaults but didn't run or failed. Requeue to wait for webhook.
	if charts == nil && catalog.Spec.Helm != nil && catalog.Spec.Helm.IncludeDefaults {
		l.Info("Waiting for webhook to populate default charts")
		return errRequeueAfter10Secs
	}

	// If charts is nil and includeDefaults is false (or spec.helm is nil),
	// this is a valid empty catalog. Convert nil to empty slice.
	if charts == nil {
		charts = []catalogv1alpha1.ChartConfig{}
	}

	var errs []error
	generatedApps := make(map[string]bool)

	for i := range charts {
		chart := &charts[i]
		desired := convertChartToApplicationDefinition(catalog, chart)
		generatedApps[desired.Name] = true

		if err := r.reconcileApplicationDefinition(ctx, l, desired); err != nil {
			errs = append(errs, fmt.Errorf("chart %q: %w", chart.ChartName, err))
		}
	}

	err := r.unmanageOrphans(ctx, catalog.Name, generatedApps)
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return kerrors.NewAggregate(errs)
	}

	l.Infof("Reconciliation complete")
	return nil
}

// reconcileApplicationDefinition creates or updates an ApplicationDefinition.
func (r *Reconciler) reconcileApplicationDefinition(
	ctx context.Context,
	l *zap.SugaredLogger,
	desired *appskubermaticv1.ApplicationDefinition,
) error {
	l.Debugw("reconciling", "applicationcatalog", ctrlruntimeclient.ObjectKeyFromObject(desired))

	existing := &appskubermaticv1.ApplicationDefinition{}

	err := r.Get(ctx, ctrlruntimeclient.ObjectKey{Name: desired.Name}, existing)
	if err != nil {
		if apierrors.IsNotFound(err) {
			l.Debugw("Creating ApplicationDefinition", "name", desired.Name)
			return r.Create(ctx, desired)
		}
		return fmt.Errorf("failed to get ApplicationDefinition %q: %w", desired.Name, err)
	}

	if isSystemApplication(existing) {
		l.Debugw("Skipping system-application reconciliation", "name", existing.Name)
		return nil
	}

	return r.updateApplicationDefinition(ctx, l, existing, desired)
}

// updateApplicationDefinition updates an existing ApplicationDefinition.
// It preserves user customizations using KKP's pattern:
// promote preserved fields to desired, then do full spec replacement.
//
// Fields preserved from existing (cluster state has higher precedence):
//   - Enforced: if true in cluster, preserve it
//   - Default: if true in cluster, preserve it
//   - Selector.Datacenters: if set in cluster, preserve it
//   - DefaultValuesBlock: if non-empty in cluster, preserve it
//   - Versions: merged (existing versions preserved, new ones added/updated)
func (r *Reconciler) updateApplicationDefinition(
	ctx context.Context,
	l *zap.SugaredLogger,
	existing, desired *appskubermaticv1.ApplicationDefinition,
) error {
	l.Debugw("Updating ApplicationDefinition", "name", existing.Name)

	return kubernetes.PatchObject(ctx, r.Client, existing, func() {
		kubernetes.EnsureLabels(existing, desired.Labels)
		kubernetes.EnsureAnnotations(existing, desired.Annotations)

		// Preserve fields where cluster state has higher precedence than catalog.
		// This follows KKP's pattern from pkg/ee/default-application-catalog/application_catalog.go
		if existing.Spec.Enforced {
			desired.Spec.Enforced = true
		}
		if existing.Spec.Default {
			desired.Spec.Default = true
		}
		if existing.Spec.Selector.Datacenters != nil {
			desired.Spec.Selector.Datacenters = existing.Spec.Selector.Datacenters
		}

		// Preserve the user customization unless the defaultValuesBlock is empty or "{}"
		// In case of empty or "{}", application-catalog enforces the desired state to keep the KKP's existing pattern
		// in order to prevent breaking changes.
		if existing.Spec.DefaultValuesBlock != "" && existing.Spec.DefaultValuesBlock != "{}" {
			desired.Spec.DefaultValuesBlock = existing.Spec.DefaultValuesBlock
		}

		desired.Spec.Versions = mergeVersions(existing.Spec.Versions, desired.Spec.Versions)
		// Sort versions to have a deterministic order
		sort.Slice(desired.Spec.Versions, func(i, j int) bool {
			return desired.Spec.Versions[i].Version < desired.Spec.Versions[j].Version
		})
		existing.Spec = desired.Spec
	})
}

func (r *Reconciler) handleDeletion(ctx context.Context, catalogName string) error {
	return r.unmanageOrphans(ctx, catalogName, nil)
}

// unmanageOrphans removes managed labels from ApplicationDefinitions that are no longer in the catalog.
// If `generatedApps` is nil, it unmanages all ApplicationDefinitions for the catalog.
func (r *Reconciler) unmanageOrphans(ctx context.Context, catalogName string, generatedApps map[string]bool) error {
	appDefList := &appskubermaticv1.ApplicationDefinitionList{}
	listOpts := &ctrlruntimeclient.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{
			catalogv1alpha1.LabelApplicationCatalogName: catalogName,
		}),
	}

	if err := r.List(ctx, appDefList, listOpts); err != nil {
		return fmt.Errorf("failed to list ApplicationDefinitions: %w", err)
	}

	var errs []error
	for i := range appDefList.Items {
		appDef := &appDefList.Items[i]
		if generatedApps == nil || !generatedApps[appDef.Name] {
			r.logger.Debug("Removing managed labels from ApplicationDefinition %s", appDef.Name)
			if err := r.removeManagedLabels(ctx, appDef); err != nil {
				errs = append(errs, fmt.Errorf("failed to unmanage %s: %w", appDef.Name, err))
			}
		}
	}

	return kerrors.NewAggregate(errs)
}

// removeManagedLabels removes the catalog management labels from an ApplicationDefinition.
func (r *Reconciler) removeManagedLabels(ctx context.Context, appDef *appskubermaticv1.ApplicationDefinition) error {
	return kubernetes.PatchObject(ctx, r.Client, appDef, func() {
		if appDef.Labels != nil {
			delete(appDef.Labels, catalogv1alpha1.LabelManagedByApplicationCatalog)
			delete(appDef.Labels, catalogv1alpha1.LabelApplicationCatalogName)
		}
	})
}

// mergeVersions merges existing versions with desired ones without removing
// existing versions (to prevent breaking changes in KKP deployments).
func mergeVersions(existing, desired []appskubermaticv1.ApplicationVersion) []appskubermaticv1.ApplicationVersion {
	if len(existing) == 0 {
		return desired
	}
	if len(desired) == 0 {
		return existing
	}

	existingIndex := make(map[string]int, len(existing))
	for i, v := range existing {
		existingIndex[v.Version] = i
	}

	result := make([]appskubermaticv1.ApplicationVersion, len(existing))
	copy(result, existing)

	for _, d := range desired {
		if idx, exists := existingIndex[d.Version]; exists {
			result[idx] = d
			continue
		}

		result = append(result, d)
	}

	return result
}

func isSystemApplication(appDef *appskubermaticv1.ApplicationDefinition) bool {
	if appDef == nil {
		return false
	}

	if labels := appDef.GetLabels(); labels != nil {
		if _, exists := labels["apps.kubermatic.k8c.io/managed-by"]; exists {
			return true
		}
	}
	return false
}
