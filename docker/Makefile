GERRIT_MAJOR_VERSION ?= 2.14
GERRIT_VERSION ?= $(GERRIT_MAJOR_VERSION).2
GERRIT_SHA1 ?= e57c23f17457e14ecfd8da6b2750a94fe65f8970
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
