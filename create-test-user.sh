#!/bin/bash

set -x

DATABUNKER_APIKEY='25ae59fb-3859-b807-7420-cdb83e089b42'
curl -s http://localhost:3000/v1/user -XPOST "X-Bunker-Token: "$DATABUNKER_APIKEY -H "Content-Type: application/json" -d '{"name":"yul","email":"test@paranoidguy.com"}'
