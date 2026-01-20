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

package main

import (
	"flag"
	"time"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"

	"k8c.io/application-catalog-manager/internal/controllers/synchronizer"
	aclog "k8c.io/application-catalog-manager/internal/pkg/log"
	catalogv1alpha1 "k8c.io/application-catalog-manager/pkg/apis/applicationcatalog/v1alpha1"
	appskubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/apps.kubermatic/v1"
	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	ctrlruntimelog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(kubermaticv1.AddToScheme(scheme))
	utilruntime.Must(appskubermaticv1.AddToScheme(scheme))
	utilruntime.Must(catalogv1alpha1.AddToScheme(scheme))
}

type flags struct {
	reconciliationInterval time.Duration
	enableLeaderElection   bool
	healthProbeAddress     string
	metricsAddress         string
	namespace              string
}

func main() {
	var f flags
	logFlags := aclog.NewDefaultOptions()
	logFlags.AddFlags(flag.CommandLine)

	flag.DurationVar(&f.reconciliationInterval, "reconciliation-interval", 10*time.Minute, "Interval for reconciling ApplicationDefinitions")
	flag.BoolVar(&f.enableLeaderElection, "leader-elect", true, "Enable leader election for controller manager.")
	flag.StringVar(&f.healthProbeAddress, "health-probe-address", "127.0.0.1:8085", "The address on which the liveness check on /healthz and readiness check on /readyz will be available")
	flag.StringVar(&f.metricsAddress, "metrics-address", "127.0.0.1:8080", "The address on which Prometheus metrics will be available under /metrics")
	flag.StringVar(&f.namespace, "namespace", "kubermatic", "The namespace where the operator is deployed")

	flag.Parse()

	rawLog := aclog.New(logFlags.Debug, logFlags.Format)
	l := rawLog.Sugar()
	ctrlruntimelog.SetLogger(zapr.NewLogger(rawLog.WithOptions(zap.AddCallerSkip(1))))

	options := manager.Options{
		Scheme:                 scheme,
		LeaderElection:         f.enableLeaderElection,
		LeaderElectionID:       "application-catalog-manager",
		HealthProbeBindAddress: f.healthProbeAddress,
		Metrics: metricsserver.Options{
			BindAddress: f.metricsAddress,
		},
	}

	mgr, err := manager.New(config.GetConfigOrDie(), options)
	if err != nil {
		l.Fatalf("Failed to create manager: %v", err)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		l.Fatalf("Failed to set up health check: %v", err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		l.Fatalf("Failed to set up ready check: %v", err)
	}

	err = synchronizer.Add(mgr, &synchronizer.ControllerConfig{
		Log:                    rawLog.Sugar().Named("synchronizer"),
		ReconciliationInterval: f.reconciliationInterval,
	})
	if err != nil {
		l.Fatalf("Failed to add synchronizer controller: %v", err)
	}

	l.Infof("Starting manager, with reconciliation interval %s", f.reconciliationInterval)

	if err = mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		l.Fatalf("Failed to start manager: %v", err)
	}
}
