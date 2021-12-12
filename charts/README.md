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
