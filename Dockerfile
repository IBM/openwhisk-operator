# Use docker multi-stage build to create the Openwhisk Controller image
FROM golang:1.10.3 as builder

# Copy in the go src
WORKDIR /go/src/github.com/ibm/cloud-functions-operator
COPY pkg/    pkg/
COPY cmd/    cmd/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager github.com/ibm/cloud-functions-operator/cmd/manager

FROM ubuntu:latest
RUN apt-get update && apt-get install --no-install-recommends -y ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /root/
COPY --from=builder /go/src/github.com/ibm/cloud-functions-operator/manager .
ENTRYPOINT ["./manager"]
