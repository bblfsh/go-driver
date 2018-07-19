FROM debian:jessie-slim
MAINTAINER source{d}

ADD build /opt/driver
ENTRYPOINT /opt/driver/bin/driver
