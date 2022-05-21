#!/bin/bash -

ENV=$2
PROJECT=nec
CURRENT_PATH=$(cd `dirname $0` && pwd)
BIN_PATH=${CURRENT_PATH}/nec
LOG_DIR=/data/logs/app/${PROJECT}

mkdir -p ${LOG_DIR}

function create_supervisor_file() {

cat > /data/etc/supervisord/${PROJECT}.ini <<EOF
[program:${PROJECT}]
directory=${CURRENT_PATH}
command=${BIN_PATH} start
numprocs=1
autostart=false
autorestart=true
startsecs=1
startretries=3
redirect_stderr=true
user=work
stdout_logfile=${LOG_DIR}/stdout.log
stdout_logfile_maxbytes=500MB
stdout_logfile_backups=10
EOF
}

function stop() {
	sudo supervisorctl stop ${PROJECT}

	if [[ $? -ne 0 ]]; then
		echo "stop ${PROJECT} fail"
		exit 1
	fi
}

function start() {
	sudo supervisorctl status ${PROJECT} | grep RUNNING > /dev/null

	if [[ $? -eq 0 ]]; then
		echo "${PROJECT} is already running"
		exit 1
	fi

	create_supervisor_file && (sudo supervisorctl update) && (sudo supervisorctl start ${PROJECT})
	if [[ $? -ne 0 ]]; then
		echo "start ${PROJECT} fail"
		exit 1
	fi
}

if [[ $# -lt 1 ]]; then
	echo "usage: $0 [start|stop|restart]"
	exit 1
else
	if [[ "$1" == 'stop' ]] ; then
		stop
		echo "${PROJECT} stop"
	elif [[ "$1" == 'start' ]] ; then
		start
		echo "${PROJECT} start"
	elif [[ "$1" == 'restart' ]] ; then
		stop
		start
		echo "${PROJECT} restart"
	else
		echo "usage: $0 [start|stop|restart]"
		exit 1
	fi
fi

exit 0
