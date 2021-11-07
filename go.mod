module github.com/DevineLiu/redis-operator

go 1.16

require (
	github.com/go-logr/logr v0.4.0
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
	sigs.k8s.io/controller-runtime v0.9.2
)
