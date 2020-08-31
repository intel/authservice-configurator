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

package v1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var configurationlog = logf.Log.WithName("configuration-resource")

// SetupWebhookWithManager connects the webhook with controller manager.
func (r *Configuration) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-authcontroller-intel-com-v1-configuration,mutating=false,failurePolicy=fail,groups=authcontroller.intel.com,resources=configurations,versions=v1,name=vconfiguration.kb.io

var _ webhook.Validator = &Configuration{}

func (r *Configuration) validateConfiguration() error {
	if r.Spec.Threads < 1 || r.Spec.Threads > 1024 {
		return fmt.Errorf("Invalid thread number")
	}

	return nil
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Configuration) ValidateCreate() error {
	configurationlog.Info("validate create", "name", r.Name)

	return r.validateConfiguration()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Configuration) ValidateUpdate(old runtime.Object) error {
	configurationlog.Info("validate update", "name", r.Name)

	return r.validateConfiguration()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Configuration) ValidateDelete() error {
	configurationlog.Info("validate delete", "name", r.Name)

	return nil
}
