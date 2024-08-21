#!/bin/sh

#/bin/busybox find /databunker

if [ ! -f /databunker/data/databunker.db ]; then
  OPTION="-init"
  if [ "$1" == "demo" ]; then
    OPTION="-demoinit"
  fi
  echo "-------------------------------------------------"
  echo "run.sh: init database, encryption, and access keys"
  #/bin/busybox mkdir -p /tmp
  RESULT=`/databunker/bin/databunker $OPTION -db /databunker/data/databunker.db -conf /databunker/conf/databunker.yaml > /tmp/init.txt 2>&1`
  if [ ! -f /databunker/data/databunker.db ]; then
    echo "Failed to init databunker database. Probably permission issue for /databunker/data directory."
    /bin/busybox sleep 60
    exit
  fi
  /bin/busybox cat /tmp/init.txt
  DATABUNKER_ROOTTOKEN=`/bin/busybox awk '/API Root token:/ {print $6}' /tmp/init.txt`
  DATABUNKER_MASTERKEY2=`/bin/busybox awk '/Master key:/ {print $5}' /tmp/init.txt`
  echo "DATABUNKER_ROOTTOKEN $DATABUNKER_ROOTTOKEN"
  echo "DATABUNKER_MASTERKEY $DATABUNKER_MASTERKEY2"
  /bin/busybox rm -rf /tmp/init.txt
  if [ -z "$DATABUNKER_MASTERKEY" ]; then
    echo "export DATABUNKER_MASTERKEY=$DATABUNKER_MASTERKEY2"
    export DATABUNKER_MASTERKEY=$DATABUNKER_MASTERKEY2
  fi
fi
if [ -z "$DATABUNKER_MASTERKEY" ]; then
  echo "DATABUNKER_MASTERKEY environment value is empty"
  /bin/busybox sleep 60
  exit
fi
echo "-------------------------------------------------"
echo "run.sh: shart databunker service"
/databunker/bin/databunker -start -db /databunker/data/databunker.db -conf /databunker/conf/databunker.yaml
