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

package controllers

import (
	"fmt"
	"reflect"
	"testing"

	configv1beta1 "github.com/deislabs/ratify/api/v1beta1"
	"github.com/deislabs/ratify/pkg/keymanagementsystem"
	"github.com/deislabs/ratify/pkg/keymanagementsystem/config"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// TestUpdateErrorStatus tests the updateErrorStatus method
func TestKMSUpdateErrorStatus(t *testing.T) {
	var parametersString = "{\"certs\":{\"name\":\"certName\"}}"
	var kmsStatus = []byte(parametersString)

	status := configv1beta1.KeyManagementSystemStatus{
		IsSuccess: true,
		Properties: runtime.RawExtension{
			Raw: kmsStatus,
		},
	}
	keyManagementSystem := configv1beta1.KeyManagementSystem{
		Status: status,
	}
	expectedErr := "it's a long error from unit test"
	lastFetchedTime := metav1.Now()
	updateKMSErrorStatus(&keyManagementSystem, expectedErr, &lastFetchedTime)

	if keyManagementSystem.Status.IsSuccess != false {
		t.Fatalf("Unexpected error, expected isSuccess to be false , actual %+v", keyManagementSystem.Status.IsSuccess)
	}

	if keyManagementSystem.Status.Error != expectedErr {
		t.Fatalf("Unexpected error string, expected %+v, got %+v", expectedErr, keyManagementSystem.Status.Error)
	}
	expectedBriedErr := fmt.Sprintf("%s...", expectedErr[:30])
	if keyManagementSystem.Status.BriefError != expectedBriedErr {
		t.Fatalf("Unexpected error string, expected %+v, got %+v", expectedBriedErr, keyManagementSystem.Status.Error)
	}

	//make sure properties of last cached cert was not overridden
	if len(keyManagementSystem.Status.Properties.Raw) == 0 {
		t.Fatalf("Unexpected properties,  expected %+v, got %+v", parametersString, string(keyManagementSystem.Status.Properties.Raw))
	}
}

// TestKMSUpdateSuccessStatus tests the updateSuccessStatus method
func TestKMSUpdateSuccessStatus(t *testing.T) {
	kmsStatus := keymanagementsystem.KeyManagementSystemStatus{}
	properties := map[string]string{}
	properties["CertName"] = "wabbit"
	properties["Version"] = "ABC"

	kmsStatus["Certificates"] = properties

	lastFetchedTime := metav1.Now()

	status := configv1beta1.KeyManagementSystemStatus{
		IsSuccess: false,
		Error:     "error from last operation",
	}
	keyManagementSystem := configv1beta1.KeyManagementSystem{
		Status: status,
	}

	updateKMSSuccessStatus(&keyManagementSystem, &lastFetchedTime, kmsStatus)

	if keyManagementSystem.Status.IsSuccess != true {
		t.Fatalf("Expected isSuccess to be true , actual %+v", keyManagementSystem.Status.IsSuccess)
	}

	if keyManagementSystem.Status.Error != "" {
		t.Fatalf("Unexpected error string, actual %+v", keyManagementSystem.Status.Error)
	}

	//make sure properties of last cached cert was updated
	if len(keyManagementSystem.Status.Properties.Raw) == 0 {
		t.Fatalf("Properties should not be empty")
	}
}

// TestKMSUpdateSuccessStatus tests the updateSuccessStatus method with empty properties
func TestKMSUpdateSuccessStatus_emptyProperties(t *testing.T) {
	lastFetchedTime := metav1.Now()
	status := configv1beta1.KeyManagementSystemStatus{
		IsSuccess: false,
		Error:     "error from last operation",
	}
	keyManagementSystem := configv1beta1.KeyManagementSystem{
		Status: status,
	}

	updateKMSSuccessStatus(&keyManagementSystem, &lastFetchedTime, nil)

	if keyManagementSystem.Status.IsSuccess != true {
		t.Fatalf("Expected isSuccess to be true , actual %+v", keyManagementSystem.Status.IsSuccess)
	}

	if keyManagementSystem.Status.Error != "" {
		t.Fatalf("Unexpected error string, actual %+v", keyManagementSystem.Status.Error)
	}

	//make sure properties of last cached cert was updated
	if len(keyManagementSystem.Status.Properties.Raw) != 0 {
		t.Fatalf("Properties should be empty")
	}
}

// TestRawToKeyManagementSystemConfig tests the rawToKeyManagementSystemConfig method
func TestRawToKeyManagementSystemConfig(t *testing.T) {
	testCases := []struct {
		name         string
		raw          []byte
		expectErr    bool
		expectConfig config.KeyManagementSystemConfig
	}{
		{
			name:         "empty Raw",
			raw:          []byte{},
			expectErr:    true,
			expectConfig: config.KeyManagementSystemConfig{},
		},
		{
			name:         "unmarshal failure",
			raw:          []byte("invalid"),
			expectErr:    true,
			expectConfig: config.KeyManagementSystemConfig{},
		},
		{
			name:      "valid Raw",
			raw:       []byte("{\"type\": \"inline\"}"),
			expectErr: false,
			expectConfig: config.KeyManagementSystemConfig{
				"type": "inline",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config, err := rawToKeyManagementSystemConfig(tc.raw, "inline")

			if tc.expectErr != (err != nil) {
				t.Fatalf("Expected error to be %t, got %t", tc.expectErr, err != nil)
			}
			if !reflect.DeepEqual(config, tc.expectConfig) {
				t.Fatalf("Expected config to be %v, got %v", tc.expectConfig, config)
			}
		})
	}
}

// TestSpecToKeyManagementSystemProvider tests the specToKeyManagementSystemProvider method
func TestSpecToKeyManagementSystemProvider(t *testing.T) {
	testCases := []struct {
		name      string
		spec      configv1beta1.KeyManagementSystemSpec
		expectErr bool
	}{
		{
			name:      "empty spec",
			spec:      configv1beta1.KeyManagementSystemSpec{},
			expectErr: true,
		},
		{
			name: "missing inline provider required fields",
			spec: configv1beta1.KeyManagementSystemSpec{
				Type: "inline",
				Parameters: runtime.RawExtension{
					Raw: []byte("{\"type\": \"inline\"}"),
				},
			},
			expectErr: true,
		},
		{
			name: "valid spec",
			spec: configv1beta1.KeyManagementSystemSpec{
				Type: "inline",
				Parameters: runtime.RawExtension{
					Raw: []byte(`{"type": "inline", "contentType": "certificate", "value": "-----BEGIN CERTIFICATE-----\nMIID2jCCAsKgAwIBAgIQXy2VqtlhSkiZKAGhsnkjbDANBgkqhkiG9w0BAQsFADBvMRswGQYDVQQD\nExJyYXRpZnkuZXhhbXBsZS5jb20xDzANBgNVBAsTBk15IE9yZzETMBEGA1UEChMKTXkgQ29tcGFu\neTEQMA4GA1UEBxMHUmVkbW9uZDELMAkGA1UECBMCV0ExCzAJBgNVBAYTAlVTMB4XDTIzMDIwMTIy\nNDUwMFoXDTI0MDIwMTIyNTUwMFowbzEbMBkGA1UEAxMScmF0aWZ5LmV4YW1wbGUuY29tMQ8wDQYD\nVQQLEwZNeSBPcmcxEzARBgNVBAoTCk15IENvbXBhbnkxEDAOBgNVBAcTB1JlZG1vbmQxCzAJBgNV\nBAgTAldBMQswCQYDVQQGEwJVUzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAL10bM81\npPAyuraORABsOGS8M76Bi7Guwa3JlM1g2D8CuzSfSTaaT6apy9GsccxUvXd5cmiP1ffna5z+EFmc\nizFQh2aq9kWKWXDvKFXzpQuhyqD1HeVlRlF+V0AfZPvGt3VwUUjNycoUU44ctCWmcUQP/KShZev3\n6SOsJ9q7KLjxxQLsUc4mg55eZUThu8mGB8jugtjsnLUYvIWfHhyjVpGrGVrdkDMoMn+u33scOmrt\nsBljvq9WVo4T/VrTDuiOYlAJFMUae2Ptvo0go8XTN3OjLblKeiK4C+jMn9Dk33oGIT9pmX0vrDJV\nX56w/2SejC1AxCPchHaMuhlwMpftBGkCAwEAAaNyMHAwDgYDVR0PAQH/BAQDAgeAMAkGA1UdEwQC\nMAAwEwYDVR0lBAwwCgYIKwYBBQUHAwMwHwYDVR0jBBgwFoAU0eaKkZj+MS9jCp9Dg1zdv3v/aKww\nHQYDVR0OBBYEFNHmipGY/jEvYwqfQ4Nc3b97/2isMA0GCSqGSIb3DQEBCwUAA4IBAQBNDcmSBizF\nmpJlD8EgNcUCy5tz7W3+AAhEbA3vsHP4D/UyV3UgcESx+L+Nye5uDYtTVm3lQejs3erN2BjW+ds+\nXFnpU/pVimd0aYv6mJfOieRILBF4XFomjhrJOLI55oVwLN/AgX6kuC3CJY2NMyJKlTao9oZgpHhs\nLlxB/r0n9JnUoN0Gq93oc1+OLFjPI7gNuPXYOP1N46oKgEmAEmNkP1etFrEjFRgsdIFHksrmlOlD\nIed9RcQ087VLjmuymLgqMTFX34Q3j7XgN2ENwBSnkHotE9CcuGRW+NuiOeJalL8DBmFXXWwHTKLQ\nPp5g6m1yZXylLJaFLKz7tdMmO355\n-----END CERTIFICATE-----\n"}`),
				},
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := specToKeyManagementSystemProvider(tc.spec)
			if tc.expectErr != (err != nil) {
				t.Fatalf("Expected error to be %t, got %t", tc.expectErr, err != nil)
			}
		})
	}
}