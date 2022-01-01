# Installation

Databunker charts can be installed only with [Helm 3](https://helm.sh/docs/).

Before installing containers, you need to add Databunker Helm repository.

Run the following commands:
```
helm repo add databunker https://databunker.org/charts/
helm repo update
```

# Starting Databunker using MySQL database

## Start Databunker with auto-generated self-signed SSL certificate
```
helm install databunker databunker/databunker --set mariadb.primary.persistence.enabled=false --set certificates.customCAs\[0\].secret="databunker"
```

## Start Databunker for local testing
```
helm install databunker databunker/databunker \
  --set mariadb.primary.persistence.enabled=false \
  --set certificates.customCAs\[0\].secret="databunker"

export POD_NAME=$(kubectl get pods --namespace default -l "app.kubernetes.io/name=databunker,app.kubernetes.io/instance=databunker" -o jsonpath="{.items[0].metadata.name}")
export CONTAINER_PORT=$(kubectl get pod --namespace default $POD_NAME -o jsonpath="{.spec.containers[0].ports[0].containerPort}")
echo "Visit http://127.0.0.1:8080 to use your application"
kubectl --namespace default port-forward $POD_NAME 8080:$CONTAINER_PORT
```


# Starting Databunker DEMO

## Start in AWS
Use the following command:
```
helm install demo databunker/databunker-demo --set service.type=LoadBalancer
```

It takes a few seconds for the load balancer starts working.

You can use the following command to get URL of the load balancer:

```
kubectl get svc | grep demo
```

You can open this url in browser. By default, it will use port 3000.


## Start using NodePort
You can run the following command:
```
helm install demo databunker/databunker-demo --set service.type=NodePort
```

### Default port for NodePort

The default port is **30300**.

## Running Databunker demo on the local machine

You can open `http://localhost:30300/` in your browser.

## Removing **databunker-demo** deployment

Use the following command:
```
helm uninstall demo
```

## Chart Parameters

| Name                            | Description                                                | Value                |
| ------------------------------- | ---------------------------------------------------------- | -------------------- |
| `service.type`                  | Databunker Service Type                                    | `ClusterIP`          |
| `service.nodePort`              | Databunker API and UI port                                 | `"30300"`            |
