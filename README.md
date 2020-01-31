# synology-csi ![Docker image](https://github.com/bokysan/synology-csi/workflows/Docker%20image/badge.svg)

A [Container Storage Interface](https://github.com/container-storage-interface) Driver for Synology NAS, updated to
work on amd64, armv7 and arm64.

# Platforms supported

 The driver supports linux only since it requires iscsid to be running on the host. It is currently tested with 
 Ubuntu 16.04, Ubuntu 18.04 and [Alpine](https://alpinelinux.org/).

# Install

Make sure that `iscsiadm` is installed on all the nodes where you want this attacher to run.

# Build

## Build package

    make

## Build docker image

    # e.g. docker build -t bokysan/synology-csi .
    docker build [-f Dockerfile] -t <repo>[:<tag>] .

## Build docker multiarch image

    # e.g. ./build.sh -t bokysan/synology-csi
    ./build.sh -t <repo>[:<tag>] .

# Test

  Here we use [gocsi](https://github.com/rexray/gocsi) to test the driver.
  

## Create a config file for testing

  You need to create a config file that contains information to connect to the Synology NAS API. See [Create a config file](#config) below

## Start plugin driver

    # You can specify any name for nodeid
    $ go run cmd/syno-csi-plugin/main.go \
        --nodeid CSINode \
        --endpoint tcp://127.0.0.1:10000 \
        --synology-config syno-config.yml 

## Get plugin info

    $ csc identity plugin-info -e tcp://127.0.0.1:10000

## Create a volume

    $ csc controller create-volume \
        --req-bytes 2147483648 \
        -e tcp://127.0.0.1:10000 \
        test-volume 
    "8.1" 2147483648 "iqn"="iqn.2000-01.com.synology:kube-csi-test-volume" "mappingIndex"="1" "targetID"="8"

## List volumes

    The first column in the output is the volume D

    $ csc controller list-volumes -e tcp://127.0.0.1:10000 
    "8.1" 2147483648 "iqn"="iqn.2000-01.com.synology:kube-csi-test-volume" "mappingIndex"="1" "targetID"="8"

## Delete the volume

    # e.g.
    # csc controller delete-volume  -e tcp://127.0.0.1:10000 8.1
    $ csc controller delete-volume  -e tcp://127.0.0.1:10000 <volume id>

# Deploy

## Ensure kubernetes cluster is configured for CSI drivers

  For kubernetes v1.12, and v1.13, feature gates need to be enabled to use CSI drivers.
  Follow instructions on https://kubernetes-csi.github.io/docs/csi-driver-object.html and https://kubernetes-csi.github.io/docs/csi-node-object.html
  to set up your kubernetes cluster.

## Create a config file <a name='config'></a>

    ---
    # syno-config.yml file
    host: <hostname>           # ip address or hostname of the Synology NAS
    port: 5000                 # change this if you use a port other than the default one
    sslVerify: false           # set this true to use https
    loginApiVersion: 2         # Optional. Login version. From 2 to 6. Defaults to "2".
    loginHttpMethod: <method>  # Optional. Method. "GET", "POST" or "auto" (default). "auto" uses POST on version >= 6
    username: <login>          # username
    password: <password>       # password
    sessionName: Core          # You won't need to touch this value
    enableSynoToken: no        # Optional. Set to 'true' to enable syno token. Only for versions 3 and above.
    enableDeviceToken: yes     # Optional. Set to 'true' to enable device token. Only for versions 6 and above.
    deviceId: <device-id>      # Optional. Only for versions 6 and above. If not set, DEVICE_ID environment var is read.
    deviceName: <name>         # Optional. Only for versions 6 and above.



## Create a k8s secret from the config file

    kubectl create secret generic synology-config --from-file=syno-config.yml

## (Optional) Use https with self-signed CA

  To use https with certificate that is issued by self-signed CA. CSI drivers needs to access the CA's certificate.
  You can add the certificate using configmap.

  Create a configmap with the certificate

    # e.g.
    #  kubectl create configmap synology-csi-ca-cert --from-file=self-ca.crt
    kubectl create configmap synology-csi-ca-cert --from-file=<ca file>

  Add the certificate to the deployments

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
    

## Deploy to kubernetes

    kubectl apply -f deploy/kubernetes/v1.15

    (v1.12 is also tested, v1.13 has not been tested)

    NOTE:

     synology-csi-attacher and synology-csi-provisioner need to run on the same node.
     (probably..)

### Parameters for volumes

By default, iscsi LUN will be created on Volume 1 (`/volume1`) location with thin provisioning.
You can set parameters in `storage_class.yml` to choose different locations or volume type. 

e.g.

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

NOTE: if you have already created storage class, you would need to delete the storage class and recreate it. 

# Synology configuration

As multiple logins are executed from this service at almost the same time, your Synology might block the
requests and you will see `407` errors (with version 6) or `400` errors in your log. It is advisable to
disable Auto block and IP checking if you want to get this working properly.

Make sure you do the following:
- go to Control Panel / Security / General: Enable "Enahnce browser compatibility by skipping IP checking"
- go to Control Panel / Security / Account: Disable "Auto block"
