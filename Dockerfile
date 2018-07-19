# Prerequisites:
#   dep ensure --vendor-only
#   bblfsh-sdk release

#==============================
# Stage 1: Native Driver Build
#==============================
FROM golang:1.10-alpine as native

ENV DRIVER_REPO=github.com/bblfsh/go-driver
ENV DRIVER_REPO_PATH=/go/src/$DRIVER_REPO

ADD vendor $DRIVER_REPO_PATH/vendor
ADD driver $DRIVER_REPO_PATH/driver
ADD native $DRIVER_REPO_PATH/native
WORKDIR $DRIVER_REPO_PATH/native

# build native driver
RUN go build -o /tmp/native native.go


#================================
# Stage 1.1: Native Driver Tests
#================================
FROM native as native_test
# run native driver tests
RUN go test ../driver/golang/...


#=================================
# Stage 2: Go Driver Server Build
#=================================
FROM native as driver

WORKDIR $DRIVER_REPO_PATH/

# build tests
RUN go test -c -o /tmp/fixtures.test ./driver/fixtures/
# build server binary
RUN go build -o /tmp/driver ./driver/main.go

#=======================
# Stage 3: Driver Build
#=======================
FROM golang:1.10-alpine

LABEL maintainer="source{d}" \
      bblfsh.language="go"

WORKDIR /opt/driver

# copy driver manifest and static files
ADD .manifest.release.toml ./etc/manifest.toml

# copy build artifacts for native driver
COPY --from=native /tmp/native ./bin/


# copy tests binary
COPY --from=driver /tmp/fixtures.test ./bin/
# move stuff to make tests work
RUN ln -s /opt/driver ../build
VOLUME /opt/fixtures

# copy driver server binary
COPY --from=driver /tmp/driver ./bin/

ENTRYPOINT ["/opt/driver/bin/driver"]