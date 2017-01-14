GERRIT_MAJOR_VERSION ?= 2.13
GERRIT_VERSION ?= $(GERRIT_MAJOR_VERSION).5
GERRIT_SHA1 ?= 4995181ada7c6dc0465c671167ac00f962d18ae6
TAGS := opalmer/gerrittest:latest opalmer/gerrittest:$(GERRIT_VERSION)
BUILD_TAGS := $(addprefix -t , $(TAGS))

build:
	docker build $(BUILD_TAGS) \
	  --build-arg GERRIT_MAJOR_VERSION=$(GERRIT_MAJOR_VERSION) \
	  --build-arg GERRIT_VERSION=$(GERRIT_VERSION) \
	  --build-arg GERRIT_WAR_SHA1=$(GERRIT_SHA1) \
	  .

push:
	docker push opalmer/gerrittest:latest
	docker push opalmer/gerrittest:$(GERRIT_VERSION)
