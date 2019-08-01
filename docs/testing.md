# Flagship: Barrelman Testing

### Pre-requisite:
Flagship cluster should be in running and stable state.
Tiller should be installed on your cluster.
Kubectl should be set in your PATH.

## Run unit tests
To run Barrelman unit tests, you can use the below command:
```
- go test ./...
- make test
```

This run will skip Barrelman acceptance tests.

## Run acceptance tests
To run Barrelman acceptance tests, you can use the below command:
```
- make testacc
```
