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
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/url"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var chainlog = logf.Log.WithName("chain-resource")

// SetupWebhookWithManager connects the webhook with controller manager.
func (r *Chain) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

func (r *Chain) validateChain() error {
	// AuthorizationURI is required
	u, err := url.ParseRequestURI(r.Spec.AuthorizationURI)
	if err != nil {
		return err
	}
	if u.Scheme != "https" {
		return fmt.Errorf("URI isn't HTTPs")
	}

	// CallbackURI is required
	u, err = url.ParseRequestURI(r.Spec.CallbackURI)
	if err != nil {
		return err
	}
	if u.Scheme != "https" {
		return fmt.Errorf("URI isn't HTTPs")
	}

	// TokenURI is required
	u, err = url.ParseRequestURI(r.Spec.TokenURI)
	if err != nil {
		return err
	}
	if u.Scheme != "https" {
		return fmt.Errorf("URI isn't HTTPs")
	}

	// ClientID is required
	if r.Spec.ClientID == "" {
		return fmt.Errorf("ClientID is required")
	}

	// ClientSecret is required
	if r.Spec.ClientSecret == "" {
		return fmt.Errorf("ClientSecret is required")
	}

	// Validate JWKS with https://tools.ietf.org/html/rfc7517

	jwksBytes := []byte(r.Spec.Jwks)

	if !json.Valid(jwksBytes) {
		return fmt.Errorf("Invalid JWKS data")
	}

	var jwksMap map[string]interface{}

	err = json.Unmarshal(jwksBytes, &jwksMap)
	if err != nil {
		return fmt.Errorf("Invalid JWKS data")
	}

	// a key set must contain "keys" member
	keysInterface, found := jwksMap["keys"]
	if !found {
		return fmt.Errorf("Invalid JWKS data")
	}
	if keys, ok := keysInterface.([]interface{}); ok {
		for _, jwkInterface := range keys {
			if jwk, ok := jwkInterface.(map[string]interface{}); ok {
				// "kty" is a mandatory value for each key
				_, found := jwk["kty"]
				if !found {
					return fmt.Errorf("Invalid JWK data")
				}
			}
		}
	}

	if r.Spec.Match.Criteria != "" && r.Spec.Match.Criteria != "prefix" && r.Spec.Match.Criteria != "equality" {
		return fmt.Errorf("If criteria is set, it needs to be \"prefix\" or \"equality\"")
	}

	// RequestAuthentication can be created iff both JwksURI and Issuer exist

	if (r.Spec.JwksURI != "" && r.Spec.Issuer == "") || (r.Spec.JwksURI == "" && r.Spec.Issuer != "") {
		return fmt.Errorf("Both issuer and jwksUri need to be set if one of them is set")
	}

	if r.Spec.JwksURI != "" {
		_, err := url.ParseRequestURI(r.Spec.AuthorizationURI)
		if err != nil {
			return err
		}
	}

	if r.Spec.TrustedCertificateAuthority != "" {
		// TrustedCertificateAuthroity is optional, but we check the value if it's present
		parsed, _ := pem.Decode([]byte(r.Spec.TrustedCertificateAuthority))
		if parsed == nil {
			return fmt.Errorf("Invalid certificate")
		}
	}

	return nil
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-authcontroller-intel-com-v1-chain,mutating=false,failurePolicy=fail,groups=authcontroller.intel.com,resources=chains,versions=v1,name=vchain.kb.io

var _ webhook.Validator = &Chain{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Chain) ValidateCreate() error {
	chainlog.Info("validate create", "name", r.Name)

	return r.validateChain()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Chain) ValidateUpdate(old runtime.Object) error {
	chainlog.Info("validate update", "name", r.Name)

	return r.validateChain()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Chain) ValidateDelete() error {
	chainlog.Info("validate delete", "name", r.Name)

	return nil
}
