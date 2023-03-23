# Go offsets tracker

This tool allows generating a JSON description of the byte offsets of a set of field structs,
for each version or Go or third-party library.

It is useful to inspect executable files that do not embed Debug information.

This is a standalone, modified version of the [Open Telemetry offsets tracker tool](https://github.com/open-telemetry/opentelemetry-go-instrumentation),
and both are licensed under [Apache Software License 2.0](./LICENSE).

## How to install

```
go install github.com/grafana/go-offsets-tracker/cmd/go-offsets-tracker@latest
```

## How to run

```
go-offsets-tracker -i examples/input_file.json examples/offsets.json
```