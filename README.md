# Simple Kubernetes Test Environment
The Simple Kubernetes Test Enviornment (sk8) project is:

  * _For developers building and testing Kubernetes and core Kubernetes components_
  * Capable of deploying Kubernetes 1.12+
  * Tailored for VMware Cloud (VMC) on AWS
  * Built on the Kubernetes [Cluster API (CAPI)](https://github.com/kubernetes-sigs/cluster-api)
  * Built to leverage existing tools such as `kubeadm`

## Quick Start
The first step when using sk8 is to create a default configuration file at either `${HOME}/.sk8/sk8.conf` or `/etc/sk8/sk8.conf`. This file is a multi-doc YAML manifest of one or more CAPI Cluster or Machine provider configuration objects. The example below leverages the AWS account linked to a VMC SDDC in order to provide external access to the deployed machines via an AWS elastic load balancer:

```yaml
apiVersion: vsphere.sk8.vmware.io/v1alpha0
kind: ClusterProviderConfig
server:   # defaults to env var VSPHERE_SERVER
username: # defaults to env var VSPHERE_USERNAME
password: # defaults to env var VSPHERE_PASSWORD
nat:
  apiVersion: vsphere.sk8.vmware.io/v1alpha0
  kind: AWSLoadBalancerConfig
  accessKeyID:     # defaults to AWS_ACCESS_KEY_ID
  secretAccessKey: # defaults to AWS_SECRET_ACCESS_KEY
  region:          # defaults to AWS_DEFAULT_REGION, AWS_REGION
  healthCheckPort: # defaults to 8888
  subnetID: subnet-fdee56b6
  vpcID: vpc-8f7048f6

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
    network: sddc-cgw-network-3
```

This branch is not configured for automated builds, so to use `sk8` a binary must be [built from source](#build-from-source) or downloaded for a pre-built OS and architecture:

| Binary | MD5 |
|--------|-----|
| [sk8.darwin_amd64](https://s3-us-west-2.amazonaws.com/cnx.vmware/sk8.darwin_amd64) | `207f6e359906ee6e5510c608967c1740` |
| [sk8.linux_amd64](https://s3-us-west-2.amazonaws.com/cnx.vmware/sk8.linux_amd64)   | `93a47fd1242fba698e4444a6632993c2` |

The next step is to...run sk8! 

```shell
$ sk8 cluster up
Creating cluster "sk8-3eabf10" ...
 ✓ Verifying prerequisites 🎈 
 ✓ Creating 1 machines(s) 🖥 
 ✓ Configuring control plane 👀 
Name:        sk8-3eabf10
Created:     2019-05-06 12:35:28.983283 -0500 CDT m=+6.580907332
Kubeconfig:  /Users/akutz/.sk8/sk8-3eabf10/kube.conf
Machines:
  Name:      c01.3eabf10.sk8
  Created:   2019-05-06 12:40:18.196913 -0500 CDT m=+295.791836818
  Roles:     control-plane,worker
  Versions:
    Control: v1.14.1
    Kubelet: v1.14.1

Print the nodes with the following command:
  kubectl --kubeconfig /Users/akutz/.sk8/sk8-3eabf10/kube.conf get nodes

Query the state of the Kubernetes system components:
  kubectl --kubeconfig /Users/akutz/.sk8/sk8-3eabf10/kube.conf -n kube-system get all

Finally, the cluster may be deleted with:
  sk8 cluster down sk8-3eabf10
```

The above command deploys the most recent GA build of Kubernetes:

```shell
$ kubectl --kubeconfig $(sk8 config k8s sk8-3eabf10) version
Client Version: version.Info{Major:"1", Minor:"11", GitVersion:"v1.11.3", GitCommit:"a4529464e4629c21224b3d52edfe0ea91b072862", GitTreeState:"clean", BuildDate:"2018-09-09T18:02:47Z", GoVersion:"go1.10.3", Compiler:"gc", Platform:"darwin/amd64"}
Server Version: version.Info{Major:"1", Minor:"14", GitVersion:"v1.14.1", GitCommit:"b7394102d6ef778017f2ca4046abbaa23b88c290", GitTreeState:"clean", BuildDate:"2019-04-08T17:02:58Z", GoVersion:"go1.12.1", Compiler:"gc", Platform:"linux/amd64"}
```

## Configuration
sk8 is configured using CAPI configuration objects. Here's the `sk8.conf` data produced by the above `sk8 cluster up` operation:

```yaml
apiVersion: cluster.k8s.io/v1alpha1
kind: Cluster
metadata:
  creationTimestamp: "2019-05-06T17:35:28Z"
  labels:
    sk8.vmware.io/config-dir: /Users/akutz/.sk8/sk8-3eabf10
    sk8.vmware.io/kubernetes-build-id: release/stable
    sk8.vmware.io/kubernetes-build-url: https://storage.googleapis.com/kubernetes-release/release/v1.14.1
  name: sk8-3eabf10
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
        accessKeyID:     # redacted
        apiVersion: vsphere.sk8.vmware.io/v1alpha0
        healthCheckPort: 8888
        kind: AWSLoadBalancerConfig
        region:          # redacted
        secretAccessKey: # redacted
        subnetID: subnet-fdee56b6
        vpcID: vpc-8f7048f6
      ova:
        method: content-library
        source: https://s3-us-west-2.amazonaws.com/cnx.vmware/photon3-cloud-init.ova
        target: /sk8/photon3-cloud-init
      password: # redacted
      server:   # redacted
      ssh:
        privateKey: # redacted
        publicKey:  # redacted
        username: sk8
      username:     # redacted
status:
  apiEndpoints:
  - host: sk8-3eabf10-65533edb50203bd1.elb.us-west-2.amazonaws.com
    port: 443
  providerStatus:
    aws:
      apiTargetGroupARN: arn:aws:elasticloadbalancing:us-west-2:571501312763:targetgroup/sk8-3eabf10-api/6d96ba18d519ebd8
      loadBalancerARN: arn:aws:elasticloadbalancing:us-west-2:571501312763:loadbalancer/net/sk8-3eabf10/65533edb50203bd1
      loadBalancerDNS: sk8-3eabf10-65533edb50203bd1.elb.us-west-2.amazonaws.com
      sshTargetGroupARN: arn:aws:elasticloadbalancing:us-west-2:571501312763:targetgroup/sk8-3eabf10-ssh/9b1cf2c82a8a28a4
    kubeJoinCmd: kubeadm join 192.168.3.218:443 --token uzq5qr.e4vll2ltsi5irksj --discovery-token-ca-cert-hash
      sha256:8d6f4985b8dcf8a00a4ef2f56972b770893b878d55b86dc63529d8c847bc46b9
    ovaID: 4a92e848-b44d-4ebb-8ba8-9b75fe0fe45a
    ssh:
      addr: sk8-3eabf10-65533edb50203bd1.elb.us-west-2.amazonaws.com
      port: 22


---

apiVersion: cluster.k8s.io/v1alpha1
items:
- apiVersion: cluster.k8s.io/v1alpha1
  kind: Machine
  metadata:
    creationTimestamp: "2019-05-06T17:40:18Z"
    labels:
      cluster.k8s.io/cluster-name: sk8-3eabf10
      sk8.vmware.io/cluster-role: control-plane,worker
      sk8.vmware.io/kubernetes-build-id: release/stable
      sk8.vmware.io/kubernetes-build-url: https://storage.googleapis.com/kubernetes-release/release/v1.14.1
    name: c01.3eabf10.sk8
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
            network: sddc-cgw-network-3
        ova:
          method: content-library
          source: /sk8/photon3-cloud-init
        resourcePool: /dc-1/host/Cluster-1/Resources/Compute-ResourcePool/sk8
    versions:
      controlPlane: v1.14.1
      kubelet: v1.14.1
  status:
    addresses:
    - address: 192.168.3.218
      type: InternalIP
kind: MachineList
metadata: {}
```

As illustrated above, a sk8 cluster configuration file is a multi-doc YAML file that contains a CAPI Cluster object and a CAPI MachineList object.

## Cloud provider
Currently sk8 does not do any configuration of any external components, including the Kubernetes cloud provider. However, sk8 *can* bootstrap the cluster to expect that a cloud provider *will be* configured. Use `sk8 cluster up --cloud-provider NAME` to specify that a cloud provider will be configured for the cluster. The `NAME` placeholder can also be set to `external` to indicate an external cloud-controller manager.

## External access
Machines provisioned on VMC on AWS are not, by default, accessible from the public internet. sk8 deploys clusters that are externally accessible by using one of two methods:

1. [LVS host](#linux-virtual-switch)
2. [AWS load balancer](#aws-load-balancer)

### Linux virtual switch
TODO

### AWS load balancer
TODO

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
