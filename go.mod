module github.com/jparklab/synology-csi

go 1.12

replace k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190418200329-18908d120c6b

replace k8s.io/apimachinery => k8s.io/apimachinery v0.17.5-beta.0

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/avast/retry-go v2.5.0+incompatible
	github.com/aws/aws-sdk-go v1.28.2 // indirect
	github.com/cilium/ebpf v0.0.0-20191025125908-95b36a581eed // indirect
	github.com/container-storage-interface/spec v1.2.0
	github.com/coredns/corefile-migration v1.0.6 // indirect
	github.com/docker/libnetwork v0.8.0-dev.2.0.20190925143933-c8a5fca4a652 // indirect
	github.com/elazarl/goproxy v0.0.0-20180725130230-947c36da3153 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/mock v1.3.1 // indirect
	github.com/google/go-querystring v1.0.0
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/googleapis/gnostic v0.1.0 // indirect
	github.com/kubernetes-csi/drivers v1.0.0
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/onsi/ginkgo v1.11.0 // indirect
	github.com/pborman/uuid v1.2.0
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/rexray/gocsi v1.2.1 // indirect
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	golang.org/x/net v0.0.0-20191004110552-13f9640d40b9
	gonum.org/v1/gonum v0.6.2 // indirect
	google.golang.org/grpc v1.26.0
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/gengo v0.0.0-20200114144118-36b2048a9120 // indirect
	k8s.io/klog v1.0.0
	k8s.io/kube-openapi v0.0.0-20200121204235-bf4fb3bd569c // indirect
	k8s.io/kubernetes v1.17.4 // indirect
	k8s.io/utils v0.0.0-20200324210504-a9aa75ae1b89
	sigs.k8s.io/yaml v1.2.0 // indirect
)

replace k8s.io/api => k8s.io/api v0.17.4

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.4

replace k8s.io/apiserver => k8s.io/apiserver v0.17.4

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.17.4

replace k8s.io/client-go => k8s.io/client-go v0.17.4

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.17.4

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.17.4

replace k8s.io/code-generator => k8s.io/code-generator v0.17.5-beta.0

replace k8s.io/component-base => k8s.io/component-base v0.17.4

replace k8s.io/cri-api => k8s.io/cri-api v0.17.5-beta.0

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.17.4

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.17.4

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.17.4

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.17.4

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.17.4

replace k8s.io/kubectl => k8s.io/kubectl v0.17.4

replace k8s.io/kubelet => k8s.io/kubelet v0.17.4

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.17.4

replace k8s.io/metrics => k8s.io/metrics v0.17.4

replace k8s.io/node-api => k8s.io/node-api v0.17.4

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.17.4

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.17.4

replace k8s.io/sample-controller => k8s.io/sample-controller v0.17.4
