package app // import "vmw.io/sk8/app"

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"strconv"
	"time"
)

const (
	selfSignedCACommonName = "Self-Signed CA"

	defaultTLSRSABits         = 2048
	defaultTLSCountry         = "US"
	defaultTLSStateOrProvince = "California"
	defaultTLSLocality        = "Palo Alto"
	defaultTLSOrg             = "VMware"
	defaultTLSOrgUnit         = "CNX"
)

// The default expiration for certs is 20 years.
var defaultTLSExpiresAfter time.Duration

func init() {
	now := time.Now()
	defaultTLSExpiresAfter = now.AddDate(20, 0, 0).Sub(now)
}

// TLSConfig contains the CA cert/key pair for generating certificates.
type TLSConfig struct {
	// The PEM-encoded CA certificate and key files used to generate
	// certificates by the sk8 script.
	CACrt []byte `json:"ca-crt,omitempty"`
	CAKey []byte `json:"-"`

	// The following fields are used as the defaults for generating
	// the CA and certificates .
	RSABits            uint16        `json:"rsa-bits,omitempty"`
	ExpiresAfter       time.Duration `json:"expires-after,omitempty"`
	Country            string        `json:"country,omitempty"`
	StateOrProvince    string        `json:"state-or-province,omitempty"`
	Locality           string        `json:"locality,omitempty"`
	Organization       string        `json:"org,omitempty"`
	OrganizationalUnit string        `json:"ou,omitempty"`
}

func (c *TLSConfig) readEnv(ctx context.Context) error {
	if c.RSABits == 0 {
		i, _ := strconv.ParseUint(os.Getenv("SK8_TLS_RSA_BITS"), 10, 16)
		c.RSABits = uint16(i)
	}
	if c.ExpiresAfter.Nanoseconds() == 0 {
		c.ExpiresAfter, _ = time.ParseDuration(
			os.Getenv("SK8_TLS_EXPIRES_AFTER"))
	}
	if c.Country == "" {
		c.Country = os.Getenv("SK8_TLS_COUNTRY")
	}
	if c.StateOrProvince == "" {
		c.StateOrProvince = os.Getenv("SK8_TLS_STATE_OR_PROVINCE")
	}
	if c.Locality == "" {
		c.Locality = os.Getenv("SK8_TLS_LOCALITY")
	}
	if c.Organization == "" {
		c.Organization = os.Getenv("SK8_TLS_ORG")
	}
	if c.OrganizationalUnit == "" {
		c.OrganizationalUnit = os.Getenv("SK8_TLS_OU")
	}
	return nil
}

func (c *TLSConfig) validate(ctx context.Context) error {
	return nil
}

func (c *TLSConfig) setDefaults(ctx context.Context, cfg Config) error {
	if c.RSABits == 0 {
		c.RSABits = defaultTLSRSABits
	}
	if c.ExpiresAfter.Nanoseconds() == 0 {
		c.ExpiresAfter = defaultTLSExpiresAfter
	}
	if c.Country == "" {
		c.Country = defaultTLSCountry
	}
	if c.StateOrProvince == "" {
		c.StateOrProvince = defaultTLSStateOrProvince
	}
	if c.Locality == "" {
		c.Locality = defaultTLSLocality
	}
	if c.Organization == "" {
		c.Organization = defaultTLSOrg
	}
	if c.OrganizationalUnit == "" {
		c.OrganizationalUnit = defaultTLSOrgUnit
	}

	if len(c.CACrt) > 0 || len(c.CAKey) == 0 {
		c.CACrt = nil
		c.CAKey = nil

		priv, err := rsa.GenerateKey(crand.Reader, int(c.RSABits))
		if err != nil {
			return err
		}

		now := time.Now()

		crtTpl := x509.Certificate{
			SerialNumber: big.NewInt(1),
			Subject: pkix.Name{
				CommonName:         selfSignedCACommonName,
				Country:            []string{c.Country},
				Province:           []string{c.StateOrProvince},
				Locality:           []string{c.Locality},
				Organization:       []string{c.Organization},
				OrganizationalUnit: []string{c.OrganizationalUnit},
			},
			IsCA:      true,
			NotBefore: now,
			NotAfter:  now.Add(c.ExpiresAfter),
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

		c.CACrt = caCrtPem.Bytes()
		c.CAKey = caKeyPem.Bytes()
	}
	return nil
}

func (c *TLSConfig) setEnv(
	ctx context.Context, env map[string]string) error {

	env["TLS_DEFAULT_BITS"] = strconv.Itoa(int(c.RSABits))
	env["TLS_DEFAULT_DAYS"] = strconv.Itoa(int(c.ExpiresAfter.Hours() / 24))
	env["TLS_COUNTRY_NAME"] = c.Country
	env["TLS_STATE_OR_PROVINCE_NAME"] = c.StateOrProvince
	env["TLS_LOCALITY_NAME"] = c.Locality
	env["TLS_ORG_NAME"] = c.Organization
	env["TLS_OU_NAME"] = c.OrganizationalUnit
	return nil
}
