FROM golang:1.11-alpine AS test
COPY src/ .
RUN CGO_ENABLED=0 go test

FROM golang:1.11-alpine AS build
COPY src/ /go/src/github.com/charter-oss/go-web-service
RUN CGO_ENABLED=0 go build -o /go/bin/go-web-service /go/src/github.com/charter-oss/go-web-service

FROM golang:1.11-alpine
COPY --from=build /go/bin/go-web-service /go/bin/
EXPOSE 8080
ENTRYPOINT [ "/go/bin/go-web-service" ]
