#!/bin/sh

#/bin/busybox find /databunker

if [ ! -f /databunker/data/databunker.db ]; then
  OPTION="-init"
  if [ "$1" == "demo" ]; then
    OPTION="-demoinit"
  fi
  echo "-------------INIT------------"
  #/bin/busybox mkdir -p /tmp
  RESULT=`/databunker/bin/databunker $OPTION -db /databunker/data/databunker.db -conf /databunker/conf/databunker.yaml > /tmp/init.txt`
  if [ ! -f /databunker/data/databunker.db ]; then
    echo "Failed to init databunker database. Probably permission issue for /databunker/data directory."
    /bin/busybox sleep 60
    exit
  fi
  DATABUNKER_ROOTTOKEN=`/bin/busybox awk '/API Root token:/ {print $4}' /tmp/init.txt`
  DATABUNKER_MASTERKEY2=`/bin/busybox awk '/Master key:/ {print $3}' /tmp/init.txt`
  echo "DATABUNKER_ROOTTOKEN $DATABUNKER_ROOTTOKEN"
  echo "DATABUNKER_MASTERKEY $DATABUNKER_MASTERKEY2"
  /bin/busybox rm -rf /tmp/init.txt
  if [ -z "$DATABUNKER_MASTERKEY" ]; then
    export DATABUNKER_MASTERKEY=$DATABUNKER_MASTERKEY2
  fi
fi
if [ -z "$DATABUNKER_MASTERKEY" ]; then
  echo "DATABUNKER_MASTERKEY environment value is empty"
  /bin/busybox sleep 60
  exit
fi
#echo "-------------ENV-------------"
#/bin/busybox env
#echo "-------------FIND------------"
#/bin/busybox find /databunker
echo "-------------RUN-------------"
/databunker/bin/databunker -start -db /databunker/data/databunker.db -conf /databunker/conf/databunker.yaml
