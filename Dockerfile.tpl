FROM alpine:3.6
MAINTAINER source{d}

ADD build /opt/driver
ENTRYPOINT /opt/driver/bin/driver
