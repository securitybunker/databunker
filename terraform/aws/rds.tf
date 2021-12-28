
resource "aws_db_subnet_group" "mydb" {
  name       = "mydb"
  subnet_ids = module.vpc.private_subnets

  tags = {
    Name = "mydb_rds_subnet"
    "${local.tag_name}" = "${local.tag_value}"
  }
}

resource "aws_security_group" "rds" {
  name   = "mydb_rds"
  vpc_id = module.vpc.vpc_id

  ingress {
    from_port   = 3306
    to_port     = 3306
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    #cidr_blocks = module.vpc.private_subnets
  }

  egress {
    from_port   = 3306
    to_port     = 3306
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "mydb_rds"
    "${local.tag_name}" = "${local.tag_value}"
  }
}

resource "aws_db_parameter_group" "mydb" {
  name   = "mydb"
  family = "mysql8.0"

  #parameter {
  #  name  = "log_connections"
  #  value = "1"
  #}
}

resource "aws_db_instance" "mydb" {
  # https://github.com/tmknom/terraform-aws-rds-mysql/blob/master/main.tf
  # The name of the database. If this parameter is not specified, no database is created in the DB instance.
  name                   = "bunkerdb"
  identifier             = "mydb"
  tags                   = { "Name" = "mydb"}
  instance_class         = "db.t3.medium"
  allocated_storage      = 5
  engine                 = "mysql"
  engine_version         = "8.0.25"
  username               = "bunkeruser"
  #password               = var.db_password
  password               = random_password.db_password.result
  db_subnet_group_name   = aws_db_subnet_group.mydb.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  parameter_group_name   = aws_db_parameter_group.mydb.name
  publicly_accessible    = false
  skip_final_snapshot    = true
  # The following list briefly describes the three storage types:
  #
  # - General Purpose SSD – Also known as gp2, volumes offer cost-effective storage that is ideal for a broad range of workloads.
  # - Provisioned IOPS – Also known as io1, that require low I/O latency and consistent I/O throughput.
  # - Magnetic – RDS also supports magnetic storage for backward compatibility.
  #
  # https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_Storage.html
  #storage_type = var.storage_type
  # The amount of Provisioned IOPS (input/output operations per second) to be initially allocated for the DB instance.
  # Must be a multiple between 1 and 50 of the storage amount, and range of Provisioned IOPS is 1000–32,000
  # https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_Storage.html#USER_PIOPS
  #iops = var.iops
}

resource "kubernetes_secret" "databunker-mysql-rds" {
  metadata {
    name = "databunker-mysql-rds"
  }

  data = {
    #host = aws_db_instance.mydb.address
    #port = aws_db_instance.mydb.port
    #dbname = aws_db_instance.mydb.name
    #username = aws_db_instance.mydb.username
    "mariadb-password" = aws_db_instance.mydb.password
  }
  type = "Opaque"
}

output "rds_hostname" {
  description = "RDS instance hostname"
  value       = aws_db_instance.mydb.address
}

output "rds_port" {
  description = "RDS instance port"
  value       = aws_db_instance.mydb.port
}

output "rds_dbname" {
  description = "RDS database name"
  value = aws_db_instance.mydb.name
}

output "rds_username" {
  description = "RDS instance root username"
  value       = aws_db_instance.mydb.username
}

output "rds_password" {
  description = "RDS instance database password"
  value       = aws_db_instance.mydb.password
  sensitive = true
}
