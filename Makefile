
check: build test

build:
	$(MAKE) -C docker build
	pip install -e .

test:
	./test.sh
