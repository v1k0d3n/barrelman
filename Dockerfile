FROM golang:1.11

ARG VERSION=latest
ARG COMMIT=master
ARG BRANCH=master
ARG GOOS=linux
ARG GOARCH=amd64

ENV VERSION=$VERSION
ENV COMMIT=$COMMIT
ENV BRANCH=$BRANCH
ENV GOOS=$GOOS
ENV GOARCH=$GOARCH

WORKDIR /go/src/github.com/charter-oss/barrelman
COPY . .

RUN CGO_ENABLED=0 go build -ldflags "-w -s -X github.com/charter-oss/barrelman/version.version=${VERSION} -X github.com/charter-oss/barrelman/version.commit=${COMMIT} -X github.com/charter-oss/barrelman/version.branch=${BRANCH}" -a -installsuffix cgo -o /barrelman

FROM scratch AS build
COPY --from=0 /barrelman /barrelman

VOLUME /data

ENTRYPOINT [ "/barrelman" ]
CMD [ "--help" ]
