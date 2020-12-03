# synology-csi ![Docker image](https://github.com/jparklab/synology-csi/workflows/Docker%20image/badge.svg) ![Go Report Card](https://goreportcard.com/badge/github.com/jparklab/synology-csi)

A [Container Storage Interface](https://github.com/container-storage-interface) Driver for Synology NAS.

# Platforms Supported

 The driver supports linux only since it requires `iscsiadm` to be installed on the host. It is currently tested with Ubuntu 16.04, Ubuntu 18.04, and [Alpine](https://alpinelinux.org/).

 We have pre-built Docker [images](https://hub.docker.com/r/jparklab/synology-csi) for amd64, arm64, armv7 architectures.

# Quickstart Guide

## Synology Configuration:

 1. Create a Storage Pool
 2. Create a Volume
 3. Go to Control Panel > Security > General: Enable "Enhance browser compatibility by skipping IP checking"
 4. Go to Control Panel > Security > Account: Disable "Auto block"
 5. Create a Storage User service account and add it to the "administrators" group

## Synology CSI Configuration and Setup

 1. Clone the repository
 2. Perform the following on all your Kubernetes cluster nodes (Ubuntu):
   * Install the build dependencies: `apt-get update && apt-get install -y open-iscsi make && snap install go --classic`
   * Change (`cd`) to the cloned repository
   * Run the Makefile build: `make`
   * Create the kubelet plugins directory: `mkdir -p /var/lib/kubelet/plugins/csi.synology.com/`
   * Copy the built binary from `<cloned repository path>/bin/` to the kubelet plugin directory: `cp ./bin/synology-csi-driver /var/lib/kubelet/plugins/csi.synology.com/`
 3. Edit the appropriate deployment for your Kubernetes version as needed
   * You will probably want to adjust the `storage_class.yml` based on your Synology volume, filesystem, and provisioning type (thick/thin) and additionally the name of the class as well to meet your requirements/preferences.
   * You may also have to build containers depending on your hardware architecture and adjust the container images in the deployments.  This will involve building the container images for this project and the following projects:  
       - https://github.com/kubernetes-csi/node-driver-registrar
       - https://github.com/kubernetes-csi/external-provisioner
       - https://github.com/kubernetes-csi/external-attacher
 4. Create a [syno-config.yml](syno-config.yml)
 5. Create the namespace: `kubectl create ns synology-csi`
 6. Create a secret from the customized [syno-config.yml](syno-config.yml): `kubectl create secret -n synology-csi generic synology-config --from-file=syno-config.yml`
 7. Apply the deployment: `kubectl apply -f deploy/<kubernetes version>`

 At this point you should be able to deploy persistent volume claims with the new storage class.

---

# Install Details

Make sure that `iscsiadm` is installed on all the nodes where you want this attacher to run.

# Build

## Build package

```bash
make
```

## Build Docker Image

```bash
# e.g. docker build -t jparklab/synology-csi .
docker build [-f Dockerfile] -t <repo>[:<tag>] .
```

## Build Docker Multi-architecture Images

In order to build a multiarch image, you must have Docker 19.03 or higher version that supports [buildx](https://docs.docker.com/buildx/working-with-buildx/)

```bash
# e.g. ./build.sh -t jparklab/synology-csi
./build.sh -t <repo>[:<tag>] .
```

# Test

Here we use [gocsi](https://github.com/rexray/gocsi) to test the driver.

## Create a Configuration File for Testing

You need to create a config file that contains information to connect to the Synology NAS API. See [Create a config file](#config) below

## Start Plugin Driver

```bash
# You can specify any name for nodeid
go run cmd/syno-csi-plugin/main.go \
  --nodeid CSINode \
  --endpoint tcp://127.0.0.1:10000 \
  --synology-config syno-config.yml
```

## Get plugin info

```bash
csc identity plugin-info -e tcp://127.0.0.1:10000
```

## Create a volume

```bash
csc controller create-volume \
  --req-bytes 2147483648 \
  -e tcp://127.0.0.1:10000 \
  test-volume
"8.1" 2147483648 "iqn"="iqn.2000-01.com.synology:kube-csi-test-volume" "mappingIndex"="1" "targetID"="8"
```

## List Volumes

The first column in the output is the volume D

```bash
csc controller list-volumes -e tcp://127.0.0.1:10000
"8.1" 2147483648 "iqn"="iqn.2000-01.com.synology:kube-csi-test-volume" "mappingIndex"="1" "targetID"="8"
```

## Delete the Volume

```bash
# e.g.
## csc controller delete-volume  -e tcp://127.0.0.1:10000 8.1
csc controller delete-volume  -e tcp://127.0.0.1:10000 <volume id>
```
# Deploy

## Ensure Kubernetes Cluster is Configured for CSI Drivers

  For Kubernetes v1.12, and v1.13, feature gates need to be enabled to use CSI drivers.
  Follow instructions on https://kubernetes-csi.github.io/docs/csi-driver-object.html and https://kubernetes-csi.github.io/docs/csi-node-object.html
  to set up your Kubernetes cluster.

## Create a config file <a name='config'></a>

```yaml
---
# syno-config.yml file
host: <hostname>           # ip address or hostname of the Synology NAS
port: 5000                 # change this if you use a port other than the default one
sslVerify: false           # set this true to use https
username: <login>          # username
password: <password>       # password
loginApiVersion: 2         # Optional. Login version. From 2 to 6. Defaults to "2".
loginHttpMethod: <method>  # Optional. Method. "GET", "POST" or "auto" (default). "auto" uses POST on version >= 6
sessionName: Core          # You won't need to touch this value
enableSynoToken: no        # Optional. Set to 'true' to enable syno token. Only for versions 3 and above.
enableDeviceToken: yes     # Optional. Set to 'true' to enable device token. Only for versions 6 and above.
deviceId: <device-id>      # Optional. Only for versions 6 and above. If not set, DEVICE_ID environment var is read.
deviceName: <name>         # Optional. Only for versions 6 and above.
```


## Create a Secret from the syno-config.yml file

    kubectl create secret -n synology-csi generic synology-config --from-file=syno-config.yml

### (Optional) Use https with self-signed CA

  To use https with certificate that is issued by self-signed CA. CSI drivers needs to access the CA's certificate.
  You can add the certificate using configmap.

  Create a configmap with the certificate

```bash
# e.g.
##  kubectl create configmap synology-csi-ca-cert --from-file=self-ca.crt
kubectl create configmap synology-csi-ca-cert --from-file=<ca file>
```

  Add the certificate to the deployments

```yaml
# Add to attacher.yml, node.yml, and provisioner.yml
..
spec:
...
- name: csi-plugin
...
  volumeMounts:
  ...
  - mountPath: /etc/ssl/certs/self-ca.crt
    name: cert
    subPath: self-ca.crt      # this should be the same as the file name that is used to create the configmap
...
volumes:
- configMap:
    defaultMode: 0444
    name: synology-csi-ca-cert
```

## Deploy to Kubernetes

```bash
kubectl apply -f deploy/kubernetes/v1.15
```

### Parameters for the StorageClass and Synology

By default, iscsi LUN will be created on Volume 1 (`/volume1`) location with thin provisioning.
You can set parameters in `storage_class.yml` to choose different locations or volume type.

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
name: synology-iscsi-storage
...
provisioner: csi.synology.com
parameters:
  location: '/volume2'
  type: 'FILE'          # if the location has ext4 file system, use FILE for thick provisioning, and THIN for thin provisioning.
                        # for btrfs file system, use BLUN_THICK for thick provisioning, and BLUN for thin provisioning.
reclaimPolicy: Delete
allowVolumeExpansion: true # support from Kubernetes 1.16
```

***NOTE:*** if you have already created storage class, you would need to delete the storage class and recreate it.

# Synology Configuration Details

As multiple logins are executed from this service at almost the same time, your Synology might block the
requests and you will see `407` errors (with version 6) or `400` errors in your log. It is advisable to
disable Auto block and IP checking if you want to get this working properly.

Make sure you do the following:
- go to Control Panel / Security / General: Enable "Enhance browser compatibility by skipping IP checking"
- go to Control Panel / Security / Account: Disable "Auto block"
