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
	"log"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"

	applicationcatalogmutation "k8c.io/application-catalog-manager/internal/pkg/admission/applicationcatalog/mutation"
	applicationcatalogvalidation "k8c.io/application-catalog-manager/internal/pkg/admission/applicationcatalog/validation"
	aclog "k8c.io/application-catalog-manager/internal/pkg/log"
	catalogv1alpha1 "k8c.io/application-catalog-manager/pkg/apis/applicationcatalog/v1alpha1"
	appskubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/apps.kubermatic/v1"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	ctrlruntimelog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(catalogv1alpha1.AddToScheme(scheme))
	utilruntime.Must(appskubermaticv1.AddToScheme(scheme))
}

type options struct {
	metricsAddr string
	probeAddr   string
	certDir     string
	webhookPort int
}

func main() {
	var opt options
	logFlags := aclog.NewDefaultOptions()
	logFlags.AddFlags(flag.CommandLine)

	flag.StringVar(&opt.metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to")
	flag.StringVar(&opt.probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to")
	flag.StringVar(&opt.certDir, "cert-dir", "/tmp/k8s-webhook-server/serving-certs", "Directory containing TLS certificates for the webhook server")
	flag.IntVar(&opt.webhookPort, "webhook-port", 9443, "Port for the webhook server")
	flag.Parse()

	rawLog := aclog.New(logFlags.Debug, logFlags.Format)
	l := rawLog.Sugar()

	ctrlruntimelog.SetLogger(zapr.NewLogger(rawLog.WithOptions(zap.AddCallerSkip(1))))

	l.Info("Initializing application-catalog webhook")

	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{
		Scheme:         scheme,
		LeaderElection: false,
		Metrics: metricsserver.Options{
			BindAddress: opt.metricsAddr,
		},
		HealthProbeBindAddress: opt.probeAddr,
		WebhookServer: webhook.NewServer(webhook.Options{
			CertDir: opt.certDir,
			Port:    opt.webhookPort,
		}),
	})
	if err != nil {
		log.Fatalf("Failed to create manager: %v", err)
	}

	applicationcatalogmutation.NewAdmissionHandler(
		rawLog.Sugar().Named("applicationcatalog-mutation"),
		scheme,
	).SetupWebhookWithManager(mgr)
	l.Info("ApplicationCatalog mutation webhook registered")

	applicationcatalogvalidation.NewAdmissionHandler(
		rawLog.Sugar().Named("applicationcatalog-validation"),
		scheme,
		mgr.GetClient(),
	).SetupWebhookWithManager(mgr)
	l.Info("ApplicationCatalog validation webhook registered")

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Fatalf("Failed to add health check: %v", err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Fatalf("Failed to add readiness check: %v", err)
	}

	l.Info("Starting webhook server")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Fatalf("Failed to start manager: %v", err)
	}
}
