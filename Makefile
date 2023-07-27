# JS_FILES := $(wildcard *.txt)
# BINS := $(JS_FILES:%.txt=%)

SHELL := /bin/bash

.PHONY: run docal getclocks dbuild drun din dcal clean books2bin \
	rmbooks  reset eep testengine cpbadbooks push gobuild dexec \
	getassets test startk8 dbuild2 dbuild3 dodcal doservice  \
	compose-yowking.yaml up down

export CPU1=10
export CPU2=11
export LOGICAL_PROCESSOR=${CPU1},${CPU2}
export NAME=yowking-${CPU1}${CPU2}
export VOL_NAME=cal45
export CPU_START=2
export CPU_END=5



up:
	docker compose -f compose.yaml -f compose-yowking.yaml up -d

down:
	docker compose -f compose.yaml -f compose-yowking.yaml down

compose-yowking.yaml:
	node buildKingYaml.mjs

doservice:
	docker rm ${NAME} || true
	docker run \
		--cpus=1 --cpuset-cpus=${LOGICAL_PROCESSOR} \
		-v ${VOL_NAME}:/opt/yowking/calibrations \
		--env-file docker.env \
		--name ${NAME} \
		--network=yow \
		-d ace:5000/yowking
		# -it ace:5000/yowking /bin/sh
		# -l "traefik.enable=true" \
		# -l 'traefik.http.routers.yowking.rule=Host("yowking.localhost")' \
		# -l "traefik.http.routers.yowking.entrypoints=websecure" \
		# -l "traefik.http.routers.yowking.tls=true" \
		# -l "traefik.http.routers.yowking.service=yowking" \
		# -l "traefik.http.services.yowking.loadbalancer.server.port=8080" \
		# -l 'traefik.http.routers.yowking-http.rule=Host("yowking.localhost")' \
		# -l "traefik.http.routers.yowking-http.entrypoints=web" \
		# -l "traefik.http.routers.yowking-http.service=yowking" \
		# --network=traefik_default \

certs: 
	mkdir certs
	openssl req -x509 -newkey rsa:4096 -keyout certs/key.pem -out certs/cert.pem \
	-days 1825  -nodes -subj '/CN=localhost'

startk8:
	minikube start --insecure-registry="ace:5000"

push: 
	# docker push ace:5000/yowking
	docker push thinktt/yowking:latest

test: dist
	source .env; \
	export ROOT_DIR=$(shell pwd);  \
	go test ./...


gobuild:
	GOOS=windows GOARCH=386 go build -o dist/enginewrap.exe  ./cmd/enginewrap 
	GOOS=linux CGO_ENABLED=0 go build -o dist/kingapi  ./cmd/kingapi
	GOOS=linux CGO_ENABLED=0 go build -o dist/kingworker  ./cmd/kingworker


dbuild: dist
	docker rm yowking || true
	docker image rm ace:5000/yowking:latest || true
	docker build -t ace:5000/yowking .
	# docker push ace:5000/yowking  

dbuild2: 
	docker rm yowking || true
	docker image rm us-central1-docker.pkg.dev/thinktt/yowking/yowking:latest || true
	docker build -t us-central1-docker.pkg.dev/thinktt/yowking/yowking:latest .
	docker push us-central1-docker.pkg.dev/thinktt/yowking/yowking

dbuild3: 
	docker rm yowking || true
	docker image thinktt/yowking:latest || true
	docker build -t thinktt/yowking:latest .
	docker push thinktt/yowking:latest

dist: assets
	mkdir dist 
	mkdir dist/calibrations
	cp -r assets/* dist/
	mv dist/clockTimes.json dist/calibrations/clockTimes.json
	make gobuild

clean: 
	rm -rf assets
	rm -rf dist

drun: 
	docker run --rm -it --name yowking  -p 8080:8080 ace:5000/yowking
	docker run --rm -it --name yowking  -p 8080:8080 thinktt/yowking:latest


dexec: 
	# docker run --rm -it --name yowking ace:5000/yowking /bin/bash
	docker exec -it yowking /bin/bash

run: export IS_WSL=true
run: dist
	source .env; \
	cd dist && go run ../cmd/kingworker
	# cd dist && go run ../cmd/kingapi

# later we will import and build these from CM11 folder, for now borrowed locally
assets:
	mkdir assets
	# cp -r ../yeoldwiz-lnx/yowbot/dist/books assets/books
	cp ../yeoldwiz-lnx/yowbot/cals/xps/run1/clockTimes.json assets/clockTimes.json
	cp ../yeoldwiz-lnx/yowbot/dist/TheKing350noOpk.exe assets/TheKing350noOpk.exe
	cp ../yeoldwiz-lnx/yowbot/dist/personalities.json assets/personalities.json
	cp ../yeoldwiz-lnx/yowbot/dist/runbook assets/runbook
	cp -r ../yeoldwiz-lnx/yowbot/dist/books assets/books