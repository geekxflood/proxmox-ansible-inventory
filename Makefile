IMAGE_NAME = $(shell basename $(shell pwd))
TAG_LATEST = $(IMAGE_NAME):latest

.PHONY: all image binary clean

all: image build

image:
	docker build . \
		-t $(TAG_LATEST) 

build:
	mkdir -p binary
	go build -a  \
		-gcflags=all="-l -B" \
		-ldflags="-w -s" \
		-o binary/$(IMAGE_NAME) \
		./...

run: build
	binary/$(IMAGE_NAME)

clean:
	rm -rf binary
