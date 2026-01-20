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

package validation

import (
	"context"
	"strings"
	"testing"

	"go.uber.org/zap"

	catalogv1alpha1 "k8c.io/application-catalog-manager/pkg/apis/applicationcatalog/v1alpha1"
	appskubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/apps.kubermatic/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlruntimefakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func setupTestHandler(t *testing.T, objects ...ctrlruntimeclient.Object) *AdmissionHandler {
	t.Helper()

	scheme := runtime.NewScheme()
	if err := catalogv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add catalogv1alpha1 to scheme: %v", err)
	}
	if err := appskubermaticv1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add appskubermaticv1 to scheme: %v", err)
	}

	fakeClient := ctrlruntimefakeclient.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objects...).
		Build()

	logger := zap.NewNop().Sugar()

	return &AdmissionHandler{
		log:    logger,
		client: fakeClient,
	}
}

func TestDetectConflicts_NoConflicts_NoExistingAppDefs(t *testing.T) {
	handler := setupTestHandler(t)

	catalog := &catalogv1alpha1.ApplicationCatalog{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-catalog",
		},
		Spec: catalogv1alpha1.ApplicationCatalogSpec{
			Helm: &catalogv1alpha1.HelmSpec{
				Charts: []catalogv1alpha1.ChartConfig{
					{
						ChartName: "nginx",
						ChartVersions: []catalogv1alpha1.ChartVersion{
							{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
						},
					},
					{
						ChartName: "redis",
						ChartVersions: []catalogv1alpha1.ChartVersion{
							{ChartVersion: "2.0.0", AppVersion: "v2.0.0"},
						},
					},
				},
			},
		},
	}

	conflicts, err := handler.detectConflicts(context.Background(), catalog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts, got %d: %+v", len(conflicts), conflicts)
	}
}

func TestDetectConflicts_NoConflicts_SameCatalogOwns(t *testing.T) {
	existingAppDef := &appskubermaticv1.ApplicationDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nginx",
			Labels: map[string]string{
				catalogv1alpha1.LabelManagedByApplicationCatalog: "true",
				catalogv1alpha1.LabelApplicationCatalogName:      "my-catalog",
			},
		},
	}

	handler := setupTestHandler(t, existingAppDef)

	catalog := &catalogv1alpha1.ApplicationCatalog{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-catalog",
		},
		Spec: catalogv1alpha1.ApplicationCatalogSpec{
			Helm: &catalogv1alpha1.HelmSpec{
				Charts: []catalogv1alpha1.ChartConfig{
					{
						ChartName: "nginx",
						ChartVersions: []catalogv1alpha1.ChartVersion{
							{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
						},
					},
				},
			},
		},
	}

	conflicts, err := handler.detectConflicts(context.Background(), catalog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts when same catalog owns, got %d: %+v", len(conflicts), conflicts)
	}
}

func TestDetectConflicts_NoConflicts_UnmanagedAppDef(t *testing.T) {
	// ApplicationDefinition exists but has no catalog ownership labels
	existingAppDef := &appskubermaticv1.ApplicationDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "nginx",
			Labels: map[string]string{}, // No ownership labels
		},
	}

	handler := setupTestHandler(t, existingAppDef)

	catalog := &catalogv1alpha1.ApplicationCatalog{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-catalog",
		},
		Spec: catalogv1alpha1.ApplicationCatalogSpec{
			Helm: &catalogv1alpha1.HelmSpec{
				Charts: []catalogv1alpha1.ChartConfig{
					{
						ChartName: "nginx",
						ChartVersions: []catalogv1alpha1.ChartVersion{
							{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
						},
					},
				},
			},
		},
	}

	conflicts, err := handler.detectConflicts(context.Background(), catalog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts for unmanaged AppDef, got %d: %+v", len(conflicts), conflicts)
	}
}

func TestDetectConflicts_SingleConflict(t *testing.T) {
	existingAppDef := &appskubermaticv1.ApplicationDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nginx",
			Labels: map[string]string{
				catalogv1alpha1.LabelManagedByApplicationCatalog: "true",
				catalogv1alpha1.LabelApplicationCatalogName:      "other-catalog",
			},
		},
	}

	handler := setupTestHandler(t, existingAppDef)

	catalog := &catalogv1alpha1.ApplicationCatalog{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-catalog",
		},
		Spec: catalogv1alpha1.ApplicationCatalogSpec{
			Helm: &catalogv1alpha1.HelmSpec{
				Charts: []catalogv1alpha1.ChartConfig{
					{
						ChartName: "nginx",
						ChartVersions: []catalogv1alpha1.ChartVersion{
							{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
						},
					},
				},
			},
		},
	}

	conflicts, err := handler.detectConflicts(context.Background(), catalog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}

	if conflicts[0].AppDefName != "nginx" {
		t.Errorf("expected AppDefName 'nginx', got %q", conflicts[0].AppDefName)
	}
	if conflicts[0].OwnerCatalog != "other-catalog" {
		t.Errorf("expected OwnerCatalog 'other-catalog', got %q", conflicts[0].OwnerCatalog)
	}
}

func TestDetectConflicts_MultipleConflicts(t *testing.T) {
	existingAppDefs := []ctrlruntimeclient.Object{
		&appskubermaticv1.ApplicationDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "nginx",
				Labels: map[string]string{
					catalogv1alpha1.LabelManagedByApplicationCatalog: "true",
					catalogv1alpha1.LabelApplicationCatalogName:      "catalog-a",
				},
			},
		},
		&appskubermaticv1.ApplicationDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "redis",
				Labels: map[string]string{
					catalogv1alpha1.LabelManagedByApplicationCatalog: "true",
					catalogv1alpha1.LabelApplicationCatalogName:      "catalog-b",
				},
			},
		},
		&appskubermaticv1.ApplicationDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "postgres", // Unmanaged - no conflict
				Labels: map[string]string{},
			},
		},
	}

	handler := setupTestHandler(t, existingAppDefs...)

	catalog := &catalogv1alpha1.ApplicationCatalog{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-catalog",
		},
		Spec: catalogv1alpha1.ApplicationCatalogSpec{
			Helm: &catalogv1alpha1.HelmSpec{
				Charts: []catalogv1alpha1.ChartConfig{
					{ChartName: "nginx", ChartVersions: []catalogv1alpha1.ChartVersion{{ChartVersion: "1.0.0", AppVersion: "v1.0.0"}}},
					{ChartName: "redis", ChartVersions: []catalogv1alpha1.ChartVersion{{ChartVersion: "2.0.0", AppVersion: "v2.0.0"}}},
					{ChartName: "postgres", ChartVersions: []catalogv1alpha1.ChartVersion{{ChartVersion: "3.0.0", AppVersion: "v3.0.0"}}},
					{ChartName: "new-app", ChartVersions: []catalogv1alpha1.ChartVersion{{ChartVersion: "4.0.0", AppVersion: "v4.0.0"}}},
				},
			},
		},
	}

	conflicts, err := handler.detectConflicts(context.Background(), catalog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conflicts) != 2 {
		t.Fatalf("expected 2 conflicts, got %d: %+v", len(conflicts), conflicts)
	}

	conflictMap := make(map[string]string)
	for _, c := range conflicts {
		conflictMap[c.AppDefName] = c.OwnerCatalog
	}

	if conflictMap["nginx"] != "catalog-a" {
		t.Errorf("expected nginx conflict with catalog-a, got %q", conflictMap["nginx"])
	}
	if conflictMap["redis"] != "catalog-b" {
		t.Errorf("expected redis conflict with catalog-b, got %q", conflictMap["redis"])
	}
}

func TestDetectConflicts_WithMetadataAppName(t *testing.T) {
	// AppDef uses appName from metadata, not chartName
	existingAppDef := &appskubermaticv1.ApplicationDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nvidia-gpu-operator", // This is the appName
			Labels: map[string]string{
				catalogv1alpha1.LabelManagedByApplicationCatalog: "true",
				catalogv1alpha1.LabelApplicationCatalogName:      "other-catalog",
			},
		},
	}

	handler := setupTestHandler(t, existingAppDef)

	catalog := &catalogv1alpha1.ApplicationCatalog{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-catalog",
		},
		Spec: catalogv1alpha1.ApplicationCatalogSpec{
			Helm: &catalogv1alpha1.HelmSpec{
				Charts: []catalogv1alpha1.ChartConfig{
					{
						ChartName: "gpu-operator", // Different chart name
						Metadata: &catalogv1alpha1.ChartMetadata{
							AppName: "nvidia-gpu-operator", // But same appName
						},
						ChartVersions: []catalogv1alpha1.ChartVersion{
							{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
						},
					},
				},
			},
		},
	}

	conflicts, err := handler.detectConflicts(context.Background(), catalog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}

	if conflicts[0].AppDefName != "nvidia-gpu-operator" {
		t.Errorf("expected conflict on 'nvidia-gpu-operator', got %q", conflicts[0].AppDefName)
	}
}

func TestDetectConflicts_NilCharts(t *testing.T) {
	handler := setupTestHandler(t)

	catalog := &catalogv1alpha1.ApplicationCatalog{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-catalog",
		},
		Spec: catalogv1alpha1.ApplicationCatalogSpec{
			Helm: nil, // No helm spec
		},
	}

	conflicts, err := handler.detectConflicts(context.Background(), catalog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts for nil charts, got %d", len(conflicts))
	}
}

func TestDetectConflicts_EmptyCharts(t *testing.T) {
	handler := setupTestHandler(t)

	catalog := &catalogv1alpha1.ApplicationCatalog{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-catalog",
		},
		Spec: catalogv1alpha1.ApplicationCatalogSpec{
			Helm: &catalogv1alpha1.HelmSpec{
				Charts: []catalogv1alpha1.ChartConfig{}, // Empty list
			},
		},
	}

	conflicts, err := handler.detectConflicts(context.Background(), catalog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts for empty charts, got %d", len(conflicts))
	}
}

func TestDetectConflicts_IntraCatalogDuplicates(t *testing.T) {
	handler := setupTestHandler(t)

	catalog := &catalogv1alpha1.ApplicationCatalog{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-catalog",
		},
		Spec: catalogv1alpha1.ApplicationCatalogSpec{
			Helm: &catalogv1alpha1.HelmSpec{
				Charts: []catalogv1alpha1.ChartConfig{
					{
						ChartName: "nginx",
						ChartVersions: []catalogv1alpha1.ChartVersion{
							{ChartVersion: "1.0.0", AppVersion: "v1.0.0"},
						},
					},
					{
						ChartName: "my-nginx",
						Metadata: &catalogv1alpha1.ChartMetadata{
							AppName: "nginx", // Same appName as first chart
						},
						ChartVersions: []catalogv1alpha1.ChartVersion{
							{ChartVersion: "2.0.0", AppVersion: "v2.0.0"},
						},
					},
				},
			},
		},
	}

	conflicts, err := handler.detectConflicts(context.Background(), catalog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict for duplicate appName, got %d: %+v", len(conflicts), conflicts)
	}

	if conflicts[0].AppDefName != "nginx" {
		t.Errorf("expected AppDefName 'nginx', got %q", conflicts[0].AppDefName)
	}

	if !strings.Contains(conflicts[0].OwnerCatalog, "duplicate") {
		t.Errorf("expected OwnerCatalog to mention duplicate, got %q", conflicts[0].OwnerCatalog)
	}
}

func TestFormatConflictMessage(t *testing.T) {
	tests := []struct {
		name      string
		conflicts []ConflictInfo
		contains  []string
	}{
		{
			name:      "empty conflicts",
			conflicts: []ConflictInfo{},
			contains:  nil, // Empty string expected
		},
		{
			name: "single conflict",
			conflicts: []ConflictInfo{
				{AppDefName: "nginx", OwnerCatalog: "other-catalog"},
			},
			contains: []string{
				"ApplicationCatalog conflicts detected",
				"nginx",
				"other-catalog",
				"To resolve this conflict",
			},
		},
		{
			name: "multiple conflicts",
			conflicts: []ConflictInfo{
				{AppDefName: "nginx", OwnerCatalog: "catalog-a"},
				{AppDefName: "redis", OwnerCatalog: "catalog-b"},
			},
			contains: []string{
				"nginx",
				"catalog-a",
				"redis",
				"catalog-b",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := formatConflictMessage(tc.conflicts)

			if len(tc.conflicts) == 0 {
				if msg != "" {
					t.Errorf("expected empty message for empty conflicts, got %q", msg)
				}
				return
			}

			for _, expected := range tc.contains {
				if !strings.Contains(msg, expected) {
					t.Errorf("expected message to contain %q, got:\n%s", expected, msg)
				}
			}
		})
	}
}
