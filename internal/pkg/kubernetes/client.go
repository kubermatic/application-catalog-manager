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

package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func PatchObject(ctx context.Context, client ctrlruntimeclient.Client, obj ctrlruntimeclient.Object, modify func()) error {
	if modify == nil {
		return nil
	}

	oldObj := obj.DeepCopyObject().(ctrlruntimeclient.Object)
	modify()
	return client.Patch(ctx, obj, ctrlruntimeclient.MergeFrom(oldObj))
}

// GetCredentialFromSecret get the secret and returns secret.Data[key].
func GetCredentialFromSecret(ctx context.Context, client ctrlruntimeclient.Client, namespce string, name string, key string) (string, error) {
	secret := &corev1.Secret{}
	if err := client.Get(ctx, types.NamespacedName{Namespace: namespce, Name: name}, secret); err != nil {
		return "", fmt.Errorf("failed to get credential secret: %w", err)
	}

	cred, found := secret.Data[key]
	if !found {
		return "", fmt.Errorf("key '%s' does not exist in secret '%s'", key, fmt.Sprintf("%s/%s", secret.GetNamespace(), secret.GetName()))
	}
	return string(cred), nil
}
