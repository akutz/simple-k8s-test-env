package app // import "vmw.io/sk8/app"

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"text/template"

	"vmw.io/sk8/config"
)

// GetCloudInitUserData returns the cloud-init user data.
func GetCloudInitUserData(
	ctx context.Context,
	cfg config.Config,
	hostFQDN, nodeType string) ([]byte, error) {

	var sk8ScriptURL string
	if cfg.Sk8ScriptPath != "" {
		if strings.HasPrefix(cfg.Sk8ScriptPath, "http") {
			sk8ScriptURL = cfg.Sk8ScriptPath
		} else {
			data, err := Base64GzipFile(cfg.Sk8ScriptPath)
			if err != nil {
				return nil, err
			}
			sk8ScriptRes = data
		}
	}

	var sk8ScriptData string
	if len(sk8ScriptURL) == 0 {
		sk8ScriptData = sk8ScriptRes
	}

	sk8DefaultText := &bytes.Buffer{}
	fmt.Fprintf(sk8DefaultText, `NODE_TYPE=%s`, nodeType)
	fmt.Fprintln(sk8DefaultText)
	for k, v := range cfg.Env {
		fmt.Fprintf(sk8DefaultText, `%s=%q`, k, v)
		fmt.Fprintln(sk8DefaultText)
	}
	sk8DefaultData, err := Base64GzipBytes(sk8DefaultText.Bytes())
	if err != nil {
		return nil, err
	}

	tpl := template.Must(template.New("t").Parse(cloudInitUserDataTplFormat))
	buf := &bytes.Buffer{}
	if err := tpl.Execute(buf, struct {
		Users          []config.UserConfig
		Sk8ScriptURL   string
		Sk8ScriptData  string
		Sk8DefaultData string
		CACrt          string
		CAKey          string
		HostFQDN       string
		HostName       string
		DomainFQDN     string
		Sk8ServiceData string
	}{
		cfg.Users,
		sk8ScriptURL,
		sk8ScriptData,
		sk8DefaultData,
		base64.StdEncoding.EncodeToString(cfg.TLS.CACrt),
		base64.StdEncoding.EncodeToString(cfg.TLS.CAKey),
		hostFQDN,
		hostFQDN[:len(hostFQDN)-(len(cfg.Network.DomainFQDN)+1)],
		cfg.Network.DomainFQDN,
		base64.StdEncoding.EncodeToString([]byte(sk8ServiceFormat)),
	}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GetCloudInitNetworkData returns the cloud-init network data.
func GetCloudInitNetworkData(
	ctx context.Context,
	cfg config.Config) ([]byte, error) {

	tpl := template.Must(template.New("t").Parse(
		cloudInitNetworkConfigTplFormat))
	buf := &bytes.Buffer{}
	if err := tpl.Execute(buf, struct {
		NetworkDevice        string
		DNS1                 string
		DNS2                 string
		NetworkSearchDomains string
	}{
		"ens192",
		cfg.Network.DNS1,
		cfg.Network.DNS2,
		cfg.Network.DomainFQDN,
	}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GetCloudInitMetaData returns the cloud-init metadata.
func GetCloudInitMetaData(
	ctx context.Context,
	cfg config.Config,
	networkData []byte,
	hostFQDN string) ([]byte, error) {

	encNetworkData, err := Base64GzipBytes(networkData)
	if err != nil {
		return nil, err
	}

	tpl := template.Must(template.New("t").Parse(
		cloudInitMetaDataTplFormat))
	buf := &bytes.Buffer{}
	if err := tpl.Execute(buf, struct {
		NetworkConfig string
		HostFQDN      string
		InstanceID    string
	}{
		encNetworkData,
		hostFQDN,
		hostFQDN,
	}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GetExtraConfig gets the extraconfig data for a VM.
func GetExtraConfig(
	ctx context.Context,
	cfg config.Config,
	hostFQDN, nodeType string) (ExtraConfig, error) {

	userData, err := GetCloudInitUserData(ctx, cfg, hostFQDN, nodeType)
	if err != nil {
		return nil, err
	}
	networkData, err := GetCloudInitNetworkData(ctx, cfg)
	if err != nil {
		return nil, err
	}
	metadata, err := GetCloudInitMetaData(ctx, cfg, networkData, hostFQDN)
	if err != nil {
		return nil, err
	}

	var extraConfig ExtraConfig
	if err := extraConfig.SetCloudInitMetadata(metadata); err != nil {
		return nil, err
	}
	if err := extraConfig.SetCloudInitUserData(userData); err != nil {
		return nil, err
	}

	return extraConfig, nil
}

const cloudInitNetworkConfigTplFormat = `version: 1
config:
  - type: physical
    name: {{.NetworkDevice}}
    subnets:
      - type: dhcp
  - type: nameserver
    address:
      - {{.DNS1}}
      - {{.DNS2}} 
    search: {{.NetworkSearchDomains}}
`

const cloudInitMetaDataTplFormat = `{
  "network": "{{.NetworkConfig}}",
  "network.encoding": "gzip+base64",
  "local-hostname": "{{.HostFQDN}}",
  "instance-id": "{{.HostFQDN}}"
}
`

const cloudInitUserDataTplFormat = `#cloud-config

groups:
  - k8s-admin

{{if .Users}}users:{{range .Users}}
  - name: {{.Name}}
    primary_group: {{.Name}}
    sudo: ALL=(ALL) NOPASSWD:ALL
    groups: sudo, wheel, k8s-admin
    ssh_import_id: None
    lock_passwd: true
    ssh_authorized_keys:
      - {{.SSHPublicKey}}{{end}}{{end}}

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
{{if .CACrt}}  - path: /etc/ssl/ca.crt
    owner: root:root
    permissions: 0644
    encoding: b64
    content: {{.CACrt}}{{end}}
{{if .CAKey}}  - path: /etc/ssl/ca.key
    owner: root:root
    permissions: 0400
    encoding: b64
    content: {{.CAKey}}{{end}}
  - path: /etc/default/sk8
    owner: root:root
    permissions: 0644
    encoding: gzip
    content: !!binary |
      {{.Sk8DefaultData}}
  - path: /var/lib/sk8/sk8.service
    owner: root:root
    permissions: 0644
    encoding: b64
    content: {{.Sk8ServiceData}}
{{if .Sk8ScriptData}}  - path: /var/lib/sk8/sk8.sh
    owner: root:root
    permissions: 0755
    encoding: gzip
    content: !!binary |
      {{.Sk8ScriptData}}{{end}}

runcmd:
  - rm -f /usr/local/bin/sysprep.sh
  - hostname '{{.HostFQDN}}'
  - tdnf upgrade -y
  - tdnf install -y gawk ipvsadm unzip lsof bindutils iputils tar inotify-tools{{if .Sk8ScriptURL}}
  - mkdir -p /var/lib/sk8
  - chmod 0755 /var/lib/sk8
  - curl -sSL -o /var/lib/sk8/sk8.sh {{.Sk8ScriptURL}}
  - chmod 0755 /var/lib/sk8/sk8.sh{{end}}
  - systemctl -l enable /var/lib/sk8/sk8.service
  - systemctl -l --no-block start sk8
`

const sk8ServiceFormat = `[Unit]
Description=sk8.service

After=network.target network-online.target \
      syslog.target rc-local.service \
      cloud-final.service

ConditionPathExists=!/var/lib/sk8/.sk8.service.done

[Install]
WantedBy=multi-user.target

[Service]
Type=oneshot
RemainAfterExit=yes
TimeoutSec=0
WorkingDirectory=/var/lib/sk8

# Create the sk8 log directory.
ExecStartPre=/bin/mkdir -p /var/log/sk8

# The sk8 script is responsible for turning up the Kubernetes cluster.
ExecStart=/bin/sh -c '/var/lib/sk8/sk8.sh 2>&1 | tee /var/log/sk8/sk8.log'

# This command ensures that this service is not run on subsequent boots.
ExecStartPost=/bin/touch /var/lib/sk8/.sk8.service.done

# Finally, this command moves the sk8 configuration file to the
# /tmp directory so the file is cleaned up automatically the next time
# the temp space is reclaimed. This ensures the configuration file is
# still available for debugging errors, but *will* get cleaned up 
# eventually.
ExecStartPost=/bin/mv -f /etc/default/sk8 /tmp/sk8.defaults
`
