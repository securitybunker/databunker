#!/bin/sh

/bin/busybox find /databunker

DATABUNKER_MASTERKEY=""
if [ ! -f /databunker/data/databunker.db ]; then
  echo "-------------INIT------------"
  /bin/busybox mkdir -p /tmp
  RESULT=`/databunker/bin/databunker -init -db /databunker/data/databunker.db -conf /databunker/conf/databunker.yaml > /tmp/init.txt`
  echo $RESULT
  ROOT_TOKEN=`/bin/busybox awk '/API Root token:/ {print $4}' /tmp/init.txt`
  DATABUNKER_MASTERKEY=`/bin/busybox awk '/Master key:/ {print $3}' /tmp/init.txt`
  echo "DATABUNKER_ROOTTOKEN $ROOT_TOKEN"
  echo "DATABUNKER_MASTERKEY $DATABUNKER_MASTERKEY"
  /bin/busybox rm -rf /tmp/init.txt
fi
echo "-------------FIND------------"
/bin/busybox find /databunker
echo "-------------RUN-------------"
/databunker/bin/databunker -masterkey $DATABUNKER_MASTERKEY -db /databunker/data/databunker.db -conf /databunker/conf/databunker.yaml