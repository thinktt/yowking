# JS_FILES := $(wildcard *.txt)
# BINS := $(JS_FILES:%.txt=%)

.PHONY: run docal getclocks dbuild drun din dcal clean books2bin \
	rmbooks  reset eep testengine cpbadbooks push gobuild dexec

gobuild:
	GOOS=windows GOARCH=386 go build -o dist/enginewrap.exe  ./cmd/enginewrap 
	GOOS=linux go build -o dist/kingapi  ./cmd/kingapi

dbuild: dist
	docker rm yowking || true
	docker image rm ace:5000/yowking:latest || true
	docker build -t ace:5000/yowking .
	# docker push ace:5000/yowking  

dist:
	mkdir dist 
	make gobuild
	cp assets/* dist/

clean: 
	rm -rf dist

drun: dbuild
	docker run --rm -it --name yowking ace:5000/yowking

dexec: dbuild
	docker run --rm -it --name yowking ace:5000/yowking /bin/bash

run: export IS_WSL=true
run: 
	go run ./cmd/kingapi/main.go