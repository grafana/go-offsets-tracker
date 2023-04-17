# Change log

## v0.1.4
* Fixes a crash when trying to regenerate an offsets file containing a non-semantic branch name.

## v0.1.3
* Changes in the [input file JSON schema](examples/input_file.json):
  - The `"branch"` property will override the `"versions"` constraint and will directly
    download the offsets from a given branch. This is useful for repositories not having
    any release tag. The downloaded version will be set as `0.0.0`.
  - The `"packages"` property will explicitly inspect the provided list of packages.
    If empty, the program downloads and inspects the module name as package. This is useful
    for modules without any package in the root folder.

## v0.1.2
* Enabled optional `inspect` field in input file

## v0.1.1

* Added Go handling methods to load and search offsets within the file

## v0.1.0

* Initial release, mimicking the functionality of [this PR in the Open Telemetry repository](https://github.com/open-telemetry/opentelemetry-go-instrumentation/pull/45).
* Plus:
  * Allow specifying source offsets and constraints as an external JSON file, instead of hardcoding it.
