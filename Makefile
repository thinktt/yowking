# JS_FILES := $(wildcard *.txt)
# BINS := $(JS_FILES:%.txt=%)

.PHONY: run docal getclocks dbuild drun din dcal clean books2bin \
	rmbooks  reset eep testengine cpbadbooks push gobuild dexec \
	getassets

gobuild:
	GOOS=windows GOARCH=386 go build -o dist/enginewrap.exe  ./cmd/enginewrap 
	GOOS=linux go build -o dist/kingapi  ./cmd/kingapi

dbuild: dist
	docker rm yowking || true
	docker image rm ace:5000/yowking:latest || true
	docker build -t ace:5000/yowking .
	# docker push ace:5000/yowking  

dist: assets
	mkdir dist 
	cp assets/* dist/
	make gobuild

clean: 
	rm -rf assets
	rm -rf dist

drun: dbuild
	docker run --rm -it --name yowking ace:5000/yowking

dexec: dbuild
	docker run --rm -it --name yowking ace:5000/yowking /bin/bash

run: export IS_WSL=true
run: dist
	cd dist && 	go run ../cmd/kingapi

# later we will import and build these from CM11 folder, for now borrowed locally
assets:
	mkdir assets
	# cp -r ../yeoldwiz-lnx/yowbot/dist/books assets/books
	cp ../yeoldwiz-lnx/yowbot/dist/calibrations/clockTimes.json assets/clockTimes.json
	cp ../yeoldwiz-lnx/yowbot/dist/TheKing350noOpk.exe assets/TheKing350noOpk.exe
	cp ../yeoldwiz-lnx/yowbot/dist/personalities.json assets/personalities.json