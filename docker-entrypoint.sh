#!/bin/sh
set -e

: ${SERVER_BIND:=":8080"}
: ${SERVER_ENCRYPTION_KEY:="changeme91234567890123456789012"}
: ${CRAWLER_STRESSFAKTOR_URI:="https://stressfaktor.squat.net/termine.php?display=30"}
: ${CRAWLER_LOCATION:="Europe/Berlin"}
: ${DB_PATH:="/var/lib/stressfaktor-api/db.sqlite3"}
: ${LOG_VERBOSE:=false}

if [ "$1" = 'stressfaktor-api' ]; then

touch /var/log/stressfaktor-api/out.log

exec stressfaktor-api \
	-server.bind=${SERVER_BIND} \
	-server.encryption.key=${SERVER_ENCRYPTION_KEY} \
	-crawler.stressfaktor.uri=${CRAWLER_STRESSFAKTOR_URI} \
	-crawler.location=${CRAWLER_LOCATION} \
	-db.path=${DB_PATH} \
	-log.verbose=${LOG_VERBOSE} | tee /var/log/stressfaktor-api/out.log 2>&1

fi

exec "$@"