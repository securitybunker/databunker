# Building Databunker

Use the following command:

```
./build.sh
```

This command will generate the ```databunker``` executable, bundled with the web UI interface.

## Building Databunker Container

To generate the ```securitybunker/databunker``` container, use the following command:

```
VERSION=$(cat ./version.txt)
docker build -t securitybunker/databunker:$VERSION .
```

## Pushing Container

For project admins only:

```
docker login
VERSION=$(cat ./version.txt)
docker push securitybunker/databunker:$VERSION
# Optionally, push container with the latest tag
docker tag securitybunker/databunker:$VERSION securitybunker/databunker:latest
docker push securitybunker/databunker:latest
```

## Other Useful Commands for Working with Containers:

```
docker container stats databunker
docker run --rm -ti alpine
/bin/busybox wget localhost:3000
```

## Check what packages require cgo
```
go list -f "{{if .CgoFiles}}{{.ImportPath}}{{end}}" $(go list -f "{{.ImportPath}}{{range .Deps}} {{. }}{{end}}")
```
