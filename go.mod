module vmware.io/sk8

go 1.12

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190221213512-86fb29eff628

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190307161346-7621a5ebb88b

replace k8s.io/api => k8s.io/api v0.0.0-20190222213804-5cb15d344471

require (
	github.com/appscode/jsonpatch v0.0.0-20190108182946-7c0e3b262f30 // indirect
	github.com/aws/aws-sdk-go v1.19.15
	github.com/go-logr/logr v0.1.0 // indirect
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/gogo/protobuf v1.2.1 // indirect
	github.com/golang/groupcache v0.0.0-20190129154638-5b532d6fd5ef // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/google/go-cmp v0.2.0
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20190212212710-3befbb6ad0cc // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/json-iterator/go v1.1.6 // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.1
	github.com/spf13/cobra v0.0.3
	github.com/vmware/govmomi v0.20.1-0.20190329012354-d3ffbeb9b353
	github.com/vmware/vmw-guestinfo v0.0.0-20170707015358-25eff159a728
	go.uber.org/atomic v1.3.2 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.9.1 // indirect
	golang.org/x/crypto v0.0.0-20190418165655-df01cb2cc480
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a // indirect
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	k8s.io/api v0.0.0-20190419092548-c5cad27821f6
	k8s.io/apiextensions-apiserver v0.0.0-20190419213629-3ff9b39ce3da // indirect
	k8s.io/apimachinery v0.0.0-20190419212445-b874eabb9a4e
	k8s.io/client-go v11.0.0+incompatible // indirect
	k8s.io/code-generator v0.0.0-20190419212335-ff26e7842f9d
	k8s.io/kube-openapi v0.0.0-20190418160015-6b3d3b2d5666 // indirect
	k8s.io/utils v0.0.0-20190308190857-21c4ce38f2a7 // indirect
	sigs.k8s.io/cluster-api v0.0.0-20190419194154-77224b1d1add
	sigs.k8s.io/controller-runtime v0.1.10 // indirect
	sigs.k8s.io/kind v0.0.0-20190413011403-161151a26faf
	sigs.k8s.io/testing_frameworks v0.1.1 // indirect
	sigs.k8s.io/yaml v1.1.0
)
