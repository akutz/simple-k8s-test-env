package app // import "github.com/vmware/sk8/pkg/app"

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/vmware/sk8/pkg/config"
)

// ValidateConfig validates the provided config and updates missing
// properties with default values or environment variables where
// applicable.
func ValidateConfig(
	ctx context.Context, cfg *config.Config) error {

	if cfg.Env == nil {
		cfg.Env = map[string]string{}
	}
	readVSphereConfigEnv(ctx, &cfg.VSphere)
	if err := validateVSphereConfig(ctx, &cfg.VSphere); err != nil {
		return err
	}
	if err := setConfigDefaults(ctx, cfg); err != nil {
		return err
	}

	return nil
}

func setConfigDefaults(ctx context.Context, c *config.Config) error {
	if c.Debug {
		c.Env["DEBUG"] = "true"
	}

	if err := setConfigNodeDefaults(ctx, c); err != nil {
		return err
	}

	if err := setConfigTLSDefaults(ctx, c); err != nil {
		return err
	}

	setConfigNetworkDefaults(ctx, c)
	if err := setConfigKubernetesDefaults(ctx, c); err != nil {
		return err
	}

	return nil
}

func setConfigNodeDefaults(ctx context.Context, c *config.Config) error {
	// If the config does not define any nodes then provide a default node.
	numNodes := len(c.Nodes)
	if numNodes == 0 {
		c.Nodes = []config.NodeConfig{config.NodeConfig{}}
		numNodes = 1
	}

	// Set the defaults for the configured nodes.
	var numControllers int
	for i := range c.Nodes {
		setNodeConfigDefaults(ctx, &c.Nodes[i])
		switch c.Nodes[i].Type {
		case config.ControlPlaneNode, config.ControlPlaneWorkerNode:
			numControllers++
		}
	}

	c.Env["NUM_NODES"] = strconv.Itoa(numNodes)
	c.Env["NUM_CONTROLLERS"] = strconv.Itoa(numControllers)

	if numNodes > 1 && c.Env["ETCD_DISCOVERY"] == "" {
		discoURL := fmt.Sprintf("https://discovery.etcd.io/new?size=%d", numControllers)
		resp, err := http.Get(discoURL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		c.Env["ETCD_DISCOVERY"] = string(buf)
	}

	return nil
}

func setConfigTLSDefaults(ctx context.Context, c *config.Config) error {
	if len(c.TLS.CACrt) > 0 || len(c.TLS.CAKey) == 0 {
		c.TLS.CACrt = nil
		c.TLS.CAKey = nil

		priv, err := rsa.GenerateKey(crand.Reader, 2048)
		if err != nil {
			return err
		}

		crtTpl := x509.Certificate{
			SerialNumber: big.NewInt(1),
			Subject: pkix.Name{
				CommonName:         "Self-Signed CA",
				Country:            []string{"US"},
				Province:           []string{"California"},
				Locality:           []string{"Palo Alto"},
				Organization:       []string{"VMware"},
				OrganizationalUnit: []string{"CNX"},
			},
			IsCA:      true,
			NotBefore: time.Now(),
			NotAfter:  time.Now().Add(time.Hour * 24 * 365 * 20),
			KeyUsage: x509.KeyUsageCRLSign |
				x509.KeyUsageCertSign |
				x509.KeyUsageDigitalSignature,
			BasicConstraintsValid: true,
		}

		derBytes, err := x509.CreateCertificate(
			crand.Reader, &crtTpl, &crtTpl, &priv.PublicKey, priv)
		if err != nil {
			return err
		}

		caCrtBlock := &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: derBytes,
		}
		caKeyBlock := &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(priv),
		}

		caCrtPem := &bytes.Buffer{}
		caKeyPem := &bytes.Buffer{}

		if err := pem.Encode(caCrtPem, caCrtBlock); err != nil {
			return err
		}
		if err := pem.Encode(caKeyPem, caKeyBlock); err != nil {
			return err
		}

		c.TLS.CACrt = caCrtPem.Bytes()
		c.TLS.CAKey = caKeyPem.Bytes()
	}
	return nil
}

func setConfigKubernetesDefaults(
	ctx context.Context, c *config.Config) error {

	if len(c.K8s.EncryptionKey) > 0 {
		c.Env["ENCRYPTION_KEY"] = c.K8s.EncryptionKey
	}

	return setConfigCloudProviderDefaults(ctx, c)
}

func setConfigCloudProviderDefaults(
	ctx context.Context, c *config.Config) error {

	return initCloudProviderConfig(ctx, c)
}

func setConfigNetworkDefaults(ctx context.Context, c *config.Config) {
	if len(c.Network.DNS1) == 0 {
		c.Network.DNS1 = "8.8.8.8"
	}
	if len(c.Network.DNS2) == 0 {
		c.Network.DNS2 = "8.8.4.4"
	}

	// Generate the domain ID using random data. It's possible to make
	// the ID deterministic by setting the environment variable
	// SK8_DOMAIN_ID_RAND_SEED to a valid integer. This value is then
	// used as the seed for the random generator. For example, setting
	// SK8_DOMAIN_ID_RAND_SEED=0 causes the domain ID to be 8860b1b.
	var randSeed int64
	if i, err := strconv.ParseInt(
		os.Getenv("SK8_DOMAIN_ID_RAND_SEED"), 10, 64); err == nil {
		randSeed = i
	} else {
		randSeed = time.Now().UTC().UnixNano()
	}
	r := rand.New(rand.NewSource(randSeed))
	b := make([]byte, 32)
	r.Read(b)
	domainID := fmt.Sprintf("%x", sha256.Sum256(b))[:7]
	c.Network.DomainFQDN = fmt.Sprintf("%s.sk8", domainID)
	c.Env["NETWORK_DOMAIN"] = c.Network.DomainFQDN
}

func setNodeConfigDefaults(ctx context.Context, c *config.NodeConfig) {
	switch c.Type {
	case config.ControlPlaneNode:
		if c.Cores == 0 && c.CoresPerSocket == 0 {
			c.Cores = 8
			c.CoresPerSocket = 4
		} else if c.Cores == 0 {
			c.Cores = c.CoresPerSocket
		} else if c.CoresPerSocket == 0 {
			c.CoresPerSocket = c.Cores
		}
		if c.MemoryMiB == 0 {
			c.MemoryMiB = 32768
		}
		if c.DiskGiB == 0 {
			c.DiskGiB = 20
		}
	case config.ControlPlaneWorkerNode, config.WorkerNode:
		if c.Cores == 0 && c.CoresPerSocket == 0 {
			c.Cores = 16
			c.CoresPerSocket = 4
		} else if c.Cores == 0 {
			c.Cores = c.CoresPerSocket
		} else if c.CoresPerSocket == 0 {
			c.CoresPerSocket = c.Cores
		}
		if c.MemoryMiB == 0 {
			c.MemoryMiB = 65536
		}
		if c.DiskGiB == 0 {
			c.DiskGiB = 100
		}
	}
}
