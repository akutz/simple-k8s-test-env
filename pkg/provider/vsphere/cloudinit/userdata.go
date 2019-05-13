/*
simple-kubernetes-test-environment

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloudinit

import (
	"bytes"
	"context"
	"strings"
	"text/template"

	capi "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"vmware.io/sk8/pkg/config"
	"vmware.io/sk8/pkg/util"
)

// GetUserData returns the cloud-init userdata for a machine.
func GetUserData(
	ctx context.Context,
	cluster *capi.Cluster,
	machine *capi.Machine,
	sshConfig config.SSHCredential) ([]byte, error) {

	var images []string
	role := util.GetMachineRole(machine)
	if role.Has(config.MachineRoleControlPlane) {
		images = append(images,
			"kube-apiserver.tar",
			"kube-controller-manager.tar",
			"kube-scheduler.tar",
			"kube-proxy.tar")
	}
	if role.Has(config.MachineRoleWorker) {
		images = append(images,
			"kube-proxy.tar")
	}

	tpl := template.Must(template.New("t").Parse(userdataPatt))
	buf := &bytes.Buffer{}
	if err := tpl.Execute(buf, struct {
		CloudProvider   string
		Username        string
		HostFQDN        string
		HostName        string
		DomainFQDN      string
		SSHPublicKey    string
		HealthCheckPort uint16
		KubeBuildID     string
		KubeBuildURL    string
		KubeVersion     string
		CurlFlags       string
		Packages        []string
		KubeBins        []string
		Images          []string
	}{
		CloudProvider:   cluster.Labels[config.CloudProviderLabelName],
		Username:        sshConfig.Username,
		HostFQDN:        machine.Name,
		HostName:        strings.Split(machine.Name, ".")[0],
		DomainFQDN:      strings.SplitN(machine.Name, ".", 2)[1],
		SSHPublicKey:    string(sshConfig.PublicKey),
		HealthCheckPort: 8888,
		KubeBuildID:     machine.Labels[config.KubernetesBuildIDLabelName],
		KubeBuildURL:    machine.Labels[config.KubernetesBuildURLLabelName],
		KubeVersion:     machine.Spec.Versions.Kubelet,
		CurlFlags:       "--retry 5 --retry-delay 1 --retry-max-time 120 -sSL",
		Packages: []string{
			"bindutils",
			"cni",
			"ebtables",
			"ethtool",
			"gawk",
			"inotify-tools",
			"ipset",
			"iputils",
			"ipvsadm",
			"libnetfilter_conntrack",
			"libnetfilter_cthelper",
			"libnetfilter_cttimeout",
			"libnetfilter_queue",
			"lsof",
			"socat",
			"sudo",
			"tar",
			"unzip",
		},
		KubeBins: []string{
			"kubeadm",
			"kubectl",
			"kubelet",
		},
		Images: images,
	}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

const userdataPatt = `{{$dot := .}}#cloud-config

users:
- name: {{.Username}}
  primary_group: {{.Username}}
  sudo: ALL=(ALL) NOPASSWD:ALL
  groups: sudo, wheel, docker
  ssh_import_id: None
  lock_passwd: true
  ssh_authorized_keys:
  - {{.SSHPublicKey}}

write_files:
- path: /etc/hostname
  owner: root:root
  permissions: 0644
  content: |
    {{.HostFQDN}}
- path: /etc/hosts
  owner: root:root
  permissions: 0644
  content: |
    ::1         ipv6-localhost ipv6-loopback
    127.0.0.1   localhost
    127.0.0.1   localhost.{{.DomainFQDN}}
    127.0.0.1   {{.HostName}}
    127.0.0.1   {{.HostFQDN}}
- path: /etc/systemd/system/health-check.service
  owner: root:root
  permissions: 0644
  content: |
    [Unit]
    Description=health-check.service
    After=network.target network-online.target syslog.target rc-local.service cloud-final.service

    [Install]
    WantedBy=multi-user.target

    [Service]
    WorkingDirectory=/health-check
    ExecStart=/usr/bin/python3 -m http.server {{.HealthCheckPort}}
- path: /etc/systemd/system/kubelet.service
  owner: root:root
  permissions: 0644
  content: |
    [Unit]
    Description=kubelet: The Kubernetes Node Agent
    Documentation=http://kubernetes.io/docs/

    [Service]
    ExecStart=/usr/bin/kubelet
    Restart=always
    StartLimitInterval=0
    RestartSec=10

    [Install]
    WantedBy=multi-user.target
- path: /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
  owner: root:root
  permissions: 0644
  content: |
    # Note: This dropin only works with kubeadm and kubelet v1.11+
    [Service]
    Environment="KUBELET_KUBECONFIG_ARGS=--bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --kubeconfig=/etc/kubernetes/kubelet.conf{{if .CloudProvider}} --cloud-provider={{.CloudProvider}}{{if ne .CloudProvider "external"}} --cloud-config /etc/kubernetes/cloud.conf{{end}}{{end}}"
    Environment="KUBELET_CONFIG_ARGS=--config=/var/lib/kubelet/config.yaml"
    # This is a file that "kubeadm init" and "kubeadm join" generates at runtime, populating the KUBELET_KUBEADM_ARGS variable dynamically
    EnvironmentFile=-/var/lib/kubelet/kubeadm-flags.env
    # This is a file that the user can use for overrides of the kubelet args as a last resort. Preferably, the user should use
    # the .NodeRegistration.KubeletExtraArgs object in the configuration files instead. KUBELET_EXTRA_ARGS should be sourced from this file.
    EnvironmentFile=-/etc/default/kubelet
    ExecStart=
    ExecStart=/usr/bin/kubelet $KUBELET_KUBECONFIG_ARGS $KUBELET_CONFIG_ARGS $KUBELET_KUBEADM_ARGS $KUBELET_EXTRA_ARGS
- path: /etc/sysctl.d/50-kubeadm.conf
  owner: root:root
  permissions: 0644
  content: |
    net.ipv4.ip_forward = 1
- path: /etc/sysctl.d/50-cni.conf
  owner: root:root
  permissions: 0644
  content: |
    net.bridge.bridge-nf-call-iptables = 1
- path: /etc/modules-load.d/kubeadm.conf
  owner: root:root
  permissions: 0644
  content: |
    br_netfilter
- path: /etc/docker/daemon.json
  owner: root:root
  permissions: 0644
  content: |
    {
      "exec-opts": ["native.cgroupdriver=systemd"],
      "log-driver": "json-file",
      "log-opts": {
        "max-size": "100m"
      },
      "storage-driver": "overlay2"
    }
- path: /etc/profile.d/kube-envvars.sh
  owner: root:root
  permissions: 0644
  content: |
    #!/bin/sh
    export KUBE_BUILD_ID='{{.KubeBuildID}}'
    export KUBE_BUILD_URL='{{.KubeBuildURL}}'
    export KUBE_VERSION='{{.KubeVersion}}'
- path: /etc/profile.d/prompt.sh
  owner: root:root
  permissions: 0644
  content: |
    #!/bin/sh
    export PS1="[\$?]\[\e[32;1m\]\u\[\e[0m\]@\[\e[32;1m\]\h\[\e[0m\]:\W$ \[\e[0m\]"
- path: /etc/systemd/scripts/ip4save
  owner: root:root
  permissions: 0644
  content: |
    *filter
    :INPUT ACCEPT [0:0]
    :FORWARD ACCEPT [0:0]
    :OUTPUT ACCEPT [0:0]
    COMMIT
- path: /etc/ssh/sshd_config
  owner: root:root
  permissions: 0600
  content: |
    AllowAgentForwarding   no
    AllowTcpForwarding     yes
    AllowUsers             {{.Username}}
    AuthorizedKeysFile     .ssh/authorized_keys
    ClientAliveCountMax    2
    Compression            no
    MaxAuthTries           2
    PasswordAuthentication no
    PermitEmptyPasswords   no
    PermitRootLogin        no
    PubkeyAuthentication   yes
    TCPKeepAlive           no
    UsePAM                 yes

runcmd:
- rm -fr /root/.ssh
- printf 'changeme\nchangeme' | passwd
- rm -f /usr/local/bin/sysprep.sh
- hostname '{{.HostFQDN}}'
- systemctl -l restart iptables
- mkdir -p /health-check
- systemctl -l enable health-check.service
- systemctl -l --no-block start health-check.service
- tdnf update --assumeno
- tdnf install -y{{range .Packages}} {{.}}{{end}} || true
- sysctl --system
- modprobe br_netfilter
- mkdir -p /etc/systemd/system/docker.service.d
- systemctl -l enable docker
- systemctl -l start docker.service{{range .KubeBins}}
- curl {{$dot.CurlFlags}} {{$dot.KubeBuildURL}}/bin/linux/amd64/{{.}} -o /usr/bin/{{.}}
- chmod 0755 /usr/bin/{{.}}{{end}}{{range .Images}}
- curl {{$dot.CurlFlags}} {{$dot.KubeBuildURL}}/bin/linux/amd64/{{.}} -o /{{.}}
- docker load </{{.}} && rm -f /{{.}}{{end}}
- systemctl -l enable kubelet.service
`
