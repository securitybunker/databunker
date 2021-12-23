# Starting Databunker with MySQL backend

## Start Databunker with auto-generated self-signed SSL certificate
```
helm install --debug databunker ./databunker --set mariadb.primary.persistence.enabled=false --set certificates.customCAs\[0\].secret="databunker"
```

## Start Databunker for local testing
```
helm install --debug databunker ./databunker --set mariadb.primary.persistence.enabled=false --set certificates.customCAs\[0\].secret="databunker"

export POD_NAME=$(kubectl get pods --namespace default -l "app.kubernetes.io/name=databunker,app.kubernetes.io/instance=databunker" -o jsonpath="{.items[0].metadata.name}")
export CONTAINER_PORT=$(kubectl get pod --namespace default $POD_NAME -o jsonpath="{.spec.containers[0].ports[0].containerPort}")
echo "Visit http://127.0.0.1:8080 to use your application"
kubectl --namespace default port-forward $POD_NAME 8080:$CONTAINER_PORT
```


# Starting Databunker DEMO pod using Kubernetes

You can run the following command:
```
helm install databunker-demo ./databunker-demo --set service.type=NodePort
```

The `./databunker-demo` directory must contain all help chart files for the **databunker-demo** project.

## Default port for NodePort

The default port is **30300**.

## Running Kubernetes on the local machine

You can open `http://localhost:30300/` in your browser.

## Using AWS KMS

You will need to get to know what is the public IP address of the Kubernetes node used to run the service.

You need to open port **30300** in the security groups configuration for the same EC2 instance.

## Removing **databunker-demo** pod

Use the following command:
```
helm uninstall databunker-demo
```

## Chart Parameters

| Name                            | Description                                                | Value                |
| ------------------------------- | ---------------------------------------------------------- | -------------------- |
| `service.type`                  | Databunker Service Type                                    | `ClusterIP`          |
| `service.nodePort`              | Databunker API and UI port                                 | `"30300"`            |
