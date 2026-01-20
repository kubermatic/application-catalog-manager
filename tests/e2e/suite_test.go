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

package e2e_test

import (
	"context"
	"errors"
	"time"

	catalogv1alpha1 "k8c.io/application-catalog-manager/pkg/apis/applicationcatalog/v1alpha1"
	appskubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/apps.kubermatic/v1"
	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/klient/wait"
)

var errClientNotInitialized = errors.New("client is not initialized")

type suite struct {
	client client.Client
}

func (s *suite) withClient(kl klient.Client) error {
	scheme := runtime.NewScheme()

	cl, err := client.New(kl.RESTConfig(), client.Options{Scheme: scheme})
	if err != nil {
		return err
	}

	schemeBuilders := []runtime.SchemeBuilder{
		appskubermaticv1.SchemeBuilder,
		kubermaticv1.SchemeBuilder,
		corev1.SchemeBuilder,
		catalogv1alpha1.SchemeBuilder,
	}

	for _, builder := range schemeBuilders {
		err = builder.AddToScheme(scheme)
		if err != nil {
			return err
		}
	}

	s.client = cl
	return nil
}

func (s *suite) cleanupAllApplicationDefinitions(ctx context.Context) error {
	if s.client == nil {
		return errClientNotInitialized
	}

	return waitFor(ctx, func(ctx context.Context) (bool, error) {
		appDefs := appskubermaticv1.ApplicationDefinitionList{}
		err := s.client.List(ctx, &appDefs)
		if err != nil {
			return false, err
		}

		for _, app := range appDefs.Items {
			err := s.client.Delete(ctx, &app)
			if err != nil && !apierrors.IsNotFound(err) {
				return false, nil
			}
		}

		err = s.client.List(ctx, &appDefs)
		if err != nil {
			return false, err
		}

		return len(appDefs.Items) == 0, nil
	})
}

func (s *suite) cleanup(ctx context.Context) error {
	appDefErr := s.cleanupAllApplicationDefinitions(ctx)

	if appDefErr != nil {
		return appDefErr
	}

	return nil
}

const (
	timeout  = time.Minute * 1
	interval = time.Second * 1
)

func waitFor(ctx context.Context, f func(ctx context.Context) (bool, error)) error {
	err := wait.For(
		f,
		wait.WithTimeout(timeout),
		wait.WithInterval(interval),
		wait.WithContext(ctx),
	)

	return err
}
