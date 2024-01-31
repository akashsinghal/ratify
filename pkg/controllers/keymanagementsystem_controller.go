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
	"context"
	"encoding/json"
	"fmt"

	_ "github.com/deislabs/ratify/pkg/keymanagementsystem/azurekeyvault" // register azure key vault certificate provider
	_ "github.com/deislabs/ratify/pkg/keymanagementsystem/inline"        // register inline certificate provider
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	configv1beta1 "github.com/deislabs/ratify/api/v1beta1"
	c "github.com/deislabs/ratify/config"
	"github.com/deislabs/ratify/pkg/keymanagementsystem"
	"github.com/deislabs/ratify/pkg/keymanagementsystem/config"
	"github.com/deislabs/ratify/pkg/keymanagementsystem/factory"
	"github.com/deislabs/ratify/pkg/keymanagementsystem/types"
	"github.com/sirupsen/logrus"
)

// KeyManagementSystemReconciler reconciles a KeyManagementSystem object
type KeyManagementSystemReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=keymanagementsystems,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=keymanagementsystems/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=config.ratify.deislabs.io,resources=keymanagementsystems/finalizers,verbs=update
func (r *KeyManagementSystemReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logrus.WithContext(ctx)

	var resource = req.NamespacedName.String()
	var keyManagementSystem configv1beta1.KeyManagementSystem

	logger.Infof("reconciling key management system '%v'", resource)

	if err := r.Get(ctx, req.NamespacedName, &keyManagementSystem); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("deletion detected, removing key management system %v", resource)
			keymanagementsystem.DeleteCertificatesFromMap(resource)
		} else {
			logger.Error(err, "unable to fetch key management system")
		}

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	lastFetchedTime := metav1.Now()
	isFetchSuccessful := false

	// get certificate store list to check if certificate store is configured
	var certificateStoreList configv1beta1.CertificateStoreList
	if err := r.List(ctx, &certificateStoreList); err != nil {
		logger.Error(err, "unable to list certificate stores")
		return ctrl.Result{}, err
	}
	// if certificate store is configured, return error. Only one of certificate store and key management system can be configured
	if len(certificateStoreList.Items) > 0 {
		err := fmt.Errorf("certificate store already exists: key management system and certificate store cannot be configured together")
		logger.Error(err)
		writeKMSStatus(ctx, r, keyManagementSystem, logger, isFetchSuccessful, err.Error(), lastFetchedTime, nil)
		return ctrl.Result{}, err
	}

	provider, err := specToKeyManagementSystemProvider(keyManagementSystem.Spec)
	if err != nil {
		writeKMSStatus(ctx, r, keyManagementSystem, logger, isFetchSuccessful, err.Error(), lastFetchedTime, nil)
		return ctrl.Result{}, err
	}

	certificates, certAttributes, err := provider.GetCertificates(ctx)
	if err != nil {
		writeKMSStatus(ctx, r, keyManagementSystem, logger, isFetchSuccessful, err.Error(), lastFetchedTime, nil)
		return ctrl.Result{}, fmt.Errorf("Error fetching certificates in KMS %v with %v provider, error: %w", resource, keyManagementSystem.Spec.Type, err)
	}
	keymanagementsystem.SetCertificatesInMap(resource, certificates)
	isFetchSuccessful = true
	emptyErrorString := ""
	writeKMSStatus(ctx, r, keyManagementSystem, logger, isFetchSuccessful, emptyErrorString, lastFetchedTime, certAttributes)

	logger.Infof("%v certificates fetched for key management system %v", len(certificates), resource)

	// returning empty result and no error to indicate we’ve successfully reconciled this object
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KeyManagementSystemReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.GenerationChangedPredicate{}

	// status updates will trigger a reconcile event
	// if there are no changes to spec of CRD, this event should be filtered out by using the predicate
	// see more discussions at https://github.com/kubernetes-sigs/kubebuilder/issues/618
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1beta1.KeyManagementSystem{}).WithEventFilter(pred).
		Complete(r)
}

// specToKeyManagementSystemProvider creates KeyManagementSystemProvider from  KeyManagementSystemSpec config
func specToKeyManagementSystemProvider(spec configv1beta1.KeyManagementSystemSpec) (keymanagementsystem.KeyManagementSystemProvider, error) {
	kmsConfig, err := rawToKeyManagementSystemConfig(spec.Parameters.Raw, spec.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to parse key management system config: %w", err)
	}

	// TODO: add Version and Address to KeyManagementSystemSpec
	keyManagementSystemProvider, err := factory.CreateKeyManagementSystemFromConfig(kmsConfig, "0.1.0", c.GetDefaultPluginPath())
	if err != nil {
		return nil, fmt.Errorf("failed to create key management system provider: %w", err)
	}

	return keyManagementSystemProvider, nil
}

// rawToKeyManagementSystemConfig converts raw json to KeyManagementSystemConfig
func rawToKeyManagementSystemConfig(raw []byte, keyManagamentSystemName string) (config.KeyManagementSystemConfig, error) {
	pluginConfig := config.KeyManagementSystemConfig{}

	if string(raw) == "" {
		return config.KeyManagementSystemConfig{}, fmt.Errorf("no key management system parameters provided")
	}
	if err := json.Unmarshal(raw, &pluginConfig); err != nil {
		return config.KeyManagementSystemConfig{}, fmt.Errorf("unable to decode key management system parameters.Raw: %s, err: %w", raw, err)
	}

	pluginConfig[types.Type] = keyManagamentSystemName

	return pluginConfig, nil
}

// writeKMSStatus updates the status of the key management system resource
func writeKMSStatus(ctx context.Context, r *KeyManagementSystemReconciler, keyManagementSystem configv1beta1.KeyManagementSystem, logger *logrus.Entry, isSuccess bool, errorString string, operationTime metav1.Time, kmsStatus keymanagementsystem.KeyManagementSystemStatus) {
	if isSuccess {
		updateKMSSuccessStatus(&keyManagementSystem, &operationTime, kmsStatus)
	} else {
		updateKMSErrorStatus(&keyManagementSystem, errorString, &operationTime)
	}
	if statusErr := r.Status().Update(ctx, &keyManagementSystem); statusErr != nil {
		logger.Error(statusErr, ",unable to update key management system error status")
	}
}

// updateKMSErrorStatus updates the key management system status with error, brief error and last fetched time
func updateKMSErrorStatus(keyManagementSystem *configv1beta1.KeyManagementSystem, errorString string, operationTime *metav1.Time) {
	// truncate brief error string to maxBriefErrLength
	briefErr := errorString
	if len(errorString) > maxBriefErrLength {
		briefErr = fmt.Sprintf("%s...", errorString[:maxBriefErrLength])
	}
	keyManagementSystem.Status.IsSuccess = false
	keyManagementSystem.Status.Error = errorString
	keyManagementSystem.Status.BriefError = briefErr
	keyManagementSystem.Status.LastFetchedTime = operationTime
}

// updateKMSSuccessStatus updates the key management system status if status argument is non nil
// Success status includes last fetched time and other provider-specific properties
func updateKMSSuccessStatus(keyManagementSystem *configv1beta1.KeyManagementSystem, lastOperationTime *metav1.Time, kmsStatus keymanagementsystem.KeyManagementSystemStatus) {
	keyManagementSystem.Status.IsSuccess = true
	keyManagementSystem.Status.Error = ""
	keyManagementSystem.Status.BriefError = ""
	keyManagementSystem.Status.LastFetchedTime = lastOperationTime

	if kmsStatus != nil {
		jsonString, _ := json.Marshal(kmsStatus)

		raw := runtime.RawExtension{
			Raw: jsonString,
		}
		keyManagementSystem.Status.Properties = raw
	}
}
