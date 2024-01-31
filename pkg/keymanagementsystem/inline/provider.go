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

package inline

import (
	"context"
	"crypto/x509"
	"encoding/json"

	"github.com/deislabs/ratify/errors"
	"github.com/deislabs/ratify/pkg/keymanagementsystem"
	"github.com/deislabs/ratify/pkg/keymanagementsystem/config"
	"github.com/deislabs/ratify/pkg/keymanagementsystem/factory"
)

const (
	// ValueParameter is the name of the parameter that contains the certificate (chain) as a string in PEM format
	ValueParameter                = "value"
	providerName           string = "inline"
	certificateContentType string = "certificate"
)

type InlineKMSConfig struct {
	Type        string `json:"type"`
	ContentType string `json:"contentType"`
	Value       string `json:"value"`
}

type inlineKMSProvider struct {
	certs       []*x509.Certificate
	contentType string
}
type inlineKMSFactory struct{}

// init calls to register the provider
func init() {
	factory.Register(providerName, &inlineKMSFactory{})
}

// Create creates a new instance of the inline key management system provider
// checks contentType is set to 'certificate' and value is set to a valid certificate
func (f *inlineKMSFactory) Create(_ string, keyManagementSystemConfig config.KeyManagementSystemConfig, _ string) (keymanagementsystem.KeyManagementSystemProvider, error) {
	conf := InlineKMSConfig{}

	keyManagementSystemConfigBytes, err := json.Marshal(keyManagementSystemConfig)
	if err != nil {
		return nil, errors.ErrorCodeConfigInvalid.WithError(err).WithComponentType(errors.KeyManagementSystemProvider)
	}

	if err := json.Unmarshal(keyManagementSystemConfigBytes, &conf); err != nil {
		return nil, errors.ErrorCodeConfigInvalid.NewError(errors.KeyManagementSystemProvider, "", errors.EmptyLink, err, "failed to parse AKV key management system configuration", errors.HideStackTrace)
	}

	if conf.ContentType == "" {
		return nil, errors.ErrorCodeConfigInvalid.WithComponentType(errors.KeyManagementSystemProvider).WithDetail("contentType parameter is not set")
	}

	// only support certificate content type for now
	if conf.ContentType != certificateContentType {
		return nil, errors.ErrorCodeConfigInvalid.WithComponentType(errors.KeyManagementSystemProvider).WithDetail("contentType parameter is not set to 'certificate'")
	}

	if conf.Value == "" {
		return nil, errors.ErrorCodeConfigInvalid.WithComponentType(errors.KeyManagementSystemProvider).WithDetail("value parameter is not set")
	}

	certs, err := keymanagementsystem.DecodeCertificates([]byte(conf.Value))
	if err != nil {
		return nil, errors.ErrorCodeCertInvalid.WithComponentType(errors.KeyManagementSystemProvider)
	}

	return &inlineKMSProvider{certs: certs, contentType: conf.ContentType}, nil
}

// GetCertificates returns previously fetched certificates
func (s *inlineKMSProvider) GetCertificates(_ context.Context) ([]*x509.Certificate, keymanagementsystem.KeyManagementSystemStatus, error) {
	return s.certs, nil, nil
}
