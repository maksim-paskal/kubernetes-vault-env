FROM golang:1.12 as build

COPY main.go /usr/src/kubernetes-vault/main.go
COPY go.mod /usr/src/kubernetes-vault/go.mod
COPY go.sum /usr/src/kubernetes-vault/go.sum

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

RUN cd /usr/src/kubernetes-vault \
  && go mod download \
  && go mod verify \
  && go build -v -o kubernetes-vault -ldflags "-X main.buildTime=$(date +"%Y%m%d%H%M%S")"

FROM alpine:3.10

COPY --from=build /usr/src/kubernetes-vault/kubernetes-vault /usr/local/bin/kubernetes-vault

CMD /usr/local/bin/kubernetes-vault