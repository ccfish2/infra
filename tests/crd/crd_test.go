/*
Copyright 2025 The Kubernetes Authors.

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

package crd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

// TestCRDValidation starts an envtest control plane and verifies that the
// generated Dolphin CRDs are installed and discoverable.
func TestCRDValidation(t *testing.T) {
	testEnv := &envtest.Environment{
		ErrorIfCRDPathMissing: true,
		// Set KUBEBUILDER_ASSETS to the directory containing etcd and kube-apiserver.
		BinaryAssetsDirectory: os.Getenv("KUBEBUILDER_ASSETS"),
		CRDInstallOptions: envtest.CRDInstallOptions{
			Paths: []string{
				filepath.Join("..", "..", "pkg", "k8s", "apis", "dolphin.io", "client", "crd", "bases"),
			},
			CleanUpAfterUse: true,
		},
	}

	_, err := testEnv.Start()
	require.NoError(t, err, "failed to initialize envtest")
	t.Cleanup(func() {
		require.NoError(t, testEnv.Stop())
	})

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(testEnv.Config)
	require.NoError(t, err)

	apiResources, err := discoveryClient.ServerResourcesForGroupVersion("dolphin.io/v1")
	require.NoError(t, err)

	resourceNames := make([]string, 0, len(apiResources.APIResources))
	for _, resource := range apiResources.APIResources {
		resourceNames = append(resourceNames, resource.Name)
	}
	require.Subset(t, resourceNames, []string{
		"dolphinendpoints",
		"dolphinendpointslices",
		"dolphinenvoyconfigs",
		"dolphinidentities",
		"dolphinnodes",
	}, resourceNames)
}
