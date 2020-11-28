# Building Databunker

## Build for release

```
./release.sh
```

It will generate **databunker** executable. HTML files are built inside executable.

## Debug version

```
./build.sh
```

It will generate **databunker** executable that can be run on the same box.
Web UI files will be fetched from ui/ directory.

## Build container

It will generate "securitybunker/databunker" container and save it locally.

```
docker build -t securitybunker/databunker:latest .
```

## Push container

**Only for project admin:**

```
docker login
docker push securitybunker/databunker:latest
```


## Other usefull commands for working with containers:

```
docker rm dbunker
docker kill dbunker
docker container stats dbunker
docker run --rm -ti alpine
/bin/busybox wget localhost:3000
```
