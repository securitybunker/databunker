

```
terraform apply
terraform destroy -target aws_eks_cluster.yuli-cluster
terraform destroy -target module.eks.aws_eks_cluster.this\[0\]
terraform output rds_password

export KUBECONFIG=/Users/yuli/Desktop/code/databunker/terraform/kubeconfig_yuli-cluster
export KUBE_CONFIG_PATH=/Users/yuli/Desktop/code/databunker/terraform/kubeconfig_yuli-cluster

```

