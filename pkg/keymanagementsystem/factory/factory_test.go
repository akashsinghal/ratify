/*
Copyright The Ratify Authors.
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

package factory

import (
	"testing"

	"github.com/deislabs/ratify/pkg/keymanagementsystem"
	"github.com/deislabs/ratify/pkg/keymanagementsystem/config"
	"github.com/deislabs/ratify/pkg/keymanagementsystem/mocks"
)

type TestKeyManagementSystemProviderFactory struct{}

func (f TestKeyManagementSystemProviderFactory) Create(_ string, _ config.KeyManagementSystemConfig, _ string) (keymanagementsystem.KeyManagementSystemProvider, error) {
	return &mocks.TestKeyManagementSystemProvider{}, nil
}

// TestCreatePolicyProvidersFromConfig_BuiltInPolicyProviders_ReturnsExpected checks the correct registered policy provider is invoked based on config
func TestCreatePolicyProvidersFromConfig_BuiltInPolicyProviders_ReturnsExpected(t *testing.T) {
	builtInKeyManagementSystems = map[string]KeyManagementSystemFactory{
		"test-kmsprovider": TestKeyManagementSystemProviderFactory{},
	}

	config := config.KeyManagementSystemConfig{
		"type": "test-kmsprovider",
	}

	_, err := CreateKeyManagementSystemFromConfig(config, "", "")
	if err != nil {
		t.Fatalf("create key management system provider should not have failed: %v", err)
	}
}

// TestCreatePolicyProvidersFromConfig_NonexistentPolicyProviders_ReturnsExpected checks the auth provider creation fails if auth provider specified does not exist
func TestCreatePolicyProvidersFromConfig_NonexistentPolicyProviders_ReturnsExpected(t *testing.T) {
	builtInKeyManagementSystems = map[string]KeyManagementSystemFactory{
		"testkeymanagementsystemprovider": TestKeyManagementSystemProviderFactory{},
	}

	config := config.KeyManagementSystemConfig{
		"type": "test-nonexistent",
	}

	_, err := CreateKeyManagementSystemFromConfig(config, "", "")
	if err == nil {
		t.Fatal("create key management system provider should have failed")
	}
}
