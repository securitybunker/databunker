## Terraform script to prepare environment for Databunker

1. Create VPC
2. Create MySQL RDS
3. Create EKS

### How to set up everything

```
terraform init
terraform apply
```

Make sure to save the database hostname displayed as **rds_hostname** variable.

Same RDS hostname is printed using the following command:

```
terraform output rds_hostname
```

### Next steps
1. Set KUBECONFIG to point to new generated kubernetes config file
2. Create SSL certificate for Databunker and save it as Kubernetes secret
3. Start Databunker process

```
export KUBECONFIG=`pwd`/`ls -1 kubeconfig_*`
cd ../../charts
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout tls.key -out tls.crt -subj "/CN=localhost"
kubectl create secret tls databunkertls --key="tls.key" --cert="tls.crt"
helm install myprj ./databunker --set mariadb.enabled=false \
  --set externalDatabase.host=MYSQL-RDS-HOST \
  --set externalDatabase.existingSecret=databunker-mysql-rds \
  --set certificates.customCertificate.certificateSecret=databunkertls
```

The **MYSQL-RDS-HOST** is the same as ```terraform output rds_hostname```.

### View generated database password

```
terraform output rds_password
```

### Troubleshooting
```
terraform destroy -target aws_eks_cluster.yuli-cluster
terraform destroy -target module.eks.aws_eks_cluster.this\[0\]
terraform destroy
helm uninstall myprj
kubectl get secret databunkertls -o json
kubectl get secret databunker-mysql-rds -o json
```
