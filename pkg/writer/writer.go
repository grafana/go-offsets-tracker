package writer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"

	"github.com/grafana/go-offsets-tracker/pkg/versions"

	"github.com/grafana/go-offsets-tracker/pkg/offsets"
	"github.com/grafana/go-offsets-tracker/pkg/target"
)

func WriteResults(fileName string, results ...*target.Result) error {
	offsets := offsets.Track{
		Data: map[string]offsets.Struct{},
	}
	for _, r := range results {
		convertResult(r, &offsets)
	}

	jsonData, err := json.Marshal(&offsets)
	if err != nil {
		return err
	}

	var prettyJson bytes.Buffer
	err = json.Indent(&prettyJson, jsonData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(fileName, prettyJson.Bytes(), fs.ModePerm)
}

func convertResult(r *target.Result, track *offsets.Track) {
	offsetsMap := make(map[string][]offsets.Versioned)
	for _, vr := range r.ResultsByVersion {
		for _, od := range vr.OffsetData.DataMembers {
			key := fmt.Sprintf("%s,%s", od.StructName, od.Field)
			offsetsMap[key] = append(offsetsMap[key], offsets.Versioned{
				Offset: od.Offset,
				Since:  versions.OrZero(vr.Version).String(),
			})
		}
	}

	// normalize offsets: just annotate the offsets from the version
	// that changed them
	fieldVersionsMap := map[string]hiLoSemVers{}
	for key, offs := range offsetsMap {
		if len(offs) == 0 {
			continue
		}
		// the algorithm below assumes offsets versions are sorted from older to newer
		sort.Slice(offs, func(i, j int) bool {
			return versions.MustParse(offs[i].Since).
				LessThanOrEqual(versions.MustParse(offs[j].Since))
		})

		hilo := hiLoSemVers{}
		var om []offsets.Versioned
		var last offsets.Versioned
		for n, off := range offs {
			hilo.updateModuleVersion(off.Since)
			// only append versions that changed the field value from its predecessor
			if n == 0 || off.Offset != last.Offset {
				om = append(om, off)
			}
			last = off
		}
		offsetsMap[key] = om
		fieldVersionsMap[key] = hilo
	}

	// Append offsets as fields to the existing file map map
	for key, offs := range offsetsMap {
		parts := strings.Split(key, ",")
		strFields, ok := track.Data[parts[0]]
		if !ok {
			strFields = offsets.Struct{}
			track.Data[parts[0]] = strFields
		}
		hl := fieldVersionsMap[key]
		strFields[parts[1]] = offsets.Field{
			Offsets: offs,
			Versions: offsets.VersionInfo{
				Oldest: hl.lo.String(),
				Newest: hl.hi.String(),
			},
		}
	}
}

// hiLoSemVers track highest and lowest version
type hiLoSemVers struct {
	hi *version.Version
	lo *version.Version
}

func (hl *hiLoSemVers) updateModuleVersion(vr string) {
	// if at this point the version does not parse, this means data is downloaded
	// from a branch instead of a tag. Then we default versin to "0.0.0"
	ver := versions.OrZero(vr)

	if hl.lo == nil || ver.LessThan(hl.lo) {
		hl.lo = ver
	}
	if hl.hi == nil || ver.GreaterThan(hl.hi) {
		hl.hi = ver
	}
}
