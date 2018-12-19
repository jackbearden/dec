FROM golang:1.10.6-alpine3.7

# System setup
RUN apk update && apk add git curl build-base autoconf automake libtool

# Install protoc
ENV PROTOBUF_URL https://github.com/google/protobuf/releases/download/v3.3.0/protobuf-cpp-3.3.0.tar.gz
RUN curl -L -o /tmp/protobuf.tar.gz $PROTOBUF_URL
WORKDIR /tmp/
RUN tar xvzf protobuf.tar.gz
WORKDIR /tmp/protobuf-3.3.0
RUN ./autogen.sh && \
    ./configure && \
    make -j 3 && \
    make check && \
    make install

# Install protoc-gen-go
RUN go get github.com/golang/protobuf/protoc-gen-go

# Install protoactor
RUN go get github.com/gogo/protobuf/protoc-gen-gogoslick && \
    go get github.com/stretchr/testify/assert && \
    go get github.com/AsynkronIT/protoactor-go && \
    cd $GOPATH/src/github.com/AsynkronIT/protoactor-go && \
    go get ./... && \
    make

# Install dec
COPY . /go/src/dec
WORKDIR /go/src/dec

RUN go get github.com/olekukonko/tablewriter && \
    go clean ./... && \
    cd cli && go build && go install && cd .. && \
    cd service && go build && go install && cd ..
