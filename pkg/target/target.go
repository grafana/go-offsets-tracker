package target

import (
	"fmt"
	"os"
	"strings"

	"github.com/grafana/go-offsets-tracker/pkg/offsets"

	"github.com/hashicorp/go-version"

	"github.com/grafana/go-offsets-tracker/pkg/binary"
	"github.com/grafana/go-offsets-tracker/pkg/cache"
	"github.com/grafana/go-offsets-tracker/pkg/downloader"
	"github.com/grafana/go-offsets-tracker/pkg/versions"
)

type VersionsStrategy int
type BinaryFetchStrategy int

const (
	GoListVersionsStrategy    VersionsStrategy = 0
	GoDevFileVersionsStrategy VersionsStrategy = 1

	WrapAsGoAppBinaryFetchStrategy         BinaryFetchStrategy = 0
	DownloadPreCompiledBinaryFetchStrategy BinaryFetchStrategy = 1
)

type Result struct {
	ModuleName       string
	ResultsByVersion []*VersionedResult
}

type VersionedResult struct {
	Version    string
	OffsetData *binary.Result
}

type targetData struct {
	name                string
	VersionsStrategy    VersionsStrategy
	BinaryFetchStrategy BinaryFetchStrategy
	packages            []string
	branch              string
	versionConstraint   *version.Constraints
	Cache               *cache.Cache
}

func New(name string, fileName string) *targetData {
	return &targetData{
		name:                name,
		VersionsStrategy:    GoListVersionsStrategy,
		BinaryFetchStrategy: WrapAsGoAppBinaryFetchStrategy,
		Cache:               cache.NewCache(fileName),
	}
}

func (t *targetData) Packages(names []string) *targetData {
	t.packages = names
	return t
}

func (t *targetData) Branch(branchName string) *targetData {
	t.branch = branchName
	return t
}

func (t *targetData) VersionConstraint(constraint *version.Constraints) *targetData {
	t.versionConstraint = constraint
	return t
}

func (t *targetData) FindVersionsBy(strategy VersionsStrategy) *targetData {
	t.VersionsStrategy = strategy
	return t
}

func (t *targetData) DownloadBinaryBy(strategy BinaryFetchStrategy) *targetData {
	t.BinaryFetchStrategy = strategy
	return t
}

func (t *targetData) FindOffsets(goLib offsets.LibQuery) (*Result, error) {

	dm := fieldsAsDataMembers(goLib.Fields)

	var vers []string
	if t.branch != "" {
		vers = []string{t.branch}
	} else {
		fmt.Printf("%s: Discovering available versions\n", t.name)
		var err error
		vers, err = t.findVersions()
		if err != nil {
			return nil, err
		}
	}

	result := &Result{
		ModuleName: t.name,
	}
	for _, v := range vers {
		if t.Cache != nil {
			cachedResults, found := t.Cache.IsAllInCache(v, dm)
			if found {
				fmt.Printf("%s: Found all requested offsets in cache for version %s\n", t.name, v)
				result.ResultsByVersion = append(result.ResultsByVersion, &VersionedResult{
					Version: v,
					OffsetData: &binary.Result{
						DataMembers: cachedResults,
					},
				})
				continue
			}
		}

		fmt.Printf("%s: Downloading version %s\n", t.name, v)
		exePath, dir, err := t.downloadBinary(t.name, goLib.Inspect, v)
		if err != nil {
			return nil, err
		}

		fmt.Printf("%s: Analyzing binary for version %s\n", t.name, v)
		res, err := t.analyzeFile(v, exePath, dm)
		if err != nil {
			return nil, fmt.Errorf("%s (version: %s): %w", t.name, v, err)
		} else {
			result.ResultsByVersion = append(result.ResultsByVersion, &VersionedResult{
				Version:    v,
				OffsetData: res,
			})
		}

		os.RemoveAll(dir)
	}

	return result, nil
}

func parseFieldName(f string) (string, string, string) {
	if strings.HasPrefix(f, "[") {
		l := strings.Index(f, "]")

		if l > 0 {
			versionsStr := f[1:l]

			if len(versionsStr) > 0 {
				versions := strings.Split(versionsStr, ",")

				if len(versions) > 0 {
					if len(versions) > 1 {
						return f[l+1:], versions[0], versions[1]
					} else {
						return f[l+1:], versions[0], ""
					}
				}
			}
		}
	}

	return f, "", ""
}

// Function kept to keep interfaces' and types compatibility with old version
// TODO: remove DataMember type and use the simple map form
func fieldsAsDataMembers(fields map[string][]string) []*binary.DataMember {
	var out []*binary.DataMember

	for structName, fieldsList := range fields {
		for _, fieldName := range fieldsList {
			field, minVer, maxVer := parseFieldName(fieldName)

			out = append(out, &binary.DataMember{
				StructName: structName,
				Field:      field,
				MinVersion: minVer,
				MaxVersion: maxVer,
			})
		}
	}
	return out
}

func (t *targetData) analyzeFile(version, exePath string, dm []*binary.DataMember) (*binary.Result, error) {
	f, err := os.Open(exePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	res, err := binary.FindOffsets(version, f, dm)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (t *targetData) findVersions() ([]string, error) {
	var vers []string
	var err error
	if t.VersionsStrategy == GoListVersionsStrategy {
		vers, err = versions.FindVersionsUsingGoList(t.name)
		if err != nil {
			return nil, err
		}
	} else if t.VersionsStrategy == GoDevFileVersionsStrategy {
		vers, err = versions.FindVersionsFromGoWebsite()
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("unsupported version strategy")
	}

	if t.versionConstraint == nil {
		return vers, nil
	}

	var filteredVers []string
	for _, v := range vers {
		semver, err := version.NewVersion(v)
		if err != nil {
			return nil, err
		}

		if t.versionConstraint.Check(semver) {
			filteredVers = append(filteredVers, v)
		}
	}

	if len(filteredVers) == 0 {
		return nil, fmt.Errorf("no tags found for %q. Try expanding the constraint or set the name of a given branch", t.name)
	}
	return filteredVers, nil
}

func (t *targetData) downloadBinary(modName, inspectFile, version string) (string, string, error) {
	if t.BinaryFetchStrategy == WrapAsGoAppBinaryFetchStrategy {
		return downloader.DownloadBinary(modName, version, t.packages)
	} else if t.BinaryFetchStrategy == DownloadPreCompiledBinaryFetchStrategy {
		return downloader.DownloadBinaryFromRemote(inspectFile, version)
	}

	return "", "", fmt.Errorf("unsupported binary fetch strategy")
}
