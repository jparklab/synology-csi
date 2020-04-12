module github.com/jparklab/synology-csi

go 1.12

replace k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190418200329-18908d120c6b

replace k8s.io/apimachinery => k8s.io/apimachinery v0.18.2-beta.0

require (
	github.com/avast/retry-go v2.5.0+incompatible
	github.com/container-storage-interface/spec v1.2.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/google/go-querystring v1.0.0
	github.com/google/uuid v1.1.1 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/kubernetes-csi/drivers v1.0.0
	github.com/pborman/uuid v1.2.0
	github.com/rexray/gocsi v1.2.1 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.4.0
	golang.org/x/net v0.0.0-20191004110552-13f9640d40b9
	golang.org/x/text v0.3.2 // indirect
	google.golang.org/grpc v1.26.0
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/klog v1.0.0
	k8s.io/kubernetes v1.18.1 // indirect
	k8s.io/utils v0.0.0-20200324210504-a9aa75ae1b89
)

replace k8s.io/api => k8s.io/api v0.18.1

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.1

replace k8s.io/apiserver => k8s.io/apiserver v0.18.1

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.1

replace k8s.io/client-go => k8s.io/client-go v0.18.1

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.18.1

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.18.1

replace k8s.io/code-generator => k8s.io/code-generator v0.18.2-beta.0

replace k8s.io/component-base => k8s.io/component-base v0.18.1

replace k8s.io/cri-api => k8s.io/cri-api v0.18.2-beta.0

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.18.1

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.18.1

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.18.1

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.18.1

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.18.1

replace k8s.io/kubectl => k8s.io/kubectl v0.18.1

replace k8s.io/kubelet => k8s.io/kubelet v0.18.1

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.18.1

replace k8s.io/metrics => k8s.io/metrics v0.18.1

replace k8s.io/node-api => k8s.io/node-api v0.17.4

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.18.1

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.18.1

replace k8s.io/sample-controller => k8s.io/sample-controller v0.18.1
