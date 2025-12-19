.PHONY: build install uninstall clean test

build:
	go build -o gkc ./cmd/kubectx/*
	go build -o gkn ./cmd/kubens/*

install:
	go install ./cmd/...
	cp ./completion/*.zsh ~/zsh-site-functions/

uninstall:
	rm "${GOBIN}/kubectx"
	rm "${GOBIN}/kubens"
	rm ~/zsh-site-functions/_kubectx.zsh
	rm ~/zsh-site-functions/_kubens.zsh

clean:
	rm gkc gkn

test:
	go test ./...
