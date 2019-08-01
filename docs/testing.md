# Flagship: Barrelman Testing

### Prerequisites
Kubernetes cluster should be in running and stable state.
Tiller should be installed on the cluster.
Kubectl should be set in the PATH.

## Run unit tests
To run Barrelman unit tests, use the below command:
```
make test
```

This run will skip Barrelman acceptance tests.

## Run acceptance tests
To run Barrelman acceptance tests, use the below command:
```
make testacc
```
