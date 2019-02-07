package config // import "github.com/vmware/sk8/pkg/config"

// Config is a struct that represents the JSON config data for a sk8 client.
type Config struct {
	// Debug instructs the sk8 script to run in debug mode. This
	// causes `set -x` and sets the log level to DEBUG.
	Debug bool `json:"debug,omitempty"`

	// Sk8ScriptPath is the path to the sk8 script. This may be a local
	// file or a URL. If no value is provided then an embedded copy of
	// the script is uploaded to the nodes via cloud-init.
	Sk8ScriptPath string `json:"sk8-script-path,omitempty"`

	// Network is the network configuration.
	Network NetworkConfig `json:"network,omitempty"`

	// Nodes specifies the number and type of nodes for the cluster.
	Nodes []NodeConfig `json:"nodes,omitempty"`

	// Users is a list of people that may access the nodes remotely via SSH.
	Users []UserConfig `json:"users,omitempty"`

	// VSphere is the information used to access the vSphere endpoint
	// on which the cluster member VMs will be created.
	VSphere VSphereConfig `json:"vsphere,omitempty"`

	// VCenterSimulatorEnabled indicates whether or not to use the
	// vCenter simulator as the endpoint for the vSphere cloud provider.
	// This flag is ignored if the cloud provider is not configured for
	// the in-tree or external vSphere simulator.
	VCenterSimulatorEnabled bool `json:"vcsim,omitempty"`

	// TLS is the TLS configuration.
	TLS TLSConfig `json:"tls,omitempty"`

	// Versions of the components used when standing up the cluster.
	K8s KubernetesConfig `json:"k8s,omitempty"`

	// Env is a map of environment variable key/value pairs that are written
	// to /etc/default/sk8 and are used to configure the sk8 script.
	Env map[string]string `json:"env,omitempty"`
}

// NetworkConfig is network configuration data.
type NetworkConfig struct {
	// DomainFQDN is the domain to which the nodes belong. This domain
	// should be unique across all vCenters visible to the configured
	// vSphere endpoint. It is recommended to not set this value and
	// instead allow a unique domain be generated while turning up the
	// cluster.
	DomainFQDN string `json:"domain-fqdn,omitempty"`

	// DNS1 and DNS2 are the primary and secondary nameservers for the
	// nodes. If omitted the Google nameservers 8.8.8.8 and 8.8.4.4 are
	// used.
	DNS1 string `json:"dns1,omitempty"`
	DNS2 string `json:"dns2,omitempty"`
}

// CloudProviderConfig is the data used to configure the vSphere cloud provider.
type CloudProviderConfig struct {
	// An in-tree or external cloud provider may be configured. The
	// CloudProviderConfig is required for an in-tree cloud provider but
	// may be omitted in place of the Manifests field when configuring
	// an external cloud provider. The log level applies only to
	// external cloud providers with valid values 1-10. The default
	// log level is 2.
	Type     CloudProviderType `json:"type,omitempty"`
	Image    string            `json:"image,omitempty"`
	Config   []byte            `json:"config,omitempty"`
	LogLevel uint8             `json:"log-level,omitempty"`
}

// KubernetesConfig is the data used to configure Kubernetes.
type KubernetesConfig struct {
	// Versions of the components used when standing up the cluster.
	Version string `json:"version,omitempty"`

	// CloudProvider is the cloud provider configuration.
	CloudProvider CloudProviderConfig `json:"cloud-provider,omitempty"`

	// KubernetesLogLevel is the default log level for all of the Kubernetes
	// components. Valid log levels are 1-10. The default log level is 2.
	LogLevel                  uint8 `json:"log-level,omitempty"`
	APIServerLogLevel         uint8 `json:"api-server-log-level,omitempty"`
	SchedulerLogLevel         uint8 `json:"scheduler-log-level,omitempty"`
	ControllerManagerLogLevel uint8 `json:"controller-manager-log-level,omitempty"`
	KubeletLogLevel           uint8 `json:"kubelet-log-level,omitempty"`
	KubeProxyLogLevel         uint8 `json:"kube-proxy-log-level,omitempty"`

	// EncryptionKey is a 32 character string used by Kubernetes
	// to encrypt data. Random data is recommended unless attempting to
	// access data encrypted previously with a known key. If the value
	// is not set then a random string will be used.
	EncryptionKey string `json:"encryption-key,omitempty"`

	// Manifests is a list of Kubernetes manifests that are applied to
	// the API server once it is online.
	Manifests [][]byte `json:"manifests,omitempty"`
}

// TLSConfig contains the CA cert/key pair for generating certificates.
type TLSConfig struct {
	// The PEM-encoded CA certificate and key files used to generate
	// certificates by the sk8 script.
	CACrt []byte `json:"ca-crt,omitempty"`
	CAKey []byte `json:"-"`
}

// VSphereConfig is the information used to access a vSphere endpoint.
type VSphereConfig struct {
	Server   string `json:"server,omitempty"`
	Port     uint32 `json:"port,omitempty"`
	Username string `json:"-"`
	Password string `json:"-"`
	Insecure bool   `json:"insecure,omitempty"`

	// Template is the path to the Photon2 template used to create new
	// VMs. If the template is not available on the vSphere endpoint then
	// an attempt is made to download the template from a Photon2 OVA and
	// upload it to the vSphere endpoint.
	Template string `json:"template,omitempty"`

	// TemplateOVA is the path to a Photon2 OVA with cloud-init and the
	// the VMware GuestInfo datasource enabled. The default value is
	// https://s3-us-west-2.amazonaws.com/cnx.vmware/photon2-cloud-init.ova.
	TemplateOVA string `json:"templateOVA,omitempty"`

	// Information about where the node VMs will be created and to what
	// network they will be attached.
	Datacenter   string `json:"datacenter,omitempty"`
	Folder       string `json:"folder,omitempty"`
	Datastore    string `json:"datastore,omitempty"`
	ResourcePool string `json:"resource-pool,omitempty"`
	Network      string `json:"network,omitempty"`
}

// CloudProviderType is the type for the enumeration that defines the
// possible choices for cloud provider.
type CloudProviderType uint8

const (
	// NoCloudProvider disables the use of a cloud provider entirely.
	NoCloudProvider CloudProviderType = iota

	// InTreeCloudProvider specifies the use of an in-tree cloud provider.
	InTreeCloudProvider

	// ExternalCloudProvider indicates the use of an external cloud provider.
	ExternalCloudProvider
)

// NodeType is the type for the enumeration that defines the types of
// nodes that may be deployed.
type NodeType uint8

const (
	// ControlPlaneWorkerNode is a member of the control plane on which
	// workloads may be scheduled.
	ControlPlaneWorkerNode NodeType = iota

	// ControlPlaneNode is a member of the control plane on which
	// workloads may not be scheduled.
	ControlPlaneNode

	// WorkerNode is a node on which workloads may be scheduled.
	WorkerNode
)

// NodeConfig is the configuration information for a cluster node.
type NodeConfig struct {
	// Type is the type of node.
	Type NodeType `json:"type"`

	// Cores is the number of CPU cores for the node.
	Cores uint8 `json:"cores"`

	// CoresPerSocket is the number of CPU cores per socket.
	CoresPerSocket uint8 `json:"cores-per-socket"`

	// MemoryMiB is the amount of RAM (MiB) for the node.
	MemoryMiB uint32 `json:"mem"`

	// DiskGiB is the amount of disk space (GiB) for the node.
	DiskGiB uint32 `json:"disk"`
}

// UserConfig is the name of a user that is granted access to the
// deployed nodes. The provided SSH key may be used to access the
// nodes remotely.
type UserConfig struct {
	Name         string `json:"name"`
	SSHPublicKey string `json:"ssh-pub-key"`
}
