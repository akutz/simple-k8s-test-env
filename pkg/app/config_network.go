package app // import "vmw.io/sk8/app"

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
)

const (
	defaultDNS1 = "8.8.8.8"
	defaultDNS2 = "8.8.4.4"
)

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

func (c *NetworkConfig) readEnv(ctx context.Context) error {
	if c.DomainFQDN == "" {
		c.DomainFQDN = os.Getenv("SK8_NETWORK_DOMAIN")
	}
	if c.DNS1 == "" {
		c.DNS1 = os.Getenv("SK8_NETWORK_DNS1")
	}
	if c.DNS2 == "" {
		c.DNS2 = os.Getenv("SK8_NETWORK_DNS2")
	}
	return nil
}

func (c *NetworkConfig) validate(ctx context.Context) error {
	return nil
}

func (c *NetworkConfig) setDefaults(ctx context.Context, cfg Config) error {
	if c.DomainFQDN == "" {
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
		c.DomainFQDN = fmt.Sprintf("%s.sk8", domainID)
	}
	if c.DNS1 == "" {
		c.DNS1 = defaultDNS1
	}
	if c.DNS2 == "" {
		c.DNS2 = defaultDNS2
	}
	return nil
}

func (c *NetworkConfig) setEnv(
	ctx context.Context, env map[string]string) error {

	env["NETWORK_DOMAIN"] = c.DomainFQDN
	env["NETWORK_DNS_1"] = c.DNS1
	env["NETWORK_DNS_2"] = c.DNS2
	return nil
}
