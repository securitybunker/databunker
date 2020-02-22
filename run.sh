#!/bin/sh

/bin/busybox find /databunker

if [ ! -f /databunker/data/databunker.db ]; then
  echo "-------------INIT------------"
  #/bin/busybox mkdir -p /tmp
  RESULT=`/databunker/bin/databunker -init -db /databunker/data/databunker.db -conf /databunker/conf/databunker.yaml > /tmp/init.txt`
  echo $RESULT
  DATABUNKER_ROOTTOKEN=`/bin/busybox awk '/API Root token:/ {print $4}' /tmp/init.txt`
  DATABUNKER_MASTERKEY=`/bin/busybox awk '/Master key:/ {print $3}' /tmp/init.txt`
  echo "DATABUNKER_ROOTTOKEN $DATABUNKER_ROOTTOKEN"
  echo "DATABUNKER_MASTERKEY $DATABUNKER_MASTERKEY"
  /bin/busybox rm -rf /tmp/init.txt
fi
#echo "-------------ENV-------------"
#/bin/busybox env
echo "-------------FIND------------"
/bin/busybox find /databunker
echo "-------------RUN-------------"
/databunker/bin/databunker -start -db /databunker/data/databunker.db -conf /databunker/conf/databunker.yaml
