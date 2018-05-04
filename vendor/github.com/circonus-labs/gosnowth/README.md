# gosnowth

The Go Snowth Client.  This codebase contains client code for accessing Snowth
APIs.  IRONdb consists of a multitude of Snowth Nodes which can be queried
directly through an exposed HTTP API documented in the IRONdb API documentation:

https://github.com/circonus/irondb-docs/blob/master/api.md

Each of the documented APIs are being implemented at methods of the SnowthClient
structure defined in this repository.  In order to see the documentation of each
of the methods, you can use the `godoc` tool to autogenerate the documentation
shown below:

```bash
godoc github.com/circonus/gosnowth # plaintext
godoc -html github.com/circonus/gosnowth # html output
```

## Testing

In order to test this package, run the go unit tests:

```bash
go test github.com/circonus/gosnowth # run package unit tests
```

## Using

In order to use this package, you can follow the examples in the `cmd/example`
sub-package which shows how you would instantiate a new SnowthClient, as well
as how to use the SnowthClient to operate on SnowthNodes.

