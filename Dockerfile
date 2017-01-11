FROM blueprintprotoc

ENV GO_VERSION 1.6.2
ENV GO_SRC_URL http://golang.org/dl/go$GO_VERSION.src.tar.gz
ENV GO_SRC_SHA1 787b0b750d037016a30c6ed05a8a70a91b2e9db4bd9b1a2453aa502a63f1bccc

ENV GO_BOOTSTRAP_VERSION 1.4.3
ENV GO_BOOTSTRAP_URL http://storage.googleapis.com/golang/go$GO_BOOTSTRAP_VERSION.src.tar.gz
ENV GO_BOOTSTRAP_SHA1 486db10dc571a55c8d795365070f66d343458c48

ENV CGO_ENABLED 0

# Install protoc go generator
###################################
RUN mkdir -p /go
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN set -ex \
	&& apk --no-cache add --virtual .scm \
  bash \
  musl-dev \
  git \
  gcc \
	openssl \
  binutils \
  rsync \
  build-base \
	&& mkdir -p /usr/local/bootstrap \
	&& wget -q "$GO_BOOTSTRAP_URL" -O golang.tar.gz \
	&& echo "$GO_BOOTSTRAP_SHA1  golang.tar.gz" | sha1sum -c - \
	&& tar -C /usr/local/bootstrap -xzf golang.tar.gz \
	&& rm golang.tar.gz \
	&& cd /usr/local/bootstrap/go/src \
	&& ./make.bash \
	&& export GOROOT_BOOTSTRAP=/usr/local/bootstrap/go \
	\
	&& wget -q "$GO_SRC_URL" -O golang.tar.gz \
	&& echo "$GO_SRC_SHA1  golang.tar.gz" | sha256sum -c - \
	&& tar -C /usr/local -xzf golang.tar.gz \
	&& rm golang.tar.gz \
	&& cd /usr/local/go/src \
	&& ./make.bash

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

RUN go get -a github.com/golang/protobuf/protoc-gen-go 
RUN go get -a github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
RUN go get -a github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
RUN go get -a github.com/nstogner/protoc-gen-grpc-goservice
RUN go get -a github.com/ghodss/yaml
# RUN go get -a k8s.io/client-go/...
RUN git clone https://github.com/kubernetes/client-go $GOPATH/src/k8s.io/client-go
RUN git -C $GOPATH/src/k8s.io/client-go checkout release-2.0
# Remove this example because it gets in the way of running examples
# through the bp command
RUN rm -rf  $GOPATH/src/github.com/nstogner/protoc-gen-grpc-goservice/example
RUN go get -a golang.org/x/tools/imports

RUN mkdir /tmpl
COPY . $GOPATH/src/github.com/nstogner/blueprint
COPY ./templates /tmpl
RUN go install github.com/nstogner/blueprint/cmd/bp-in-docker
#RUN apk del .scm
#RUN rm -rf /usr/local/bootstrap
#RUN rm -rf /var/cache/apk/*

RUN mkdir /repo
WORKDIR /repo
ENV GOPATH /repo/go

#ENTRYPOINT ["bp-in-docker"]
