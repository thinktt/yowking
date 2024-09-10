# JS_FILES := $(wildcard *.txt)
# BINS := $(JS_FILES:%.txt=%)

SHELL := /bin/bash

certs:
	mkdir certs
	openssl req -new -newkey rsa:2048 -days 365 -nodes -x509 \
		-subj "/C=US/O=ACME/CN=yeoldwiz.localhost" \
		-keyout certs/key.pem -out certs/cert.pem

.PHONY: run docal getclocks dbuild drun din dcal clean books2bin \
	rmbooks  reset eep testengine cpbadbooks push gobuild dexec \
	getassets test startk8 dbuild2 dbuild3 dodcal doservice  \
	compose-yowking.yaml up down 

dist:
	mkdir dist 
	mkdir dist/calibrations
	cp -r ../yeoldwiz/yowdeps/dist/* dist/
	cp ../yeoldwiz/yowbot/cals/xps/run1/clockTimes.json dist/calibrations/clockTimes.json
	make gobuild

run: export IS_WSL=true
run: dist
	source .env; \
	# cd dist && go run ../cmd/kingworker
	cd dist && go run ../cmd/kingapi

assets:
	mkdir assets
	# cp -r ../yeoldwiz-lnx/yowbot/dist/books assets/books
	cp ../yeoldwiz-lnx/yowbot/cals/xps/run1/clockTimes.json assets/clockTimes.json
	cp ../yeoldwiz-lnx/yowbot/dist/TheKing350noOpk.exe assets/TheKing350noOpk.exe
	cp ../yeoldwiz-lnx/yowbot/dist/personalities.json assets/personalities.json
	cp ../yeoldwiz-lnx/yowbot/dist/runbook assets/runbook
	cp -r ../yeoldwiz-lnx/yowbot/dist/books assets/books

push: 
	# docker push zen:5000/yowking
	docker push thinktt/yowking:latest

test: dist
	source .env; \
	export ROOT_DIR=$(shell pwd);  \
	go test ./...


gobuild:
	GOOS=windows GOARCH=386 go build -o dist/enginewrap.exe  ./cmd/enginewrap 
	GOOS=linux CGO_ENABLED=0 go build -o dist/kingworker  ./cmd/kingworker

buildapi:
	GOOS=linux CGO_ENABLED=0 go build -o dist/kingapi  ./cmd/kingapi

runapi:
	cd dist && go run ../cmd/kingapi

dbuild: dist
	docker rm yowking || true
	docker image rm zen:5000/yowking:latest || true
	docker build -t zen:5000/yowking .
	# docker push zen:5000/yowking  

dbuildapi: dist/kingapi
	docker image rm zen:5000/kingapi:latest || true
	docker build -t zen:5000/kingapi:latest -f cmd/kingapi/Dockerfile .

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

clean: 
	rm -rf assets
	rm -rf dist

drun: 
	# docker run --rm -it --name yowking  -p 8080:8080 zen:5000/yowking
	# docker run --rm -it --name yowking  -p 8080:8080 thinktt/yowking:latest
	# docker run --rm -it --env-file ./env --name yowking zen:5000/yowking

drunapi:
	docker rm kingapi || true
	docker run --rm -it --name kingapi -p 8080:8080 \
		-e JWT_KEY=XG6paWLxhsCiVcOMDF3YXWYZLeb8oyrYjyPKqbld \
	 	-e NATS_TOKEN=tYlZBeYnAEfAJofInCML53ot3ibrFJWWCpdaIYRx \
		-e NATS_URL=nats:4222 \
	 	zen:5000/kingapi /bin/ash

dexec: 
	# docker run --rm -it --name yowking zen:5000/yowking /bin/bash
	docker exec -it yowking /bin/bash



export CPU1=10
export CPU2=11
export LOGICAL_PROCESSOR=${CPU1},${CPU2}
export NAME=yowkingdp${CPU1}${CPU2}
export VOL_NAME=cal45
export CPU_START=2
export CPU_END=11

doservice:
	docker rm ${NAME} || true
	docker run \
		--cpus=1 --cpuset-cpus=${LOGICAL_PROCESSOR} \
		-v ${VOL_NAME}:/opt/yowking/calibrations \
		--env-file docker.env \
		--name ${NAME} \
		--network=yow \
		-it zen:5000/yowking /bin/sh
		#-d zen:5000/yowking
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

jot: 
	go install ./cmd/jot/