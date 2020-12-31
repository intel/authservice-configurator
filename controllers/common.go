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
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	authcontroller "intel.com/authservice-webhook/api/v1"
	"istio.io/api/security/v1beta1"
	istiosecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apisv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// JSONMatch is part of the ConfigMap data.
type JSONMatch struct {
	Header   string `json:"header,omitempty"`
	Criteria string `json:"criteria,omitempty"`
	Prefix   string `json:"prefix,omitempty"`
	Equality string `json:"equality,omitempty"`
}

// JSONToken is part of the ConfigMap data.
type JSONToken struct {
	Preamble string `json:"preamble,omitempty"`
	Header   string `json:"header,omitempty"`
}

// JSONOidc is part of the ConfigMap data.
type JSONOidc struct {
	AuthorizationURI            string     `json:"authorization_uri"`
	TokenURI                    string     `json:"token_uri"`
	CallbackURI                 string     `json:"callback_uri"`
	ClientID                    string     `json:"client_id"`
	Jwks                        string     `json:"jwks"`
	ClientSecret                string     `json:"client_secret"`
	TrustedCertificateAuthority string     `json:"trusted_certificate_authority,omitempty"`
	Scopes                      []string   `json:"scopes"`
	CookieNamePrefix            string     `json:"cookie_name_prefix,omitempty"`
	IDToken                     *JSONToken `json:"id_token"`
	AccessToken                 *JSONToken `json:"access_token,omitempty"`
}

// JSONFilter is part of the ConfigMap data.
type JSONFilter struct {
	Oidc *JSONOidc `json:"oidc,omitempty"`
}

// JSONChain is part of the ConfigMap data.
type JSONChain struct {
	Name    string       `json:"name,omitempty"`
	Match   *JSONMatch   `json:"match,omitempty"`
	Filters []JSONFilter `json:"filters,omitempty"`
}

// JSONConfigData is part of the ConfigMap data.
type JSONConfigData struct {
	ListenAddress string       `json:"listen_address,omitempty"`
	ListenPort    string       `json:"listen_port,omitempty"`
	LogLevel      string       `json:"log_level,omitempty"`
	Threads       int          `json:"threads,omitempty"`
	Chains        []*JSONChain `json:"chains,omitempty"`
}

func createAuthserviceConfiguration(threads int, chains *authcontroller.ChainList) *JSONConfigData {
	configData := JSONConfigData{
		ListenAddress: "0.0.0.0",
		ListenPort:    "10003",
		LogLevel:      "trace",
		Threads:       threads,
		Chains:        make([]*JSONChain, len(chains.Items)),
	}

	for i, chain := range chains.Items {
		configData.Chains[i] = &JSONChain{
			Name: chain.Name,
			Filters: []JSONFilter{
				{
					Oidc: &JSONOidc{
						AuthorizationURI:            chain.Spec.AuthorizationURI,
						TokenURI:                    chain.Spec.TokenURI,
						CallbackURI:                 chain.Spec.CallbackURI,
						ClientID:                    chain.Spec.ClientID,
						ClientSecret:                chain.Spec.ClientSecret,
						TrustedCertificateAuthority: chain.Spec.TrustedCertificateAuthority,
						Jwks:                        chain.Spec.Jwks,
						Scopes:                      []string{},
						CookieNamePrefix:            chain.Spec.CookieNamePrefix,
						IDToken: &JSONToken{
							Preamble: "Bearer",
							Header:   "Authorization",
						},
						AccessToken: &JSONToken{
							Preamble: "Bearer",
							Header:   "Authorization",
						},
					},
				},
			},
		}
		if chain.Spec.Match.Header == "" && chain.Spec.Match.Criteria == "" && chain.Spec.Match.Prefix == "" && chain.Spec.Match.Equality == "" {
			configData.Chains[i].Match = nil
		} else {
			configData.Chains[i].Match = &JSONMatch{
				Header:   chain.Spec.Match.Header,
				Criteria: chain.Spec.Match.Criteria,
				Prefix:   chain.Spec.Match.Prefix,
				Equality: chain.Spec.Match.Equality,
			}
		}
	}

	return &configData
}

func createConfigMap(client client.Client, namespace string, threads int, chains *authcontroller.ChainList) (*corev1.ConfigMap, bool) {
	var configMap corev1.ConfigMap
	ctx := context.Background()
	update := true

	// Create the ConfigMap to the same namespace where the related chains are. This is for limiting
	// the configuration of AuthService from resources in unrelated namespaces. See
	// https://kubernetes.io/docs/tasks/administer-cluster/securing-a-cluster/#api-authorization

	configMapName := types.NamespacedName{
		Namespace: namespace,
		Name:      "authservice-configmap",
	}

	if err := client.Get(ctx, configMapName, &configMap); err != nil {
		// not found, create a new configmap
		update = false

		configMap = corev1.ConfigMap{}
		configMap.ObjectMeta.Namespace = namespace
		configMap.ObjectMeta.Name = "authservice-configmap"
	}

	jsonData := createAuthserviceConfiguration(threads, chains)
	bytes, err := json.Marshal(jsonData)
	if err != nil {
		return nil, false
	}

	configMap.Data = make(map[string]string, 1)
	configMap.Data["config.json"] = string(bytes)

	return &configMap, update
}

func restartAuthService(client client.Client, logger logr.Logger, name, namespace string) error {
	ctx := context.Background()
	// Restart AuthService deployment by adding/updating an annotation.
	deploymentName := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	logger.Info("Restarting AuthService deployment", "deployment", deploymentName)
	var deployment appsv1.Deployment
	if err := client.Get(ctx, deploymentName, &deployment); err != nil {
		logger.Error(err, "Failed to find AuthService deployment", "deployment", deploymentName)
		return err
	}

	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string, 0)
	}
	deployment.Spec.Template.Annotations["authservice-webhook/restartedAt"] = time.Now().Format(time.RFC3339)
	if err := client.Update(ctx, &deployment); err != nil {
		logger.Error(err, "Failed to update AuthService deployment", "deployment", deploymentName)
		return err
	}

	return nil
}

func getAllChains(client client.Client, logger logr.Logger, namespace string) (*authcontroller.ChainList, error) {
	ctx := context.Background()

	// Get all the chains in the namespace.
	var chains authcontroller.ChainList
	if err := client.List(ctx, &chains, ctrlclient.InNamespace(namespace)); err != nil {
		logger.Error(err, "Failed to get chain list, ignoring")
		return nil, err
	}

	if len(chains.Items) == 0 {
		return nil, fmt.Errorf("No chains found, ignoring")
	}

	return &chains, nil
}

func createRequestAuthentication(client client.Client, logger logr.Logger, chain *authcontroller.Chain) error {

	if chain.Spec.Issuer == "" || chain.Spec.JwksURI == "" {
		logger.Info("Not creating RequestAuthentication since required values are missing")
		return nil
	}

	ctx := context.Background()

	requestAuth := istiosecurityv1beta1.RequestAuthentication{
		ObjectMeta: apisv1.ObjectMeta{
			Name:      chain.Name,
			Namespace: chain.Namespace,
		},
	}
	_, err := ctrl.CreateOrUpdate(ctx, client, &requestAuth, func() error {
		requestAuth.Spec = v1beta1.RequestAuthentication{
			JwtRules: []*v1beta1.JWTRule{
				{
					Issuer:  chain.Spec.Issuer,
					JwksUri: chain.Spec.JwksURI,
				},
			},
		}
		return nil
	})

	if err != nil {
		logger.Error(err, "Failed to create RequestAuthentication")
		return err
	}

	return nil
}
