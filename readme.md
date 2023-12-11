# util

The code in this project is designed to be reusable across many different Go projects. It's designed to be completely general and can be reused without any adaptation.

This library calls `panic` in a number of places so you may want to capture both stdout and stderr like this:

```
go run ./fileserver/main.go 2>stderr.log 1>stdout.log
```

To run all tests:

```
go test $(go list ./... | grep -v manual)
```

For an overview of this package see https://pkg.go.dev/github.com/1f604/util

The AWS dependencies are only needed for the cloudflare stuff.