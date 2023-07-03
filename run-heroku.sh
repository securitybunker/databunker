#!/bin/sh

if [ -z "$DATABUNKER_MASTERKEY" ]; then
	echo "DATABUNKER_MASTERKEY environment value is empty"
	/bin/busybox sleep 60
	exit
fi

echo "-------------ENV-------------"
/bin/busybox env

echo "-------------FIND------------"
/bin/busybox find /databunker

echo "-------------RUN-------------"
/databunker/bin/databunker -start -db $PGSQL_DB -conf /databunker/conf/databunker-heroku.yaml
