SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

APP_NAME = sca
APP_VERSION ?= `git describe --tags --abbrev=0`
APP_BUILDTIME = `date -u +%FT%T%z`
GIT_HASH = `git rev-parse --short HEAD`
GIT_BRANCH = `git rev-parse --abbrev-ref HEAD`

LDFLAGS = \
  -s -w \
  -X main.Version=$(APP_VERSION) -X main.Branch=$(GIT_BRANCH) -X main.Commit=$(GIT_HASH) -X main.BuildTime=$(APP_BUILDTIME)

.DEFAULT_GOAL: $(APP_NAME)

$(APP_NAME): $(SOURCES)
	go build -ldflags "${LDFLAGS}" -o ${APP_NAME} main.go

.PHONY: install
install:
	go install -ldflags "${LDFLAGS}" ./...

.PHONY: clean
clean:
	if [ -f ${APP_NAME} ] ; then rm ${APP_NAME} ; fi
