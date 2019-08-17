module github.com/jparklab/synology-csi

go 1.12

replace k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190418200329-18908d120c6b

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d

require (
	github.com/container-storage-interface/spec v1.0.0
	github.com/coreos/etcd v3.3.13+incompatible // indirect
	github.com/coreos/go-systemd v0.0.0-20190620071333-e64a0ec8b42a // indirect
	github.com/emicklei/go-restful v2.9.6+incompatible // indirect
	github.com/evanphx/json-patch v4.5.0+incompatible // indirect
	github.com/go-openapi/validate v0.19.2 // indirect
	github.com/gogo/protobuf v1.2.1 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/groupcache v0.0.0-20190129154638-5b532d6fd5ef // indirect
	github.com/golang/mock v1.1.1
	github.com/google/go-querystring v1.0.0
	github.com/googleapis/gnostic v0.3.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/kubernetes-csi/drivers v1.0.0
	github.com/munnerz/goautoneg v0.0.0-20190414153302-2ae31c8b6b30 // indirect
	github.com/pborman/uuid v1.2.0
	github.com/prometheus/client_golang v1.0.0 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	github.com/stretchr/testify v1.3.0
	golang.org/x/crypto v0.0.0-20190621222207-cc06ce4a13d4 // indirect
	golang.org/x/net v0.0.0-20190628185345-da137c7871d7
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	google.golang.org/grpc v1.21.1
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0-20190606204050-af9c91bd2759 // indirect
	k8s.io/apiextensions-apiserver v0.0.0-20190606210616-f848dc7be4a4 // indirect
	k8s.io/apiserver v0.0.0-20190606205144-71ebb8303503 // indirect
	k8s.io/client-go v11.0.0+incompatible // indirect
	k8s.io/cloud-provider v0.0.0-20190606212257-347f17c60af0 // indirect
	k8s.io/component-base v0.0.0-20190627205834-327675bd8ec3 // indirect
	k8s.io/csi-api v0.0.0-20190606211019-092376fff264 // indirect
	k8s.io/klog v0.3.3 // indirect
	k8s.io/kube-openapi v0.0.0-00010101000000-000000000000 // indirect
	k8s.io/kubernetes v1.14.0
	k8s.io/utils v0.0.0-20190607212802-c55fbcfc754a
	sigs.k8s.io/structured-merge-diff v0.0.0-20190628201129-059502f64143 // indirect
	sigs.k8s.io/yaml v1.1.0 // indirect
)
