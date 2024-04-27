# Building Databunker

Use the folllowing command

```
./build.sh
```

It will generate **databunker** executable. HTML files are built inside executable.

## Build container

It will generate "securitybunker/databunker" container and save it locally.

```
VERSION=$(cat ./version.txt)
docker build -t securitybunker/databunker:$VERSION --build-arg VERSION=$VERSION .
```

## Push container

**Only for project admin:**

```
docker login
VERSION=$(cat ./version.txt)
docker push securitybunker/databunker:$VERSION
```


## Other useful commands for working with containers:

```
docker rm dbunker
docker kill dbunker
docker container stats dbunker
docker run --rm -ti alpine
/bin/busybox wget localhost:3000
```
