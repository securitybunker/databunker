#!/bin/sh

if [ -z "$DATABUNKER_MASTERKEY" ]; then
  echo "DATABUNKER_MASTERKEY environment value is empty"
  /bin/busybox sleep 60
  exit
fi
echo "-------------RUN-------------"
/databunker/bin/databunker -start -conf /databunker/conf/databunker.yaml
