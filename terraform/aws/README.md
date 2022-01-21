## Terraform configuration files to create AWS infrastructure for Databunker

Terraform is a powerful tool to manage infrastructure with configuration files rather than through a graphical user interface.

We use Terraform to streamline Databunker installation for AWS.

These Terraform configuration files create the following AWS infrastructure elements:

1. VPC
1. MySQL RDS
1. Elastic Kubernetes Service (EKS)
1. Security groups to allow connectivity

During deployment, Terraform generates a random MySQL password. This password is saved in newly created EKS cluster as a **Kubernetes secret** using the following resource path: ```databunker-mysql-rds/db-password```.


### ⚡ How to setup everything

Run the following command to initialize a working directory for Terraform. It will download all required components. You need to run this command only once.
```
terraform init
```

Run the following command to create AWS infrastructure:
```
terraform apply
```

You can use the following command to display full MySQL database domain name. You will need its value in the next part.
```
terraform output -raw rds_hostname
```

### ☕ Next steps
1. Set **KUBECONFIG** environment variable to point to a newly generated configuration file for Kubernetes
```
export KUBECONFIG=`pwd`/`ls -1 kubeconfig_*`
```
After this command, you can execute ```kubectl get nodes``` to list all nodes in newly created EKS cluster.

2. Create an SSL certificate for Databunker service and save it as Kubernetes secret
```
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout tls.key -out tls.crt -subj "/CN=localhost"
kubectl create secret tls databunkertls --key="tls.key" --cert="tls.crt"
```
3. Add Databunker charts repository using ```helm``` command and run ```helm update```
```
helm repo add databunker https://databunker.org/charts/
helm repo update
```
4. Deploy Databunker service using the ```helm``` command
```
MYSQL_RDS_HOST=$(terraform output -raw rds_hostname)
helm install databunker databunker/databunker --set mariadb.enabled=false \
  --set externalDatabase.host=$MYSQL_RDS_HOST \
  --set externalDatabase.existingSecret=databunker-mysql-rds \
  --set certificates.customCertificate.certificateSecret=databunkertls
```

🚩 The **MYSQL_RDS_HOST** above is a full MySQL domain name.


**All commands together**

```
export KUBECONFIG=`pwd`/`ls -1 kubeconfig_*`
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout tls.key -out tls.crt -subj "/CN=localhost"
kubectl create secret tls databunkertls --key="tls.key" --cert="tls.crt"
helm repo add databunker https://databunker.org/charts/
helm repo update
MYSQL_RDS_HOST=$(terraform output -raw rds_hostname)
helm install databunker databunker/databunker --set mariadb.enabled=false \
  --set externalDatabase.host=$MYSQL_RDS_HOST \
  --set externalDatabase.existingSecret=databunker-mysql-rds \
  --set certificates.customCertificate.certificateSecret=databunkertls
```

### ⚙️ Upgrade cluster to use the latest Databunker version

During the first time deployment of the Databunker helm charts, the setup process creates a special '''Kubernetes secret''' callled **databunker**.
This secret contains the **DATABUNKER_MASTERKEY** used for record encryption and the **DATABUNKER_ROOTTOKEN** used for service access.
This Kubernetes secret is never deleted. So, you can easily remove the helm char and/or update to the latest version. Databunker process will continue working with old encrypted records.
```
helm repo update
MYSQL_RDS_HOST=$(terraform output -raw rds_hostname)
helm upgrade databunker --set mariadb.enabled=false \
  --set externalDatabase.host=$MYSQL_RDS_HOST \
  --set externalDatabase.existingSecret=databunker-mysql-rds \
  --set certificates.customCertificate.certificateSecret=databunkertls
```

### 🔍 View generated database password using terraform
```
terraform output -raw rds_password
```

### 🔍 View generated database password using kubernetes secret
```
kubectl get secret databunker-mysql-rds -o json
```

### 🛠️ Troubleshooting
Different commands can be used to troubleshoot deployment:

```
terraform destroy -target aws_eks_cluster.yuli-cluster
terraform destroy -target module.eks.aws_eks_cluster.this\[0\]
terraform destroy
helm uninstall databunker
kubectl get secret databunkertls -o json
```
