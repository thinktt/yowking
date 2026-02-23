SHELL := /bin/bash

# export CPU1=10
# export CPU2=11
# export LOGICAL_PROCESSOR=${CPU1},${CPU2}
# export NAME=yowking${CPU1}${CPU2}
# export VOL_NAME=cal45
export CPU_START=2
export CPU_END=11

.PHONY: calclean up down kingup kingdown botdown resources cpuf cput dexcal \
	cpusee cpurestart cpuseegov cpuseemax offturbo botup botdown

dexcal:
	docker exec -it yowcal /bin/ash

botdown:
	docker compose down yowbot

calclean:
	rm -rf dist/calibrations/*

resources:
	docker network create proxynet || true
	docker network create mongonet || true
	docker network create yow || true
	docker volume create mongodb || true
	docker volume create cal45 || true

up:
	docker compose -f compose.yaml -f compose-yowking.yaml up -d

down:
	docker compose -f compose.yaml -f compose-yowking.yaml down

kingoneup:
	docker compose -f compose-yowking.yaml up yowking45 -d

kingup:
	docker compose -f compose-yowking.yaml up -d

kingdown:
	docker compose -f compose-yowking.yaml down

botup:
	docker compose up yowbot -d

# this command requires you add CPU_START=<num> CPU_END=<num> at th end
compose-yowking.yaml:
	node buildKingYaml.mjs
	# docker exec -e CPU_START=$(CPU_START) -e CPU_END=$(CPU_END) yowcal node buildKingYaml.mjs
	# docker cp yowcal:/opt/yeoldwiz/compose-yowking.yaml .
	# docker cp yowcal:/opt/yeoldwiz/compose-yowking.yaml .

cpuf:
	watch -n1 "grep \"^[c]pu MHz\" /proc/cpuinfo"

cput:
	watch -n .1 sensors

cpusee:
	# cat /etc/default/cpufrequtils
	cat /sys/devices/system/cpu/cpu0/cpufreq/scaling_available_frequencies
	cat /sys/devices/system/cpu/cpu2/cpufreq/scaling_available_governors
	cat /sys/devices/system/cpu/cpu2/cpufreq/scaling_max_freq
	cat /sys/devices/system/cpu/cpu3/cpufreq/scaling_max_freq
	cat /sys/devices/system/cpu/cpu3/cpufreq/scaling_governor
	cat /sys/devices/system/cpu/cpufreq/boost

cpuseemax:
	cat /sys/devices/system/cpu/cpu*/cpufreq/scaling_max_freq

cpuseegov:
	cat /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor

cpuseetrubo:
	cat /sys/devices/system/cpu/cpufreq/boost

cpurestart:
	systemctl restart cpufrequtils	

offturbo: 
	echo "0" | tee /sys/devices/system/cpu/cpufreq/boost
	cp rc.local /etc/rc.local


natsreqinfo:
	nats --server localhost:4222 str info move-req-stream --json

natsresinfo: 
	nats --server localhost:4222 str info move-res-stream --json

natsls: 
	nats --server localhost:4222 str ls

natsconreq:
	nats --server localhost:4222 con ls move-req-stream

natsconres:
	nats --server localhost:4222 con ls move-res-stream

natsconking:
	nats --server localhost:4222 con info move-req-stream kingworkers