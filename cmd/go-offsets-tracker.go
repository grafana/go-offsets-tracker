package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/grafana/go-offsets-tracker/pkg/schema"

	"github.com/hashicorp/go-version"

	"github.com/grafana/go-offsets-tracker/pkg/binary"
	"github.com/grafana/go-offsets-tracker/pkg/target"
	"github.com/grafana/go-offsets-tracker/pkg/writer"
)

var (
	inputFile = flag.String("i", "", "input JSON file with the required offsets definition")
	help      = flag.Bool("h", false, "shows this help")
)

func showHelp(isErr bool) {
	fmt.Println("usage: go-offsets-tracker -i <input file> <output file>")
	flag.PrintDefaults()
	if isErr {
		os.Exit(2)
	}
	os.Exit(0)
}

func main() {
	flag.Parse()
	outFile := flag.Arg(0)
	if help != nil && *help || outFile == "" || inputFile == nil || *inputFile == "" {
		showHelp(help == nil || !*help)
	}

	inputBytes, err := os.ReadFile(*inputFile)
	exitOnErr(err, "reading input file")

	ilibs := schema.InputLibs{}
	exitOnErr(
		json.Unmarshal(inputBytes, &ilibs),
		"parsing input file")

	var libs []*target.Result
	if std := processGoStdlib(ilibs, outFile); std != nil {
		libs = append(libs, std)
	}

	for k, v := range ilibs {
		if k == schema.GoStdLib {
			continue
		}
		if l := processThirdPartyLib(k, v, outFile); l != nil {
			libs = append(libs, l)
		}
	}

	log.Println("Done collecting offsets, writing results to file ...")
	err = writer.WriteResults(outFile, libs...)
	if err != nil {
		log.Fatalf("error while writing results to file: %v\n", err)
	}

	log.Println("Done!")
}

func processGoStdlib(input schema.InputLibs, outFileName string) *target.Result {
	goLib, ok := input[schema.GoStdLib]
	if !ok {
		return nil
	}
	minimunGoVersion, err := version.NewConstraint(goLib.Versions)
	exitOnErr(err, "invalid Go version constraint")

	stdLibOffsets, err := target.New("go", outFileName).
		FindVersionsBy(target.GoDevFileVersionsStrategy).
		DownloadBinaryBy(target.DownloadPreCompiledBinaryFetchStrategy).
		VersionConstraint(&minimunGoVersion).
		FindOffsets(fieldsAsDataMembers(goLib.Fields))
	exitOnErr(err, "loading Go standard library offsets")
	return stdLibOffsets
}

func processThirdPartyLib(name string, lib schema.LibQuery, outFileName string) *target.Result {
	tData := target.New(name, outFileName)

	if lib.Versions != "" {
		minVersion, err := version.NewConstraint(lib.Versions)
		exitOnErr(err, "invalid Lib version constraint")
		tData = tData.VersionConstraint(&minVersion)
	}

	libOffsets, err := tData.FindOffsets(fieldsAsDataMembers(lib.Fields))
	exitOnErr(err, "loading "+name+" offsets")
	return libOffsets
}

// Function kept to keep interfaces' and types compatibility with old version
// TODO: remove DataMember type and use the simple map form
func fieldsAsDataMembers(fields map[string][]string) []*binary.DataMember {
	var out []*binary.DataMember
	for structName, fieldsList := range fields {
		for _, fieldName := range fieldsList {
			out = append(out, &binary.DataMember{
				StructName: structName,
				Field:      fieldName,
			})
		}
	}
	return out
}

func exitOnErr(err error, str string) {
	if err != nil {
		log.Printf("ERROR: %s: %s", str, err.Error())
		os.Exit(1)
	}
}
