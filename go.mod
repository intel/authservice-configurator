module github.com/intel/authservice-configurator

go 1.13

require (
	github.com/go-logr/logr v0.1.0
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	istio.io/api v0.0.0-20200917160826-17ee85a2cc47
	istio.io/client-go v0.0.0-20200916161914-94f0e83444ca
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	sigs.k8s.io/controller-runtime v0.6.3
)

replace gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.4.0
