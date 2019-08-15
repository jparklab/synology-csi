# synology-csi  [![Build Status](https://dev.azure.com/jparklab/synology-csi/_apis/build/status/jparklab.synology-csi?branchName=master)](https://dev.azure.com/jparklab/synology-csi/_build/latest?definitionId=2&branchName=master) [![Go Report Card](https://goreportcard.com/badge/github.com/jparklab/synology-csi)](https://goreportcard.com/report/github.com/jparklab/synology-csi)

A [Container Storage Interface](https://github.com/container-storage-interface) Driver for Synology NAS

# Platforms supported

 The driver supports linux only since it requires iscsid to be running on the host. It is currently tested with Ubuntu 16.04 and 18.04

# Build

## Build package

    make

## Build docker image

    # e.g. docker build -t jparklab/synology-csi .
    docker build -t <repo>[:<tag>] .

  Build a docker image using ubuntu stretch as the base image.

    # e.g. docker build -f Dockerfile.ubuntu -t jparklab/synology-csi .
    e.g. docker build -f Dockerfile.ubuntu -t <repo>[:tag>] .

# Test

  Here we use [gocsi](https://github.com/rexray/gocsi) to test the driver,

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
    host: <hostname>        # ip address or hostname of the Synology NAS
    port: 5000              # change this if you use a port other than the default one
    username: <login>       # username
    password: <password>    # password
    sessionName: Core       # You won't need to touch this value
    sslVerify: false        # set this true to use https

## Create a k8s secret from the config file

    kubectl create secret generic synology-config --from-file=syno-config.yml

## Deploy to kubernetes

    kubectl apply -f deploy/kubernetes/v1.15

    (v1.12 is also tested, v1.13 has not been tested)

    NOTE:

     synology-csi-attacher and synology-csi-provisioner need to run on the same node.
     (probably..)

### Parameters for volumes

By default, iscsi LUN will be created on Volume 1(/volume1) location with thin provisioning.
You can set parameters in sotrage_class.yml to choose different locations or volume type. 

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
