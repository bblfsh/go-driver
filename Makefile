SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')
ODIR=bin
BINARY=babelfish-go-driver
DRIVER_VERSION=beta-demo-0.0.9
LDFLAGS=-ldflags "-X main.driverVersion=${DRIVER_VERSION}"
DOCKERFILE=Dockerfile
DOCKER_IMAGE_NAME=babelfish-go-driver

all: $(BINARY) build 

$(BINARY): $(SOURCES)
	if [ ! -d ${ODIR} ]; then mkdir -p ${ODIR} ; fi
	go build ${LDFLAGS} -o ${ODIR}/${BINARY} ${SOURCEDIR}


build: $(DOCKERFILE)
	docker build -f ${DOCKERFILE} -t ${DOCKER_IMAGE_NAME} ${SOURCEDIR} 

.PHONY:
clean:
	if [ -f ${ODIR}/${BINARY} ] ; then rm ${ODIR}/${BINARY} ; fi
