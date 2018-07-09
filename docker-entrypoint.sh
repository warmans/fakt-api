#!/bin/sh
set -e

: ${SERVER_BIND:=":8080"}
: ${SERVER_ENCRYPTION_KEY:="changeme91234567890123456789012"}
: ${CRAWLER_STRESSFAKTOR_URI:="https://stressfaktor.squat.net/termine.php?days=all"}
: ${CRAWLER_LOCATION:="Europe/Berlin"}
: ${DB_PATH:="/opt/fakt-api/db/db.sqlite3"}
: ${LOG_VERBOSE:=false}
: ${MIGRATIONS_PATH:="/opt/fakt-api/migrations"}
: ${MIGRATIONS_DISABLED:="false"}

if [ "$1" = 'fakt-api' ]; then

	touch /var/log/fakt-api/out.log;
	cd /opt/fakt-api/;
	exec ./fakt-api \
		-server.bind=${SERVER_BIND} \
		-server.encryption.key=${SERVER_ENCRYPTION_KEY} \
		-crawler.stressfaktor.uri=${CRAWLER_STRESSFAKTOR_URI} \
		-crawler.location=${CRAWLER_LOCATION} \
		-db.path=${DB_PATH} \
		-log.verbose=${LOG_VERBOSE} \
		-migrations.path=${MIGRATIONS_PATH} \
		-migrations.disabled=${MIGRATIONS_DISABLED} | tee /var/log/fakt-api/out.log 2>&1
fi

exec "$@"