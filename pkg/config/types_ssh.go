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

package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"path"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SSHCredentialConfig is the information used to connect to an SSH endpoint.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SSHCredentialConfig struct {
	// TypeMeta representing the type of the object and its API schema version.
	metav1.TypeMeta `json:",inline"`
	SSHCredential   `json:",inline"`
}

// SSHCredential is the information used to connect to an SSH endpoint.
type SSHCredential struct {
	// PrivateKey is the private key used to connect to the SSH server.
	// If omitted and the file PrivateKeyPath does not exist then a new
	// private and public key pair is generated.
	//
	// +optional
	PrivateKey []byte `json:"privateKey,omitempty"`

	// PrivateKeyPath is the path to the private key used to connect to the
	// SSH server. If specified then a PublicKeyPath is required as well.
	//
	// +optional
	PrivateKeyPath string `json:"privateKeyPath,omitempty"`

	// PublicKey is the public side of the keypair.
	//
	// If this value is specified but PrivateKey is not then this value
	// will be replaced because a new key pair is generatd if PrivateKey
	// is not set.
	//
	// +optional
	PublicKey []byte `json:"publicKey,omitempty"`

	// PublicKeyPath is the path to the public key used to connect to the
	// SSH server.
	//
	// If this valus is specified but PrivateKeyPath is not, then this
	// value is ignored.
	//
	// +optional
	PublicKeyPath string `json:"publicKeyPath,omitempty"`

	// Username is the username used to connect to the SSH server.
	//
	// Defaults to "sk8"
	//
	// +optional
	Username string `json:"username,omitempty"`
}

// SetDefaults_SSHCredential sets uninitialized fields to their default value.
func SetDefaults_SSHCredential(obj *SSHCredential) {
	if obj.Username == "" {
		obj.Username = "sk8"
	}
	if len(obj.PrivateKey) == 0 && len(obj.PublicKey) == 0 {
		if obj.PrivateKeyPath != "" && obj.PublicKeyPath != "" {
			if d := os.Getenv("SK8_SSH_DIR"); d != "" {
				obj.PrivateKeyPath = path.Join(d, obj.PrivateKeyPath)
				obj.PublicKeyPath = path.Join(d, obj.PublicKeyPath)
			}
			if len(obj.PrivateKey) == 0 {
				obj.PrivateKey, _ = ioutil.ReadFile(obj.PrivateKeyPath)
				obj.PublicKey, _ = ioutil.ReadFile(obj.PublicKeyPath)
			}
		} else {
			prv, pub, err := sshKeyGen()
			if err != nil {
				panic(err)
			}
			obj.PrivateKey = prv
			obj.PublicKey = pub
		}
	}
}

// SSHCredentialAndEndpoint is a composite of SSHCredential and SSHEndpoint.
type SSHCredentialAndEndpoint struct {
	SSHCredential `json:",inline"`
	SSHEndpoint   `json:",inline"`
}

// SSHEndpointConfig is information used to access an SSH server.
//
// Port defaults to 22
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SSHEndpointConfig struct {
	// TypeMeta representing the type of the object and its API schema version.
	metav1.TypeMeta `json:",inline"`
	SSHEndpoint     `json:",inline"`
}

// SSHEndpoint is information used to access an SSH server.
type SSHEndpoint struct {
	ServiceEndpoint `json:",inline"`

	// ProxyHost is an optional SSH endpoint used to connect to this
	// SSH endpoint. For example, if ProxyHost is defined then the client
	// connects first to ProxyHost and then makes a second hop to the
	// address defined in this object.
	//
	// +optional
	ProxyHost *SSHEndpoint `json:"proxy,omitempty"`
}

// SetDefaults_SSHEndpoint sets uninitialized fields to their default value.
func SetDefaults_SSHEndpoint(obj *SSHEndpoint) {
	if obj.Port == 0 {
		obj.Port = 22
	}
	if obj.ProxyHost != nil {
		SetDefaults_SSHEndpoint(obj.ProxyHost)
	}
}

// String returns the textual representation of an SSHEndpoint.
func (s SSHEndpoint) String() string {
	return s.ServiceEndpoint.String()
}

func sshKeyGen() ([]byte, []byte, error) {
	// Generate the SSH private key.
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, errors.Wrap(err, "rsa.GenerateKey failed")
	}
	if err := privKey.Validate(); err != nil {
		return nil, nil, errors.Wrap(err, "privKey.Validate failed")
	}
	privDER := x509.MarshalPKCS1PrivateKey(privKey)
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}
	privKeyBuf := pem.EncodeToMemory(&privBlock)

	// Generate the SSH public key.
	pubKey, err := ssh.NewPublicKey(&privKey.PublicKey)
	if err != nil {
		return nil, nil, errors.Wrap(err, "ssh.NewPublicKey failed")
	}
	pubKeyBuf := ssh.MarshalAuthorizedKey(pubKey)

	return privKeyBuf, pubKeyBuf, nil
}