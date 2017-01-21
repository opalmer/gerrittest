
check: build test

build:
	$(MAKE) -C docker build
	pip install -e .

test:
	gerrittest --log-level debug self-test
