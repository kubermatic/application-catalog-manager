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
	"flag"
	"os"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

var (
	testEnv env.Environment
)

func TestMain(m *testing.M) {
	flag.Parse()

	testEnv = env.New().
		AfterEachTest(func(ctx context.Context, config *envconf.Config, _ *testing.T) (context.Context, error) {
			return cleanUpTestsAfter(ctx, config)
		}).
		Finish(cleanUpTestsAfter)

	os.Exit(testEnv.Run(m))
}

func cleanUpTestsAfter(ctx context.Context, config *envconf.Config) (context.Context, error) {
	s := suite{}
	err := s.withClient(config.Client())
	if err != nil {
		return nil, err
	}

	err = s.cleanup(ctx)
	if err != nil {
		return nil, err
	}

	return ctx, nil
}
