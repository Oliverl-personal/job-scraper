package storage

type ProviderType string

const (
	AwsS3Bucket ProviderType = "aws_s3_bucket"
	MariaDb     ProviderType = "mariadb"
	Mysql       ProviderType = "mysql"
	Postgres    ProviderType = "postgres"
	InMemory    ProviderType = "in_memory"
)
