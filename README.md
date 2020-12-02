# Authservice-webhook

Authservice-webhook manages AuthService configuration. It creates and
controls two CRDs: Configuration and Chain. The users create CRs, and the
controller uses them to build authservice configuration files and wraps them
into ConfigMaps. The main use case to enable multi-tenant configuration of
AuthService running with Istio Ingress Gateway deployment.

![Authservice-webhook diagram](doc/images/authservice-webhook.png)

# Install Authservice-webhook

[Install cert-manager to the cluster](https://cert-manager.io/docs/installation/kubernetes/),
and install [kubebuilder](https://book.kubebuilder.io/quick-start.html#installation) and
[kustomize](https://kubernetes-sigs.github.io/kustomize/installation/) locally. Then run the
following commands. Replace `<registry>` and `<tag>` with suitable values for the Docker
registry you use.

```bash
make docker-build
docker tag controller <registry>/<tag>
docker push <registry>/<tag>
kubectl create namespace authservice-webhook
make deploy IMG=<registry>/<tag>
```

# Deploy Authservice

Install AuthService Service and Deployment objects. Note that AuthService
can't start yet because the ConfigMap is missing. If you want to integrate
with Istio Ingress Gateway, you should deploy this to istio-system namespace.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: authservice
  labels:
    app: authservice
spec:
type: ClusterIP
ports:
- port: 10003
  protocol: TCP
selector:
  app: authservice
---
apiVersion: apps/v1
kind: Deployment
metadata:
name: authservice
labels:
  app: authservice
spec:
replicas: 1
selector:
  matchLabels:
  app: authservice
template:
  metadata:
    labels:
      app: authservice
  spec:
    containers:
    - name: authservice
      image: adrianlzt/authservice:0.3.1-d3cd2d498169
      imagePullPolicy: Always
      ports:
      - containerPort: 10003
      volumeMounts:
      - name: authservice-configmap-volume
        mountPath: /etc/authservice
    volumes:
    - name: authservice-configmap-volume
      configMap:
        name: authservice-configmap
```

Install a Configuration object and at least one Chain. Make sure to change
the Chain values to correspond to your own OIDC installation. Install the CRs
to the namespace where you have your AuthService instance running. After this
the ConfigMap which the AuthService needs is dynamically created and AuthService deployment (whose name is defined with `authService` key in the
Configuration resource) in the same namespace is restarted.

```yaml
apiVersion: authcontroller.intel.com/v1
kind: Configuration
metadata:
  name: configuration-sample
spec:
  authService: "authservice"
  threads: 8
---
apiVersion: authcontroller.intel.com/v1
kind: Chain
metadata:
  name: chain-sample-1
spec:
  configuration: "configuration-sample"
  authorizationUri: "https://example.com/auth/realms/service-name/protocol/openid-connect/auth"
  tokenUri: "https://example.com/auth/realms/service-name/protocol/openid-connect/token"
  callbackUri: "https://example.com/service/oauth/callback"
  clientId: "service-name-client"
  clientSecret: "secret"
  trustedCertificateAuthority: "-----BEGIN CERTIFICATE-----\nMIIDMDCCAhigAwIBAgIJANeAVS2STWGLMA0GCSqGSIb3DQEBCwUAMC0xFTATBgNV\nBAoMDGV4YW1wbGUgSW5jLjEUMBIGA1UEAwwLZXhhbXBsZS5jb20wHhcNMjAwODIw\nMDcyNTM5WhcNMjEwODIwMDcyNTM5WjAtMRUwEwYDVQQKDAxleGFtcGxlIEluYy4x\nFDASBgNVBAMMC2V4YW1wbGUuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB\nCgKCAQEAs895MMU+yT7rivsjwJlWVmgzKOvK/9TW1esJCvkxsKpu/FnDmUcEJs9M\neUU8ahYgMWQFPNpYv/p2G8YqeIkXNyRtiiI0k9SG7KhkIpt1ltKjFJFBRW1hclln\nGDaDKHNraf84YK2Un/usJYW4/cOuySW41Bo5YSAqX0hrU/Cqeg2SCdZxit6kkYhg\nExK5mei1jNGJF8ILCuQlULQJjSb/b1pgyATDGu/hok2Bm6LXJMbF6B/Ti44VghNz\nLXscyQwjABmE230Tzm1g3wMJgCbjlR0prhWeYahP2mBJluG8cGZQ1KXMRmA7JA0i\ndCitaqxpattG2EtZX//32YlFgxVgCQIDAQABo1MwUTAdBgNVHQ4EFgQUcM9zQaUh\nEi07KEULbAxO/JnAiIkwHwYDVR0jBBgwFoAUcM9zQaUhEi07KEULbAxO/JnAiIkw\nDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAQyIrPxlzkVU9dPft\nKsJvh4sVyeAeT2apGkWangfG6Xf328Oh04snZtLo2ltKI5OHQD5y5EKNItOkGBCb\nh24tF3sk9PYQCDbl8xE7S6OWFHvxiKjB6m6QjxwcUPROEQHFntsGIcyj9sebmKg/\nIpoq6DGt5HfMVJLYQTOadsTF07sjWe6nIML7l3SC1l8y0UsUd8wWf2sdE6dznfuw\nKfGvKiB50yTSPFhVTQIJLainaLWPlQxKNdN8WMaMuz0NyZOTHjHAvYbP7wFmaCov\nO4RDbtyWeDqgnNiL9xv7E+iMIsCV1jpv2TnCa+U0s8DleFttzBks75ciXqECMKSE\nXuw4PQ==\n-----END CERTIFICATE-----"
  cookieNamePrefix: "service-name"
  jwks: "{\"keys\":[{\"kid\":\"Q-t9YDpVT4RWiYLuAuM88299TVnVz7sgILL6t6GcEJo\",\"kty\":\"RSA\",\"alg\":\"RS256\",\"use\":\"sig\",\"n\":\"08GTmhM2VABSHA_uEcu9xEEwQt3-BgAng8ejZzPtk_G2iuo2VhPjjqeNnEFoQRHsbXQOvqOBqMt5HCjey061XdqEieu-0ImG612au-zgG1KUyM8jd6u1LGHkcLR2yH4r4aVEJtuBy2QAhzokFvT8arje0NG8pJSrrf2VZiTK7ggZyKE8cK6zgcoMIc4PXZM1ya_ONkm9-KM4ApRh1lScfSMG8xubhJP-qWK136cN3kmDtsy1m2EOybOO_3P3RQHxCor4IUu253TWxmirOJrys5b-1BppFCZrYukrFAzRTrQ1Lkpx1-Vupb7mt3b1QpnX2RnRpWaba6XM-Su6zd2Imw\",\"e\":\"AQAB\",\"x5c\":[\"MIICnzCCAYcCBgF0NqDL6DANBgkqhkiG9w0BAQsFADATMREwDwYDVQQDDAhib29raW5mbzAeFw0yMDA4MjgxOTUwNDFaFw0zMDA4MjgxOTUyMjFaMBMxETAPBgNVBAMMCGJvb2tpbmZvMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA08GTmhM2VABSHA/uEcu9xEEwQt3+BgAng8ejZzPtk/G2iuo2VhPjjqeNnEFoQRHsbXQOvqOBqMt5HCjey061XdqEieu+0ImG612au+zgG1KUyM8jd6u1LGHkcLR2yH4r4aVEJtuBy2QAhzokFvT8arje0NG8pJSrrf2VZiTK7ggZyKE8cK6zgcoMIc4PXZM1ya/ONkm9+KM4ApRh1lScfSMG8xubhJP+qWK136cN3kmDtsy1m2EOybOO/3P3RQHxCor4IUu253TWxmirOJrys5b+1BppFCZrYukrFAzRTrQ1Lkpx1+Vupb7mt3b1QpnX2RnRpWaba6XM+Su6zd2ImwIDAQABMA0GCSqGSIb3DQEBCwUAA4IBAQDFN18HAnw/lO3kJZIqdPHE9ay3mZlWJS2z5G6/jQqpaafPkC1AmlYp0MPoHWP/uHuZBG155X+psOYdbYoe2HwoT2m05T3XUd2Nwjum0dotHQbtEiVt2ICHpizqgklWI053f2YzUyTd1tly8Qon/HBT8UuEHVeqspWLDJDSRoQQ5tQd9ekeKy28Tdj5XnN+FTF8RN2vEgg0h9AbxbiqpnGinNyGW0jskHXq96rhHQ95ySJyGnbqWruMgPpHtLRiTK3bIXvZgQmrrJ1dFsHmJ2mRLwI54rxj/accf/piSk4a149y6W62sBL4zZwiKr51+Yabil6ZbkWg4Py3HNSsCq2Y\"],\"x5t\":\"8g0YsgKHs2RDMtPin2s-9u4vAco\",\"x5t#S256\":\"NoRGdXXwbKRt8bUPmLp5AbbGydI3F1UOsBu0cjBAkco\"}]}"
match:
  header: ":path"
  prefix: "/service"
````

If used with Ingress Gateway controller, make sure Ingress Gateway proxy is configured to use AuthService. It's important that the AuthService pod isn't part of the service mesh or otherwise Istio AuthorizationPolicy is configured to ignore it, so that the connection there works without a JWT.

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
  name: external-authz-filter-for-ingress
  namespace: istio-system
spec:
  workloadSelector:
    labels:
      istio: ingressgateway
      app: istio-ingressgateway
  configPatches:
  - applyTo: HTTP_FILTER
    match:
      context: GATEWAY
      listener:
        filterChain:
          filter:
            name: "envoy.http_connection_manager"
            subFilter:
              name: "envoy.filters.http.jwt_authn"
    patch:
      operation: INSERT_BEFORE
      value:
        name: envoy.ext_authz
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.http.ext_authz.v3.ExtAuthz
          stat_prefix: ext_authz
          grpc_service:
            envoy_grpc:
              cluster_name: ext_authz
            timeout: 10s # Timeout for the entire request (including authcode for token exchange with the IDP)
  - applyTo: CLUSTER
    match:
      context: ANY
      cluster: {} # this line is required starting in istio 1.4.0
    patch:
      operation: ADD
      value:
        name: ext_authz
        connect_timeout: 5s # This timeout controls the initial TCP handshake timeout - not the timeout for the entire request
        type: LOGICAL_DNS
        lb_policy: ROUND_ROBIN
        http2_protocol_options: {}
        load_assignment:
          cluster_name: ext_authz
          endpoints:
          - lb_endpoints:
            - endpoint:
                address:
                  socket_address:
                    address: authservice
                    port_value: 10003
```

## AuthService over TLS connection

If you want to use AuthService over a TLS connection, use this EnvoyFilter:

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
  name: external-authz-filter-for-ingress
  namespace: istio-system
spec:
  workloadSelector:
    labels:
      istio: ingressgateway
      app: istio-ingressgateway
  configPatches:
  - applyTo: HTTP_FILTER
    match:
      context: GATEWAY
      listener:
        filterChain:
          filter:
            name: "envoy.http_connection_manager"
            subFilter:
              name: "envoy.filters.http.jwt_authn"
    patch:
      operation: INSERT_BEFORE
      value:
        name: envoy.ext_authz
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.http.ext_authz.v3.ExtAuthz
          stat_prefix: ext_authz
          grpc_service:
            envoy_grpc:
              cluster_name: ext_authz
            timeout: 10s # Timeout for the entire request (including authcode for token exchange with the IDP)
  - applyTo: CLUSTER
    match:
      context: ANY
      cluster: {} # this line is required starting in istio 1.4.0
    patch:
      operation: ADD
      value:
        name: ext_authz
        connect_timeout: 5s # This timeout controls the initial TCP handshake timeout - not the timeout for the entire request
        type: LOGICAL_DNS
        lb_policy: ROUND_ROBIN
        http2_protocol_options: {}
        load_assignment:
          cluster_name: ext_authz
          endpoints:
          - lb_endpoints:
            - endpoint:
                address:
                  socket_address:
                    address: authservice.istio-system
                    port_value: 443
        transport_socket:
          name: envoy.transport_sockets.tls
          typed_config:
            "@type": type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
          sni: authservice.istio-system
          common_tls_context:
            validation_context:
              match_subject_alt_names:
              - exact: "authservice.istio-system"
              trusted_ca:
                inline_string: "-----BEGIN CERTIFICATE-----\nMIIErjCCApYCCQDeffa6ViLglzANBgkqhkiG9w0BAQsFADAZMRcwFQYDVQQDDA5j\nYTAwMC5ob21lLmxhbjAeFw0yMDExMDIxNjAxNDFaFw0yMzA4MjMxNjAxNDFaMBkx\nFzAVBgNVBAMMDmNhMDAwLmhvbWUubGFuMIICIjANBgkqhkiG9w0BAQEFAAOCAg8A\nMIICCgKCAgEA3KBZTTx0tHvJBnJCVvjvfSpl+OgWIYO80NsSq/6Vu2QUJezZwtG8\no1m+JKN4O6WPsyRtcx1aV/4pPTgKKvF07IWFUG11O5Iu88pO5xJ5ssQjJNIuHXwG\nqxbYxk0fnNUl9mbJchGYxVIH2bJEULe167wnVYsQ2W3xYxzkH0kBJ/iClqf+z0aH\nzk9l0klS+4TOm2P+a5opi/suUEYKc+oC+KoXzBBa1j9tOpfoGGFCcHk8Ho1X6YE5\n4YeaCL0EkM7wIrlnJQuepjWbR01L7dVHePTFNxlUtipUkbjeuxPBDylT5DbRrXos\nDS1EmEJtbDpDs4BE51e0Q+bKOo2a19468iCyAirrSQWbsvat2Mu1LOowqMEF1zpH\n8BEmMTT8YPbKYB8+8UlUgySQs6jH5DOYFyK4Q2A8HHWjp937LcTm7wiEqbLgSvqL\nfS1+uw0PrUHSbu1vOJcZepFFw2UmTmwuWDiPWkKSLCFgyol1eJBFSV2uzVnKtO76\nWV9LH7H9tHM3wCnLuYs0Lmz9aiDZMzwTxhDZyoJt4vLPqAaWi0y+bHNyo8bBK75m\nbGBQZ4yWGTltu0ZWUGAa9eJ+dfqu6jFpEvGIcm0rVu0Czl2NPn3r4NyRnJWbiGKm\nf2IjGUbNTW+IimvJ7PZpZShac8qlrGpOs2aiz4HLvWSDye9zIuOCC2kCAwEAATAN\nBgkqhkiG9w0BAQsFAAOCAgEAPqsor531mVs5ofQjl2PaJxxA5xYenKrENP2YZUus\nBNsnXlxlk+XaDkMDrRaUqn9C7ZpVOyoEPG7sfmEEQDb+Cmxxlpbu8WZza3FODdSW\nCDyEDAOcggLX6AmJHs4rs7ebAN8UnPJ+8ULg/FBmHLgbu+sRpGv9Ws+f+5hwPwk2\nXfgLi6w02CVrvm9zG2JYCartlv0uQmMwwZoNu8i9JFWRPS07fnbEAnOOGTLV4JUC\nhZoV/pUi3oEe488NGMIv9pVVoCns8Jy/tHYFsvR6e+IaW7GzckElKCi0MwvWCDad\n9pBexGMBCLapaWtS2lCwgNC0stTFV/fVswUHVxY3S1Q3bI8joaWn0rhdjaLli5Gg\ncH9STlpIqlIZrob2DDvxE3PMrBRQunRfMCWFsz4It5XZfYSycJgkveIOGmouG0Bh\nmwOCIfPSjMg0ZLE0MIM8zv0A86PyQbavCOuuiD8G/2fHK604W4iMDhXtED4hKe9D\npj/4akfGbxDmN/mFfG0f+0yu55YbTLmar5R1YWnjaQwo8rKDmAyrD95Q2Wt6kYec\nRZVWmdZ32ASWQF4s1ORLsKJMZkDQEauR9qsX9TODQK25MbiVi6I/tEjVD5YvwoKW\neB5SMtlsd5aM1OFB+9PmZj52RXPvhvUJKx/0OmYrkPqlbKXzMhauBYsM4A7SUO0X\np0A=\n-----END CERTIFICATE-----"
```

You'll need to deploy AuthService together with an Envoy sidecar which will handle the TLS termination.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: authservice
  namespace: istio-system
  labels:
    app: authservice-behind-envoy
spec:
  type: ClusterIP
  ports:
    - port: 443
      targetPort: 30001
      protocol: TCP
      name: https
  selector:
    app: authservice-behind-envoy
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: envoy-config
  namespace: istio-system
data:
  envoy-conf.yaml: |
    static_resources:
      listeners:
      - name: listener_0
        address:
          socket_address: { address: 0.0.0.0, port_value: 30001 }
        filter_chains:
        - filters:
          - name: envoy.filters.network.http_connection_manager
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
              access_log:
              - name: envoy.access_loggers.file
                typed_config:
                  "@type": type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
                  path: /dev/stdout
              codec_type: AUTO
              stat_prefix: ingress_http
              route_config:
                name: local_route
                virtual_hosts:
                - name: backend
                  domains: ["*"]
                  routes:
                  - match: { prefix: "/" }
                    route: { cluster: local_service }
              http_filters:
              - name: envoy.filters.http.router
          transport_socket:
            name: envoy.transport_sockets.tls
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.DownstreamTlsContext
              common_tls_context:
                tls_certificates:
                  - certificate_chain: { filename: "/etc/envoy/tls/tls.crt" }
                    private_key: { filename: "/etc/envoy/tls/tls.key" }
                validation_context:
                  trusted_ca:
                    inline_string: "-----BEGIN CERTIFICATE-----\nMIIErjCCApYCCQDeffa6ViLglzANBgkqhkiG9w0BAQsFADAZMRcwFQYDVQQDDA5j\nYTAwMC5ob21lLmxhbjAeFw0yMDExMDIxNjAxNDFaFw0yMzA4MjMxNjAxNDFaMBkx\nFzAVBgNVBAMMDmNhMDAwLmhvbWUubGFuMIICIjANBgkqhkiG9w0BAQEFAAOCAg8A\nMIICCgKCAgEA3KBZTTx0tHvJBnJCVvjvfSpl+OgWIYO80NsSq/6Vu2QUJezZwtG8\no1m+JKN4O6WPsyRtcx1aV/4pPTgKKvF07IWFUG11O5Iu88pO5xJ5ssQjJNIuHXwG\nqxbYxk0fnNUl9mbJchGYxVIH2bJEULe167wnVYsQ2W3xYxzkH0kBJ/iClqf+z0aH\nzk9l0klS+4TOm2P+a5opi/suUEYKc+oC+KoXzBBa1j9tOpfoGGFCcHk8Ho1X6YE5\n4YeaCL0EkM7wIrlnJQuepjWbR01L7dVHePTFNxlUtipUkbjeuxPBDylT5DbRrXos\nDS1EmEJtbDpDs4BE51e0Q+bKOo2a19468iCyAirrSQWbsvat2Mu1LOowqMEF1zpH\n8BEmMTT8YPbKYB8+8UlUgySQs6jH5DOYFyK4Q2A8HHWjp937LcTm7wiEqbLgSvqL\nfS1+uw0PrUHSbu1vOJcZepFFw2UmTmwuWDiPWkKSLCFgyol1eJBFSV2uzVnKtO76\nWV9LH7H9tHM3wCnLuYs0Lmz9aiDZMzwTxhDZyoJt4vLPqAaWi0y+bHNyo8bBK75m\nbGBQZ4yWGTltu0ZWUGAa9eJ+dfqu6jFpEvGIcm0rVu0Czl2NPn3r4NyRnJWbiGKm\nf2IjGUbNTW+IimvJ7PZpZShac8qlrGpOs2aiz4HLvWSDye9zIuOCC2kCAwEAATAN\nBgkqhkiG9w0BAQsFAAOCAgEAPqsor531mVs5ofQjl2PaJxxA5xYenKrENP2YZUus\nBNsnXlxlk+XaDkMDrRaUqn9C7ZpVOyoEPG7sfmEEQDb+Cmxxlpbu8WZza3FODdSW\nCDyEDAOcggLX6AmJHs4rs7ebAN8UnPJ+8ULg/FBmHLgbu+sRpGv9Ws+f+5hwPwk2\nXfgLi6w02CVrvm9zG2JYCartlv0uQmMwwZoNu8i9JFWRPS07fnbEAnOOGTLV4JUC\nhZoV/pUi3oEe488NGMIv9pVVoCns8Jy/tHYFsvR6e+IaW7GzckElKCi0MwvWCDad\n9pBexGMBCLapaWtS2lCwgNC0stTFV/fVswUHVxY3S1Q3bI8joaWn0rhdjaLli5Gg\ncH9STlpIqlIZrob2DDvxE3PMrBRQunRfMCWFsz4It5XZfYSycJgkveIOGmouG0Bh\nmwOCIfPSjMg0ZLE0MIM8zv0A86PyQbavCOuuiD8G/2fHK604W4iMDhXtED4hKe9D\npj/4akfGbxDmN/mFfG0f+0yu55YbTLmar5R1YWnjaQwo8rKDmAyrD95Q2Wt6kYec\nRZVWmdZ32ASWQF4s1ORLsKJMZkDQEauR9qsX9TODQK25MbiVi6I/tEjVD5YvwoKW\neB5SMtlsd5aM1OFB+9PmZj52RXPvhvUJKx/0OmYrkPqlbKXzMhauBYsM4A7SUO0X\np0A=\n-----END CERTIFICATE-----"
      clusters:
      - name: local_service
        connect_timeout: 3.25s
        type: LOGICAL_DNS
        lb_policy: ROUND_ROBIN
        http2_protocol_options: {}
        load_assignment:
          cluster_name: some_service
          endpoints:
          - lb_endpoints:
            - endpoint:
                address:
                  socket_address:
                    address: 127.0.0.1
                    port_value: 10003
    admin:
      access_log_path: "/dev/null"
      address:
        socket_address:
          address: 0.0.0.0
          port_value: 9001
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: authservice-behind-envoy
  namespace: istio-system
  labels:
    app: authservice-behind-envoy
spec:
  replicas: 1
  selector:
      matchLabels:
        app: authservice-behind-envoy
  template:
    metadata:
      labels:
        app: authservice-behind-envoy
    spec:
      containers:
      - name: envoy
        image: envoyproxy/envoy:v1.16-latest
        imagePullPolicy: IfNotPresent
        securityContext:
          capabilities:
            add: ["IPC_LOCK"]
        args:
          - "-c"
          - "/etc/envoy/config/envoy-conf.yaml"
          - "--cpuset-threads"
        ports:
          - containerPort: 30001
        volumeMounts:
          - name: tls
            mountPath: /etc/envoy/tls
            readOnly: true
          - name: config
            mountPath: /etc/envoy/config
            readOnly: true
          - name: resetdir
            mountPath: /etc/ssl
      - name: authservice
        image: adrianlzt/authservice:0.3.1-d3cd2d498169
        imagePullPolicy: Always
        ports:
          - containerPort: 10003
        volumeMounts:
          - name: authservice-configmap-volume
            mountPath: /etc/authservice
      volumes:
      - name: authservice-configmap-volume
        configMap:
          name: authservice-configmap
      - name: resetdir
        emptyDir: {}
      - name: tls
        secret:
          secretName: authservice-certs
      - name: config
        configMap:
          name: envoy-config
```

In addition you'll need to create secret `authservice-certs` which has
files `/etc/envoy/tls/tls.crt` and `/etc/envoy/tls/tls.key`, and the cert
needs to be signed by the CA referenced in the `trusted_ca` field above. The
SNI has to be `authservice.istio-system`.

# Known issues and missing features
  * Better defaults for RBAC for the Configuration and Chain objects
