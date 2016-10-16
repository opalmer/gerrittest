GERRIT_MAJOR_VERSION ?= 2.13
GERRIT_VERSION ?= $(GERRIT_MAJOR_VERSION).1
GERRIT_SHA1 ?= 1db213eb013c6555e582876ebd3d845066f1ef66
TAGS := opalmer/gerrittest:latest opalmer/gerrittest:$(GERRIT_VERSION)


build:
	docker build -t opalmer/gerrittest:latest -t opalmer/gerrittest:$(GERRIT_VERSION) \
	  --build-arg GERRIT_MAJOR_VERSION=$(GERRIT_MAJOR_VERSION) \
	  --build-arg GERRIT_VERSION=$(GERRIT_VERSION) \
	  .