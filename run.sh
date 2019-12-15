#!/bin/sh

/bin/busybox find /

DATABUNKER_MASTERKEY=
if [ ! -f /databunker/data/databunker.db ]; then
  echo "-------------INIT------------"
  /bin/busybox mkdir -p /tmp
  RESULT=`/databunker/bin/databunker -init > /tmp/init.txt`
  echo $RESULT
  ROOT_TOKEN=`/bin/busybox awk '/API Root token:/ {print $4}' /tmp/init.txt`
  MASTER_KEY=`/bin/busybox awk '/Master key:/ {print $3}' /tmp/init.txt`
  echo "DATABUNKER_ROOTTOKEN $ROOT_TOKEN"
  echo "DATABUNKER_MASTERKEY $MASTER_KEY"
  /bin/busybox rm -rf /tmp/init.txt
fi
echo "-------------RUN-------------"
#/go/bin/databunker