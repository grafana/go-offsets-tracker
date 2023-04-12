package downloader

import (
	_ "embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/grafana/go-offsets-tracker/pkg/utils"
)

const (
	urlPattern = "https://go.dev/dl/go%s.%s-%s.tar.gz"
)

var (
	//go:embed wrapper/gostd.mod.txt
	goSTDMod string
)

func DownloadBinaryFromRemote(inspectFile string, version string) (string, string, error) {
	dir, err := os.MkdirTemp("", version)
	if err != nil {
		return "", "", err
	}
	dest, err := os.Create(path.Join(dir, "go.tar.gz"))
	if err != nil {
		return "", "", err
	}
	defer dest.Close()

	goos, goarch := runtime.GOOS, runtime.GOARCH
	if inspectFile == "" {
		// if we provide the inspection file, we actually need the localhost Go version
		// to execute it as a compile
		goos, goarch = "linux", "amd64"
	}
	// TODO: cache go versions so you don't need to download all of them each time
	resp, err := http.Get(fmt.Sprintf(urlPattern, version, goos, goarch))
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	_, err = io.Copy(dest, resp.Body)
	if err != nil {
		return "", "", err
	}

	output, err := utils.RunCommand("tar -xf go.tar.gz -C .", dir)
	if err != nil {
		log.Println("error uncompressing go.tar.gz:\n", output)
		return "", "", err
	}
	goCMD := fmt.Sprintf("%s/go/bin/go", dir)
	if inspectFile == "" {
		return goCMD, dir, nil
	}
	return compileProvidedFile(version, path.Join(dir, "go"), goCMD, inspectFile)
}

func compileProvidedFile(goVersion, goRootDir, goCMD, inspectFile string) (string, string, error) {
	dir, err := os.MkdirTemp("", appName)
	if err != nil {
		return "", "", err
	}

	minorVersion := strings.Join(strings.Split(goVersion, ".")[:2], ".")
	mod := fmt.Sprintf(goSTDMod, minorVersion)
	if err := os.WriteFile(path.Join(dir, "go.mod"), []byte(mod), fs.ModePerm); err != nil {
		return "", "", fmt.Errorf("creating temporary go.mod file: %w", err)
	}

	mainContents, err := os.ReadFile(inspectFile)
	if err != nil {
		return "", "", fmt.Errorf("reading %s file: %w", inspectFile, err)
	}
	if err := os.WriteFile(path.Join(dir, "main.go"), mainContents, fs.ModePerm); err != nil {
		return "", "", fmt.Errorf("writing main file: %w", err)
	}

	output, err := utils.RunCommand("go mod tidy -compat=1.17", dir)
	if err != nil {
		log.Printf("go mod tidy returned standard error:\n%s", output)
		return "", "", err
	}

	output, err = utils.RunCommand(fmt.Sprintf(`GOROOT="%s" GOOS=linux GOARCH=amd64 %s build`, goRootDir, goCMD), dir)
	if err != nil {
		log.Printf("go build returned standard error:\n%s", output)
		return "", "", err
	}

	return path.Join(dir, appName), dir, nil
}
