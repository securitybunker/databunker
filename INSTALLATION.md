# Installing Databunker

## From Docker container

The easiest method is to start docker container. Always use the latest version.

You can fetch and start Databunker with the following command:

```
mkdir -p /tmp/data
docker run -v /tmp/data:/databunker/data -p 3000:3000 \
  --rm --name dbunker paranoidguy/databunker
```

This command will init Databunker service, init database and start container.

This command will print **DATABUNKER_MASTERKEY** and **DATABUNKER_ROOTTOKEN**.

The database will be init in the /tmp/data parent directory.

**DATABUNKER_MASTERKEY** is used to encrypt database records.

**DATABUNKER_ROOTTOKEN** is an access token to databunker API.


## Stop service

To stop Databunker container you can run the following command:

```
docker kill dbunker
```

# Run it again

You can run it again, after it was initalized. Use the following command:

```
docker run -v /tmp/data:/databunker/data -p 3000:3000 \
  -e "DATABUNKER_MASTERKEY=KEY" --rm --name dbunker paranoidguy/databunker
```