package downloader

import (
	_ "embed"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path"
	"text/template"

	"github.com/grafana/go-offsets-tracker/pkg/utils"
)

const appName = "testapp"

var (
	//go:embed wrapper/go.mod.txt
	goMod string

	//go:embed wrapper/main.go.txt
	goMain string
)

func DownloadBinary(modName string, version string, inspectFile string, packages []string) (string, string, error) {
	dir, err := ioutil.TempDir("", appName)
	if err != nil {
		return "", "", err
	}

	goModContent := fmt.Sprintf(goMod, modName, version)
	err = ioutil.WriteFile(path.Join(dir, "go.mod"), []byte(goModContent), fs.ModePerm)
	if err != nil {
		return "", "", err
	}

	if inspectFile == "" {
		mainFile, err := os.OpenFile(path.Join(dir, "main.go"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fs.ModePerm)
		if err != nil {
			return "", "", fmt.Errorf("can't create main.go file: %w", err)
		}
		defer mainFile.Close()
		tmpl, err := template.New("main-file").Parse(goMain)
		if err != nil {
			panic(err)
		}
		// If no explicit packages are provided, we render the main.go import with the module name.
		if len(packages) == 0 {
			packages = []string{modName}
		}
		if err := tmpl.Execute(mainFile, packages); err != nil {
			panic(err)
		}
	} else {
		mainFile, err := os.OpenFile(path.Join(dir, "main.go"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fs.ModePerm)
		if err != nil {
			panic(err)
		}
		defer mainFile.Close()

		sourceFile, err := os.Open(inspectFile)
		if err != nil {
			panic(err)
		}
		defer sourceFile.Close()

		_, err = io.Copy(mainFile, sourceFile)
		if err != nil {
			panic(err)
		}
	}

	output, err := utils.RunCommand("go mod tidy -compat=1.17", dir)
	if err != nil {
		log.Println("go mod tidy returned error: \n", output)
		return "", "", err
	}

	output, err = utils.RunCommand("GOOS=linux GOARCH=amd64 go build", dir)
	if err != nil {
		log.Println("go build returned error: \n", output)
		return "", "", err
	}

	return path.Join(dir, appName), dir, nil
}
