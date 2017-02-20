SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')
ODIR=bin
BINARY=babelfish-go-driver
DRIVER_VERSION=beta-demo-0.0.9
LDFLAGS=-ldflags "-X main.driverVersion=${DRIVER_VERSION}"

.DEFAULT_GOAL: $(BINARY)

$(BINARY): $(SOURCES)
	if [ ! -d ${ODIR} ]; then mkdir -p ${ODIR} ; fi
	go build ${LDFLAGS} -o ${ODIR}/${BINARY} ${SOURCEDIR}

.PHONY: clean
clean:
	if [ -f ${ODIR}/${BINARY} ] ; then rm ${ODIR}/${BINARY} ; fi
