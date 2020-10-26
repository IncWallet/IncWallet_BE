FROM golang:1.14.6

ENV GOPATH /go
ENV GOBIN /go/bin

# Move current project to a valid go path
COPY . /go/src/github.com/incwallet
WORKDIR /go/src/github.com/incwallet

# Install Revel CLI
RUN go get github.com/revel/cmd/revel

COPY go.mod .
COPY go.sum .
RUN go mod download

WORKDIR /go/src/github.com/incwallet/app/wic
RUN go install

WORKDIR /go/src/github.com/incwallet

# Run app in production mode
EXPOSE 9000
ENTRYPOINT revel run github.com/incwallet dev 9000