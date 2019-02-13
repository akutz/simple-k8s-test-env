package app // import "vmw.io/sk8/app"

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"strconv"

	"golang.org/x/crypto/ssh"
)

const (
	defaultSSHUsername = "sk8"
)

// SSHConfig is SSH configuration data.
type SSHConfig struct {
	// DefaultUsername is the name of the default user that is granted access.
	// The default value is "sk8".
	DefaultUsername string `json:"default-username,omitempty"`

	// RSABits is used when generating a new SSH private key. This occurs
	// when no users are configured and one is configured automatically.
	RSABits uint16 `json:"rsa-bits,omitempty"`

	// JumpHost may be used to configure an SSH jump host for access to the
	// deployed nodes.
	JumpHost JumpHostConfig `json:"jump-host,omitempty"`

	// Users controls who has access to the deployed nodes.
	Users []UserConfig `json:"users,omitempty"`
}

// JumpHostConfig is the data used to configure an SSH jump host.
type JumpHostConfig struct {
	FQDN      string `json:"fqdn,omitempty"`
	Port      uint16 `json:"port,omitempty"`
	Username  string `json:"username,omitempty"`
	IdentFile string `json:"ident-file,omitempty"`
}

// UserConfig is the name of a user that is granted access to the
// deployed nodes. The provided SSH key may be used to access the
// nodes remotely.
type UserConfig struct {
	Name          string `json:"name,omitempty"`
	SSHPublicKey  string `json:"ssh-pub-key,omitempty"`
	SSHPrivateKey string `json:"ssh-priv-key,omitempty"`
}

func (c *SSHConfig) readEnv(ctx context.Context) error {
	if c.DefaultUsername == "" {
		c.DefaultUsername = os.Getenv("SK8_SSH_DEFAULT_USERNAME")
	}
	if c.RSABits == 0 {
		i, _ := strconv.ParseUint(os.Getenv("SK8_SSH_RSA_BITS"), 10, 16)
		c.RSABits = uint16(i)
	}
	if c.JumpHost.FQDN == "" {
		c.JumpHost.FQDN = os.Getenv("SK8_SSH_JUMP_HOST_FQDN")
	}
	if c.JumpHost.Username == "" {
		c.JumpHost.Username = os.Getenv("SK8_SSH_JUMP_HOST_USERNAME")
	}
	if c.JumpHost.IdentFile == "" {
		c.JumpHost.IdentFile = os.Getenv("SK8_SSH_JUMP_HOST_IDENT_FILE")
	}
	return nil
}

func (c *SSHConfig) validate(ctx context.Context) error {
	return nil
}

func (c *SSHConfig) setDefaults(ctx context.Context, cfg Config) error {
	if c.DefaultUsername == "" {
		c.DefaultUsername = defaultSSHUsername
	}
	if c.RSABits == 0 {
		c.RSABits = 2048
	}
	if len(c.Users) == 0 {
		if err := c.initUsers(); err != nil {
			return err
		}
	}
	return nil
}

func (c *SSHConfig) setEnv(ctx context.Context, env map[string]string) error {
	return nil
}

func (c *SSHConfig) initUsers() error {
	// Create a new SSH user.
	c.Users = []UserConfig{UserConfig{Name: c.DefaultUsername}}

	// Generate the SSH private key.
	privKey, err := rsa.GenerateKey(rand.Reader, int(c.RSABits))
	if err != nil {
		return fmt.Errorf("rsa.GenerateKey failed: %v", err)
	}
	if err := privKey.Validate(); err != nil {
		return fmt.Errorf("privKey.Validate failed: %v", err)
	}
	privDER := x509.MarshalPKCS1PrivateKey(privKey)
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}
	privKeyPEM := pem.EncodeToMemory(&privBlock)
	c.Users[0].SSHPrivateKey = base64.StdEncoding.EncodeToString(privKeyPEM)

	// Generate the SSH public key.
	pubKey, err := ssh.NewPublicKey(&privKey.PublicKey)
	if err != nil {
		return fmt.Errorf("NewPublicKey failed: %v", err)
	}
	c.Users[0].SSHPublicKey = string(ssh.MarshalAuthorizedKey(pubKey))

	return nil
}
