package main

// this file is used purely for compiling project dependencies in
// the absence of the project source. this allows us to cache those
// deps in a separate docker layer and avoids having to do a full
// recompile when building new images after every source change.

// add here any deps with large build times
import (
	_ "encoding/json"
	_ "errors"
	_ "flag"
	_ "fmt"
	_ "github.com/container-storage-interface/spec/lib/go/csi"
	_ "github.com/golang/glog"
	_ "github.com/google/go-querystring/query"
	_ "github.com/kubernetes-csi/drivers/pkg/csi-common"
	_ "github.com/pborman/uuid"
	_ "github.com/spf13/cobra"
	_ "github.com/spf13/pflag"
	_ "golang.org/x/net/context"
	_ "google.golang.org/grpc/codes"
	_ "google.golang.org/grpc/status"
	_ "gopkg.in/yaml.v2"
	_ "io/ioutil"
	_ "k8s.io/klog"
	_ "k8s.io/kubernetes/pkg/util/mount"
	_ "k8s.io/utils/exec"
	_ "k8s.io/utils/nsenter"
	_ "k8s.io/utils/path"
	_ "net/http"
	_ "net/url"
	_ "os"
	_ "strings"
	_ "time"
)

func main() {

}
