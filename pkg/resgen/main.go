package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/vmware/sk8/pkg/app"
)

var projectRoot string

func main() {
	flag.StringVar(
		&projectRoot,
		"project-root",
		"",
		"The root directory of the sk8 project")
	flag.Parse()

	sk8Path := path.Join(projectRoot, "sk8.sh")
	sk8Base64, err := app.Base64GzipFile(sk8Path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	sk8ResGoPath := path.Join(projectRoot, "pkg", "app", "resgen_sk8.go")
	sk8ResGoW, err := os.Create(sk8ResGoPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	sk8ResGoTpl := template.Must(template.New("sk8").Parse(sk8ResGoTplData))
	if err := sk8ResGoTpl.Execute(sk8ResGoW, struct {
		Data string
	}{
		sk8Base64,
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

const sk8ResGoTplData = `package app // import "github.com/vmware/sk8/pkg/app"

func init() {
	sk8ScriptRes = ` + "`" + `{{.Data}}` + "`" + `
}
`
