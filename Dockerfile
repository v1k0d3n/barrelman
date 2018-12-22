FROM golang:1.11

ARG GOOS=linux
ARG GOARCH=amd64

ENV GOOS=$GOOS
ENV GOARCH=$GOARCH

WORKDIR /go/src/github.com/charter-se/barrelman
COPY . .

RUN go get -d -v ./...

RUN CGO_ENABLED=0 go build -ldflags '-w -s' -a -installsuffix cgo -o /barrelman

FROM scratch AS build
COPY --from=0 /barrelman /barrelman

VOLUME /data

ENTRYPOINT [ "/barrelman" ]
CMD [ "--help" ]
