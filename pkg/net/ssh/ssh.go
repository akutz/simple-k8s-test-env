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

package ssh

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	"vmware.io/sk8/pkg/config"
)

// Client is a wrapper around an SSH client to ensure a possible bastion
// client connection is closed as well.
type Client struct {
	*ssh.Client
	bastion *Client
	addr    string
}

// Close the SSH client.
func (c *Client) Close() error {
	err := c.Client.Close()
	if c.bastion != nil {
		if err2 := c.bastion.Close(); err2 != nil {
			return err2
		}
	}
	return err
}

// NewClient returns a new SSH client for the provided endpoint and
// credentials.
func NewClient(
	endpoint config.SSHEndpoint,
	creds config.SSHCredential) (*Client, error) {

	sshConfig, err := NewClientConfig(creds.Username, creds.PrivateKey)
	if err != nil {
		return nil, err
	}

	// If the endpoint does not use a proxy then the connection is simple.
	if endpoint.ProxyHost == nil {
		addr := endpoint.String()
		log.WithFields(log.Fields{
			"user": creds.Username,
			"addr": addr,
		}).Debug("ssh-dial")

		client, err := ssh.Dial("tcp", addr, sshConfig)
		if err != nil {
			return nil, errors.Wrapf(err, "error dialing ssh host %q", addr)
		}

		return &Client{Client: client, addr: addr}, nil
	}

	// Okay, an SSH proxy is being used. Not horrible, but slightly more
	// interesting. First things first, dial the proxy first.
	bastionAddr := endpoint.ProxyHost.String()
	log.WithFields(log.Fields{
		"user":        creds.Username,
		"bastionAddr": bastionAddr,
	}).Debug("ssh-dial-bastion")

	bastionClient, err := ssh.Dial("tcp", bastionAddr, sshConfig)
	if err != nil {
		return nil, errors.Wrapf(
			err, "error dialing ssh bastion host %q", bastionAddr)
	}

	log.WithFields(log.Fields{
		"user":        creds.Username,
		"targetAddr":  endpoint.String(),
		"bastionAddr": bastionAddr,
	}).Debug("ssh-dial-target")

	targetClient, err := DialByProxy(bastionClient, sshConfig, endpoint)
	if err != nil {
		return nil, errors.Wrapf(
			err, "err dialing ssh target host %q from bastion host %q",
			endpoint, bastionAddr)
	}

	return targetClient, nil
}

// NewClientConfig returns a new SSH client config.
func NewClientConfig(
	username string,
	privateKey []byte) (*ssh.ClientConfig, error) {

	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing ssh private key")
	}
	return &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 30,
	}, nil
}

// DialByProxy dials a target SSH endpoint from an existing SSH client.
func DialByProxy(
	bastionClient *ssh.Client,
	sshConfig *ssh.ClientConfig,
	targetEndpoint config.SSHEndpoint) (*Client, error) {

	addr := targetEndpoint.String()
	conn, err := bastionClient.Dial("tcp", addr)
	if err != nil {
		return nil, errors.Wrap(err, "error dialing target from bastion")
	}

	ncc, chans, reqs, err := ssh.NewClientConn(conn, addr, sshConfig)
	if err != nil {
		return nil, errors.Wrap(err, "error creating target conn from bastion")
	}

	client := ssh.NewClient(ncc, chans, reqs)
	if err != nil {
		return nil, errors.Wrap(err, "error creating target client from bastion")
	}

	return &Client{
		Client:  client,
		addr:    addr,
		bastion: &Client{Client: bastionClient},
	}, nil
}

// Run executes the given command on the remote host.
func Run(
	ctx context.Context, client *Client,
	stdin io.Reader, stdout io.Writer, stderr io.Writer,
	format string, args ...interface{}) error {

	cmd := fmt.Sprintf(format, args...)
	sess, err := client.NewSession()
	if err != nil {
		return errors.Wrapf(err, "failed to create ssh session to %q", cmd)
	}
	defer sess.Close()

	stdout2 := &bytes.Buffer{}
	stderr2 := &bytes.Buffer{}

	stdoutPipes := []io.Writer{stdout2}
	stderrPipes := []io.Writer{stderr2}

	if stdout != nil {
		stdoutPipes = append(stdoutPipes, stdout)
	}
	if stderr != nil {
		stderrPipes = append(stderrPipes, stderr)
	}

	if log.GetLevel() == log.DebugLevel {
		stdoutPipes = append(stdoutPipes, os.Stdout)
		stderrPipes = append(stderrPipes, os.Stderr)
	}

	sess.Stdin = stdin
	sess.Stdout = io.MultiWriter(stdoutPipes...)
	sess.Stderr = io.MultiWriter(stderrPipes...)

	log.WithFields(log.Fields{
		"addr": client.addr,
		"cmd":  cmd,
	}).Debug("ssh-run")
	if err := sess.Run(cmd); err != nil {
		if log.GetLevel() != log.DebugLevel {
			io.Copy(os.Stdout, stdout2)
			io.Copy(os.Stderr, stderr2)
		}
		return errors.Wrapf(
			err, "failed to ssh-run %q on %s", cmd, client.addr)
	}

	return nil
}

// Exists checks to see if the provided path exists.
func Exists(
	ctx context.Context,
	client *Client,
	path, operator string) error {

	// Check to see if the path exists.
	if err := Run(ctx, client, nil, nil, nil,
		"sudo sh -c '[ -%s %q ]'", operator, path); err != nil {
		return err
	}

	return nil
}

// FileExists checks to see if a file exists at the provided path.
func FileExists(
	ctx context.Context,
	client *Client,
	path string) error {

	return Exists(ctx, client, path, "f")
}

// DirExists checks to see if a directory exists at the provided path.
func DirExists(
	ctx context.Context,
	client *Client,
	path string) error {

	return Exists(ctx, client, path, "d")
}

// MkdirAll creates all parts of a directory path on the remote host.
func MkdirAll(
	ctx context.Context,
	client *Client,
	dst, owner, group string,
	mode os.FileMode) error {

	// Ensure the destination directory exists.
	if err := Run(ctx, client, nil, nil, nil,
		"sudo mkdir -p %q", dst); err != nil {
		return err
	}

	// Set the directory's owner and group.
	if owner != "" || group != "" {
		if err := Run(ctx, client, nil, nil, nil,
			"sudo chown %s:%s %q", owner, group, dst); err != nil {
			return err
		}
	}

	// Set the directory's mode
	if mode == 0 {
		mode = 0755
	}
	if mode != 0755 {
		if err := Run(ctx, client, nil, nil, nil,
			"sudo chmod %o %q", mode, dst); err != nil {
			return err
		}
	}

	return nil
}

// Upload writes the given data to the remote host at the specified dst.
func Upload(
	ctx context.Context,
	client *Client,
	data []byte,
	dst, owner, group string,
	mode os.FileMode) error {

	dirName := path.Dir(dst)
	fileName := path.Base(dst)

	if err := MkdirAll(ctx, client, dirName, "", "", 0); err != nil {
		return errors.Wrap(err, "failed to create directory for ssh upload")
	}

	pr, pw := io.Pipe()
	errs := make(chan error)

	go func() {
		tw := tar.NewWriter(pw)
		if err := tw.WriteHeader(&tar.Header{
			Name: fileName,
			Mode: int64(mode),
			Size: int64(len(data)),
		}); err != nil {
			errs <- errors.Wrapf(err, "error writing tar header for %q", dst)
			return
		}
		if _, err := tw.Write(data); err != nil {
			errs <- errors.Wrapf(err, "error writing tar data for %q", dst)
			return
		}
		tw.Flush()
		tw.Close()
		pw.Close()
	}()

	go func() {
		defer close(errs)
		errs <- Run(ctx, client, pr, nil, nil, "sudo tar -C %q -x", dirName)
	}()

	return <-errs
}

// KeyGen generates a new SSH key pair.
func KeyGen() ([]byte, []byte, error) {
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
