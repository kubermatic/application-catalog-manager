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
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestControllerConfigValidate(t *testing.T) {
	logger := zap.NewNop().Sugar()

	tests := []struct {
		name        string
		cfg         *ControllerConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config with positive interval",
			cfg: &ControllerConfig{
				Log:                    logger,
				ReconciliationInterval: 10 * time.Minute,
			},
			expectError: false,
		},
		{
			name: "valid config with zero interval (disables periodic reconciliation)",
			cfg: &ControllerConfig{
				Log:                    logger,
				ReconciliationInterval: 0,
			},
			expectError: false,
		},
		{
			name: "invalid config with negative interval",
			cfg: &ControllerConfig{
				Log:                    logger,
				ReconciliationInterval: -1 * time.Minute,
			},
			expectError: true,
			errorMsg:    "reconciliation interval must be a non-negative duration",
		},
		{
			name: "invalid config with nil logger",
			cfg: &ControllerConfig{
				Log:                    nil,
				ReconciliationInterval: 10 * time.Minute,
			},
			expectError: true,
			errorMsg:    "log cannot be nil",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.validate()

			if tc.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
				}
				if err.Error() != tc.errorMsg {
					t.Errorf("expected error message %q, got %q", tc.errorMsg, err.Error())
				}
				return
			}
			if err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}
