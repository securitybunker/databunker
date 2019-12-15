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
Web UI will be fetched from ui directory.

## Build container

```
docker build -t paranoidguy/databunker:latest .
```

It will generate "paranoidguy/databunker" container.

Other usefull commands for working with docker:

```
docker rm dbunker
docker kill dbunker
docker container stats dbunker
docker run --rm -ti alpine
/bin/busybox wget localhost:3000
```