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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ChainMatch has the request match criteria.
type ChainMatch struct {
	Header   string `json:"header,omitempty"`
	Criteria string `json:"criteria,omitempty"`
	Prefix   string `json:"prefix,omitempty"`
	Equality string `json:"equality,omitempty"`
}

// The problem with Match struct is that the values need to be in sync with the
// actual routing data. For example, if an service gets all the traffic going to
// path /foo, the same /foo statement needs to be replicated here (if we are in
// a multi-tenant environment). It would be much better if the ChainMatch could
// be automatically generated from K8s Ingress or Istio Ingress Gateway resources.
// It's also not clear if the AuthService Match can be configured to fully cover
// all possible Ingress matching scenarios.

// ChainSpec defines the desired state of Chain
type ChainSpec struct {
	// Desired state of cluster
	Match                       ChainMatch `json:"match,omitempty"`
	AuthorizationURI            string     `json:"authorizationUri,omitempty"`
	TokenURI                    string     `json:"tokenUri,omitempty"`
	CallbackURI                 string     `json:"callbackUri,omitempty"`
	ClientID                    string     `json:"clientId,omitempty"`
	Jwks                        string     `json:"jwks,omitempty"`         // Contents of JwksURI (escaped)
	ClientSecret                string     `json:"clientSecret,omitempty"` // TODO: store this in Kubernetes Secret?
	TrustedCertificateAuthority string     `json:"trustedCertificateAuthority,omitempty"`
	CookieNamePrefix            string     `json:"cookieNamePrefix,omitempty"`
	Issuer                      string     `json:"issuer,omitempty"`  // For Istio RequestAuthentication
	JwksURI                     string     `json:"jwksUri,omitempty"` // For Istio RequestAuthentication
}

// ChainStatus defines the observed state of Chain
type ChainStatus struct {
	// Observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// Chain is the Schema for the chains API
type Chain struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ChainSpec   `json:"spec,omitempty"`
	Status ChainStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ChainList contains a list of Chain
type ChainList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Chain `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Chain{}, &ChainList{})
}
