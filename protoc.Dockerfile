FROM alpine:3.4

RUN apk --update --no-cache add bash
RUN apk --no-cache add --virtual .pb-build make cmake autoconf automake curl tar libtool g++

RUN mkdir -p /tmp/protobufs/protobuf
WORKDIR /tmp/protobufs
RUN curl -o protobufs.tar.gz -L https://github.com/google/protobuf/releases/download/v3.0.0/protobuf-cpp-3.0.0.tar.gz
RUN tar -zxvf protobufs.tar.gz -C /tmp/protobufs/protobuf --strip-components=1

WORKDIR /tmp/protobufs/protobuf
RUN ./autogen.sh
RUN ./configure --prefix=/usr
RUN make
RUN make install

WORKDIR /
RUN rm -rf /tmp/protobufs
RUN rm -rf /tmp/protobufs.tar.gz

RUN apk --no-cache add libstdc++
RUN apk del .pb-build
RUN rm -rf /var/cache/apk/*
