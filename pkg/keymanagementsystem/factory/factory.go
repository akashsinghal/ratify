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
	"fmt"

	"github.com/deislabs/ratify/pkg/keymanagementsystem"
	"github.com/deislabs/ratify/pkg/keymanagementsystem/config"
	"github.com/deislabs/ratify/pkg/keymanagementsystem/types"
)

// map of key management system names to key management system factories
var builtInKeyManagementSystems = make(map[string]KeyManagementSystemFactory)

// KeyManagementSystemFactory is an interface for creating key management system providers
type KeyManagementSystemFactory interface {
	Create(version string, keyManagementSystemConfig config.KeyManagementSystemConfig, pluginDirectory string) (keymanagementsystem.KeyManagementSystemProvider, error)
}

// Register registers a key management system factory by name
func Register(name string, factory KeyManagementSystemFactory) {
	if factory == nil {
		panic("key management system factory cannot be nil")
	}
	_, registered := builtInKeyManagementSystems[name]
	if registered {
		panic(fmt.Sprintf("key management system factory named %s already registered", name))
	}

	builtInKeyManagementSystems[name] = factory
}

// CreateKeyManagementSystemFromConfig creates a key management system provider from config
func CreateKeyManagementSystemFromConfig(keyManagementSystemConfig config.KeyManagementSystemConfig, configVersion string, pluginDirectory string) (keymanagementsystem.KeyManagementSystemProvider, error) {
	keyManagementSystemProvider, ok := keyManagementSystemConfig[types.Type]
	if !ok {
		return nil, fmt.Errorf("failed to find key management system name in the certificate stores config with key %s", types.Type)
	}

	keyManagementSystemProviderStr := fmt.Sprintf("%s", keyManagementSystemProvider)
	if keyManagementSystemProviderStr == "" {
		return nil, fmt.Errorf("key management system type cannot be empty")
	}

	factory, ok := builtInKeyManagementSystems[keyManagementSystemProviderStr]
	if !ok {
		return nil, fmt.Errorf("key management system factory with name %s not found", keyManagementSystemProviderStr)
	}

	return factory.Create(configVersion, keyManagementSystemConfig, pluginDirectory)
}
