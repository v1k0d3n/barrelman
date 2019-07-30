Flagship: Barrelman Testing
====
---
## Run unit tests
To run barrelman unit tests, you can use the below either ways:
- go test ./...
- make build
- make test

This run will skip barrelman acceptance tests.

## Run acceptance tests
To run barrelman acceptance tests, you can use the below either ways:
- BM_BIN='../barrelman' BM_TEST_E2E='Y' RETRYCOUNTACC=20 INTERVALTIME=1 go test -v -count=1 ./e2e
  * *BM_BIN* is the location value of the barrelman binary. You can even pass barrelman binary from a different location. Developer will have the advantage of maintaining different barrelman release records using different binaries. For example, BM_BIN=/opt/flagship/bin
  * *BM_TEST_E2E* has values 'Y/n' which would enable/disable running of acceptance tests.
  * *RETRYCOUNTACC* is a parameter to allow number of retries until the required count is met, if not, errors out as 'Out of Retries'.
  * *INTERVALTIME* is the sleep interval time in seconds before the next retry.

- make testacc
