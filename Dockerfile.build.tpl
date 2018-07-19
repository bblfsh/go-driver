FROM golang:1.10-stretch

RUN mkdir -p /opt/driver/src && \
    useradd --uid ${BUILD_UID} --home /opt/driver ${BUILD_USER}

RUN apt update && \
    apt install -y --no-install-recommends make git curl ca-certificates

WORKDIR /opt/driver/src
