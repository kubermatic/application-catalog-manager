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
	"fmt"
	"time"

	"go.uber.org/zap"

	catalogv1alpha1 "k8c.io/application-catalog-manager/pkg/apis/applicationcatalog/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/builder"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	controllerName = "SynchronizerController"
)

// ControllerConfig holds the configuration for the synchronizer controller.
type ControllerConfig struct {
	Log *zap.SugaredLogger

	// ReconciliationInterval is the duration after which the controller will requeue
	// the ApplicationCatalog for reconciliation. When set to 0, the default reconciliation
	// interval is used.
	ReconciliationInterval time.Duration
}

func (c *ControllerConfig) validate() error {
	if c.Log == nil {
		return fmt.Errorf("log cannot be nil")
	}

	if c.ReconciliationInterval < 0 {
		return fmt.Errorf("reconciliation interval must be a non-negative duration")
	}

	return nil
}

// Reconciler reconciles ApplicationCatalog objects.
type Reconciler struct {
	ctrlruntimeclient.Client
	cfg    *ControllerConfig
	logger *zap.SugaredLogger
}

// Add creates a new Synchronizer controller and adds it to the Manager.
// The Manager will set fields on the Reconciler and start it when the Manager is started.
func Add(mgr manager.Manager, cfg *ControllerConfig) error {
	if cfg == nil {
		return fmt.Errorf("failed to instantiate controller: config is nil")
	}

	if err := cfg.validate(); err != nil {
		return fmt.Errorf("failed to instantiate controller: %w", err)
	}

	reconciler := &Reconciler{
		Client: mgr.GetClient(),
		cfg:    cfg,
		logger: cfg.Log,
	}

	// Watch ApplicationCatalog as the primary resource.
	_, err := builder.ControllerManagedBy(mgr).
		Named(controllerName).
		For(&catalogv1alpha1.ApplicationCatalog{}).
		Build(reconciler)

	return err
}
