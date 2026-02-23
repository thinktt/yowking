SHELL := /bin/bash

.PHONY: dist run push test gobuild dbuild dbuild2 dbuild3 clean drun dpush dexec doservice

dist:
	mkdir dist 
	mkdir dist/calibrations
	cp -r ../yowdeps/dist/* dist/
	cp ../yeoldwiz/yowbot/cals/xps/run1/clockTimes.json dist/calibrations/clockTimes.json
	make gobuild

run: export IS_WSL=true
run: dist
	source .env; \
	cd dist && go run ../cmd/kingworker

push: 
	docker push thinktt/yowking:latest

test: dist
	source .env; \
	export ROOT_DIR=$(shell pwd);  \
	go test ./...


gobuild:
	GOOS=windows GOARCH=386 go build -o dist/enginewrap.exe  ./cmd/enginewrap 
	GOOS=linux CGO_ENABLED=0 go build -o dist/kingworker  ./cmd/kingworker

dbuild: dist
	docker rm yowking || true
	docker image rm zen:5000/yowking:latest || true
	docker build -t zen:5000/yowking .

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
	rm -rf dist

drun: 
	docker run --rm -it --name yowking --env-file ./docker.env --network=yow thinktt/yowking:latest

dpush: 
	docker push zen:5000/yowking

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
