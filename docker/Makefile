GERRIT_MAJOR_VERSION ?= 2.13
GERRIT_VERSION ?= $(GERRIT_MAJOR_VERSION).1
GERRIT_SHA1 ?= 1db213eb013c6555e582876ebd3d845066f1ef66
TAGS := opalmer/gerrittest:latest opalmer/gerrittest:$(GERRIT_VERSION)
BUILD_TAGS := $(addprefix -t , $(TAGS))


build:
	docker build $(BUILD_TAGS) \
	  --build-arg GERRIT_MAJOR_VERSION=$(GERRIT_MAJOR_VERSION) \
	  --build-arg GERRIT_VERSION=$(GERRIT_VERSION) \
	  .

push:
	docker push opalmer/gerrittest:latest
	docker push opalmer/gerrittest:$(GERRIT_VERSION)
