# Simple Kubernetes Test Environment
The Simple Kubernetes Test Enviornment (sk8) project is:

  * _For developers building and testing Kubernetes and core Kubernetes components_
  * Capable of deploying Kubernetes 1.12+
  * Tailored for VMware Cloud (VMC) on AWS
  * Built on the Kubernetes [Cluster API (CAPI)](https://github.com/kubernetes-sigs/cluster-api)
  * Built to leverage existing tools such as `kubeadm`

## Quick Start
The first step when using sk8 is to create a default configuration file at either `${HOME}/.sk8/sk8.conf` or `/etc/sk8/sk8.conf`. This file is a multi-doc YAML manifest of one or more CAPI Cluster or Machine provider configuration objects:

```yaml
apiVersion: vsphere.sk8.vmware.io/v1alpha0
kind: ClusterProviderConfig
server:       vcenter.com
username:     myuser
password:     mypass
nat:
  apiVersion: sk8.vmware.io/v1alpha0
  kind: LinuxVirtualSwitchConfig
  publicIPAddr:   192.168.2.20
  privateIPAddr:  192.168.20.254
  ssh:
    addr:           1.2.3.4
    privateKeyPath: /Users/akutz/.ssh/id_rsa
    publicKeyPath:  /Users/akutz/.ssh/id_rsa.pub
    username: akutz
ssh:
  privateKeyPath: /Users/akutz/.ssh/id_rsa
  publicKeyPath:  /Users/akutz/.ssh/id_rsa.pub

---

apiVersion: vsphere.sk8.vmware.io/v1alpha0
kind: MachineProviderConfig
datacenter:   /dc-1
datastore:    /dc-1/datastore/WorkloadDatastore
folder:       /dc-1/vm/Workloads/sk8
resourcePool: /dc-1/host/Cluster-1/Resources/Compute-ResourcePool/sk8
network:
  interfaces:
  - name: eth0
    network: sddc-cgw-network-lvs-1
```

This branch is not configured for automated builds, so to use `sk8` a binary must be [built from source](#build-from-source) or downloaded for a pre-built OS and architecture:

| Binary                                                                             | MD5                                |
| ---------------------------------------------------------------------------------- | ---------------------------------- |
| [sk8.darwin_amd64](https://s3-us-west-2.amazonaws.com/cnx.vmware/sk8.darwin_amd64) | `3171d1fb57383bffbd5f80bc7e34918e` |
| [sk8.linux_amd64](https://s3-us-west-2.amazonaws.com/cnx.vmware/sk8.linux_amd64)   | `8196213c3792009e9b63f4a329ac7aa9` |

The next step is to...run sk8! 

```shell
$ sk8 cluster up
Creating cluster "sk8-f666fe3" ...
 ✓ Verifying prerequisites 🎈 
 ✓ Creating 1 machines(s) 🖥 

Access Kubernetes with the following command:
  kubectl --kubeconfig $(sk8 config kube sk8-f666fe3)

The nodes may also be accessed with SSH:
  ssh -F $(sk8 config ssh sk8-f666fe3) HOST

Print the available ssh HOST values using:
  sk8 config ssh sk8-f666fe3 --hosts

Finally, the cluster may be deleted with:
  sk8 cluster down sk8-f666fe3
```

The above command deploys the most recent GA build of Kubernetes:

```shell
$ kubectl --kubeconfig $(sk8 config kube sk8-f666fe3) version
Client Version: version.Info{Major:"1", Minor:"11", GitVersion:"v1.11.3", GitCommit:"a4529464e4629c21224b3d52edfe0ea91b072862", GitTreeState:"clean", BuildDate:"2018-09-09T18:02:47Z", GoVersion:"go1.10.3", Compiler:"gc", Platform:"darwin/amd64"}
Server Version: version.Info{Major:"1", Minor:"14", GitVersion:"v1.14.1", GitCommit:"b7394102d6ef778017f2ca4046abbaa23b88c290", GitTreeState:"clean", BuildDate:"2019-04-08T17:02:58Z", GoVersion:"go1.12.1", Compiler:"gc", Platform:"linux/amd64"}
```

## Configuration
sk8 is configured using CAPI configuration objects. Here's the `sk8.conf` data produced by the above `sk8 cluster up` operation:

```yaml
apiVersion: cluster.k8s.io/v1alpha1
kind: Cluster
metadata:
  creationTimestamp: null
  labels:
    sk8.vmware.io/config-dir: /Users/akutz/.sk8/sk8-f666fe3
    sk8.vmware.io/kubernetes-build-id: release/stable
    sk8.vmware.io/kubernetes-build-url: https://storage.googleapis.com/kubernetes-release/release/v1.14.1
  name: sk8-f666fe3
spec:
  clusterNetwork:
    pods:
      cidrBlocks: null
    serviceDomain: ""
    services:
      cidrBlocks: null
  providerSpec:
    value:
      apiVersion: vsphere.sk8.vmware.io/v1alpha0
      kind: ClusterProviderConfig
      nat:
        apiVersion: sk8.vmware.io/v1alpha0
        kind: LinuxVirtualSwitchConfig
        privateIPAddr: 192.168.20.254
        publicIPAddr: 192.168.2.20
        publicNIC: eth0
        ssh:
          addr: 1.2.3.4
          port: 22
          privateKey: ***
          privateKeyPath: /Users/akutz/.ssh/id_rsa
          publicKey: ***
          publicKeyPath: /Users/akutz/.ssh/id_rsa.pub
          username: akutz
      ova:
        method: content-library
        source: https://s3-us-west-2.amazonaws.com/cnx.vmware/photon3-cloud-init.ova
        target: /sk8/photon3-cloud-init
      password: 
      server: vcenter.com
      ssh:
        privateKey: ***
        privateKeyPath: /Users/akutz/.ssh/id_rsa
        publicKey: ***
        publicKeyPath: /Users/akutz/.ssh/id_rsa.pub
        username: sk8
      username: myuser
status: {}


---

apiVersion: cluster.k8s.io/v1alpha1
items:
- apiVersion: cluster.k8s.io/v1alpha1
  kind: Machine
  metadata:
    creationTimestamp: null
    labels:
      cluster.k8s.io/cluster-name: sk8-f666fe3
      sk8.vmware.io/cluster-role: control-plane,worker
      sk8.vmware.io/kubernetes-build-id: release/stable
      sk8.vmware.io/kubernetes-build-url: https://storage.googleapis.com/kubernetes-release/release/v1.14.1
    name: c01.f666fe3.sk8
  spec:
    metadata:
      creationTimestamp: null
    providerSpec:
      value:
        apiVersion: vsphere.sk8.vmware.io/v1alpha0
        datacenter: /dc-1
        datastore: /dc-1/datastore/WorkloadDatastore
        folder: /dc-1/vm/Workloads/sk8
        kind: MachineProviderConfig
        network:
          interfaces:
          - name: eth0
            network: sddc-cgw-network-lvs-1
        ova:
          method: content-library
          source: /sk8/photon3-cloud-init
        resourcePool: /dc-1/host/Cluster-1/Resources/Compute-ResourcePool/sk8
    versions:
      controlPlane: v1.14.1
      kubelet: v1.14.1
  status: {}
kind: MachineList
metadata: {}

```

As illustrated above, a sk8 cluster configuration file is a multi-doc YAML file that contains a CAPI Cluster object and a CAPI MachineList object.

## How does sk8 work?
sk8 leverages the CAPI object model and ships with support for deploying clusters to VMware Cloud (VMC) on AWS with external access provided by a Linux Virtual Switch (LVS) host or an AWS load balancer.

## What does sk8 install?
The same components installed by `kubeadm`.

## How to provision Kubernetes with sk8
Type `sk8 cluster up --help` for help.

## Build from source
The recommended way to build sk8 from source is to use Docker:

```shell
$ docker run -it --rm \
  -v $(pwd):/out \
  golang:1.12 \
  sh -c 'git clone https://github.com/akutz/simple-k8s-test-env /sk8 && \
  git -C /sk8 checkout feature/sk8-app && \
  PROGRAM=/out/sk8 make -C /sk8'
```

There should now be a new `sk8` binary in the working directory! However, the binary was built as `linux_amd64`. To target a specific OS or platform the `GOOS` and `GOARCH` environment variables may be set with Docker's `-e` flag. The following example builds `sk8` for macOS:

```shell
$ docker run -it --rm \
  -v $(pwd):/out \
  -e GOOS=darwin \
  -e GOARCH=amd64 \
  golang:1.12 \
  sh -c 'git clone https://github.com/akutz/simple-k8s-test-env /sk8 && \
  git -C /sk8 checkout feature/sk8-app && \
  PROGRAM=/out/sk8 make -C /sk8'
```

## Todo
* Better testing
* Better documentaton
* Support additional providers via gRPC endpoints (in-mem, external executable)

## License
Please the [LICENSE](LICENSE) file for information about this project's license.
