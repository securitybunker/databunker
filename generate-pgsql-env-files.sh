#!/bin/sh

echo 'creating ./data directory'
mkdir -p data
chmod 777 data
mkdir -p .env

echo 'generating .env/postgresql-postgres.env'
POSTGRES_PASSWORD=`< /dev/urandom LC_CTYPE=C tr -dc '_\*^A-Z-a-z-0-9' | head -c${1:-32};`
echo 'POSTGRES_PASSWORD='$POSTGRES_PASSWORD > .env/postgresql-postgres.env

echo 'generating .env/postgresql.env'
PGSQL_USER_PASSWORD=`< /dev/urandom LC_CTYPE=C tr -dc '_\*^A-Z-a-z-0-9' | head -c${1:-32};`
echo 'PGSQL_DATABASE=databunkerdb' > .env/postgresql.env
echo 'PGSQL_USER=bunkeruser' >> .env/postgresql.env
echo 'PGSQL_PASSWORD='$PGSQL_USER_PASSWORD >> .env/postgresql.env

echo 'generating .env/databunker.env'
KEY=`< /dev/urandom LC_CTYPE=C tr -dc 'a-f0-9' | head -c${1:-48};`
echo 'DATABUNKER_MASTERKEY='$KEY > .env/databunker.env
echo 'PGSQL_USER_NAME=bunkeruser' >> .env/databunker.env
echo 'PGSQL_USER_PASS='$PGSQL_USER_PASSWORD >> .env/databunker.env
echo 'PGSQL_HOST=postgresql' >> .env/databunker.env
echo 'PGSQL_PORT=5432' >> .env/databunker.env

echo 'generating ssl sertificate for postgres server'
rm -rf .env/pg-*
openssl req -new -text -passout pass:abcd -subj /CN=localhost -out .env/pg-server.req -keyout .env/pg-privkey.pem
openssl rsa -in .env/pg-privkey.pem -passin pass:abcd -out .env/pg-server.key
openssl req -x509 -in .env/pg-server.req -text -key .env/pg-server.key -out .env/pg-server.crt

chmod 400 .env/pg-*
os=$(uname)
if [ "$os" != "Darwin" ]; then
  echo "sudo chown 999:0 .env/pg-*"
  sudo chown 999:0 .env/pg-*
fi

echo 'generating .env/databunker-root.env'
ROOTTOKEN=`uuid 2> /dev/null`
if [ $? -ne 0 ]; then
  ROOTTOKEN=`uuidgen`
fi
if [ $? -ne 0 ]; then
  echo "Failed to generate DATABUNKER_ROOTTOKEN"
else
  echo 'DATABUNKER_ROOTTOKEN='$ROOTTOKEN > .env/databunker-root.env
fi
