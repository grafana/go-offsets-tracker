package binary

import (
	"debug/elf"
	"os"

	"golang.org/x/mod/semver"
)

type DataMember struct {
	StructName string
	Field      string
	MinVersion string
	MaxVersion string
}

type DataMemberOffset struct {
	*DataMember
	Offset uint64
}

type Result struct {
	DataMembers []*DataMemberOffset
}

type ErrOffsetsNotFound struct {
	fieldName string
}

func (e *ErrOffsetsNotFound) Error() string {
	return "could not find offsets for " + e.fieldName
}

func FindOffsets(version string, file *os.File, dataMembers []*DataMember) (*Result, error) {
	elfF, err := elf.NewFile(file)
	if err != nil {
		return nil, err
	}

	dwarfData, err := elfF.DWARF()
	if err != nil {
		return nil, err
	}

	result := &Result{}
	for _, dm := range dataMembers {
		if dm.MinVersion != "" {
			if semver.Compare(version, "v"+dm.MinVersion) < 0 {
				continue
			}
		} else if dm.MaxVersion != "" {
			if semver.Compare(version, "v"+dm.MaxVersion) > 0 {
				continue
			}
		}

		offset, found := findDataMemberOffset(dwarfData, dm)
		if !found {
			return nil, &ErrOffsetsNotFound{fieldName: dm.StructName + " " + dm.Field}
		} else {
			result.DataMembers = append(result.DataMembers, &DataMemberOffset{
				DataMember: dm,
				Offset:     uint64(offset),
			})
		}
	}

	return result, nil
}
