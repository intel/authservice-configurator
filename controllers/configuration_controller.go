/*


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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	authcontroller "intel.com/authservice-webhook/api/v1"
)

// ConfigurationReconciler reconciles a Configuration object
type ConfigurationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=authcontroller.intel.com,resources=configurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=authcontroller.intel.com,resources=configurations/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;create;update;patch;watch
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;create;update;patch;watch
// +kubebuilder:rbac:groups=security.istio.io,resources=requestauthentications,verbs=get;list;create;update;patch;watch

// Reconcile creates/updates the AuthService configuration when the Configuration is modified.
func (r *ConfigurationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("configuration", req.NamespacedName)

	configuration, chains, err := getConfigOptions(r, r.Log, req.NamespacedName.Name, req.NamespacedName.Namespace)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Generate the ConfigMap based on the configuration and the chains.
	configMap, update := createConfigMap(r, configuration, chains)

	// TODO: switch to CreateOrUpdate
	// ctrl.CreateOrUpdate(ctx, r, configMap, func() error {
	// 	return nil
	// })

	// Create/Update the existing ConfigMap if it exists with the new JSON file.
	if update {
		if err := r.Update(ctx, configMap); err != nil {
			_ = r.Log.WithValues("Failed to update ConfigMap for authservice", req.NamespacedName)
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	} else {
		if err := r.Create(ctx, configMap); err != nil {
			_ = r.Log.WithValues("Failed to create ConfigMap for authservice", req.NamespacedName)
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	}

	for _, chain := range chains.Items {
		if err := createRequestAuthentication(r, r.Log, &chain); err != nil {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	}

	if err := restartAuthService(r, r.Log, configuration.Spec.AuthService, req.Namespace); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager connects the controller with the manager.
func (r *ConfigurationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&authcontroller.Configuration{}).
		Complete(r)
}
