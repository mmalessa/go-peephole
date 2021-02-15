module github.com/mmalessa/go-kube-test

go 1.13

require (
	github.com/spf13/viper v1.7.1
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20210206010904-48bd8381a38a
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20210206010734-c93b0f84892e
)
