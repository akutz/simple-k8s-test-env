// +build none

package app_test

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"vmw.io/sk8/app"
	"vmw.io/sk8/config"
)

func TestGetCloudInitUserData(t *testing.T) {
	ctx := context.Background()
	c := config.Config{
		Users: []config.UserConfig{
			config.UserConfig{
				Name:         "akutz",
				SSHPublicKey: "akutz-ssh-pub-key",
			},
			config.UserConfig{
				Name:         "luoh",
				SSHPublicKey: "luoh-ssh-pub-key",
			},
			config.UserConfig{
				Name:         "fabio",
				SSHPublicKey: "fabio-ssh-pub-key",
			},
		},
	}
	app.ValidateConfig(ctx, &c)
	buf, err := app.GetCloudInitUserData(
		ctx, c, fmt.Sprintf("c01.%s", c.Network.DomainFQDN), "both")
	if err != nil {
		t.Fatal(err)
	}
	var w io.Writer = ioutil.Discard
	if testing.Verbose() {
		w = os.Stdout
	}
	w.Write(buf)
}

func TestGetCloudInitMetaData(t *testing.T) {
	ctx := context.Background()
	c := config.Config{}
	app.ValidateConfig(ctx, &c)
	netBuf, err := app.GetCloudInitNetworkData(ctx, c)
	if err != nil {
		t.Fatal(err)
	}
	buf, err := app.GetCloudInitMetaData(
		ctx, c, netBuf, fmt.Sprintf("c01.%s", c.Network.DomainFQDN))
	if err != nil {
		t.Fatal(err)
	}
	var w io.Writer = ioutil.Discard
	if testing.Verbose() {
		w = os.Stdout
	}
	w.Write(netBuf)
	w.Write(buf)
}
