package config

type Minio struct {
	Endpoint          string `yaml:"endpoint" env:"MINIO_ENDPOINT"`
	AccessKeyID       string `yaml:"access_key_id" env:"MINIO_ACCESS_KEY_ID"`
	SecretAccessKey   string `yaml:"secret_access_key" env:"MINIO_SECRET_ACCESS_KEY"`
	Secure            bool   `yaml:"secure" env:"MINIO_SECURE"`
	Token             string `yaml:"token" env:"MINIO_TOKEN"`
	UploadFileThreads uint   `yaml:"upload_file_threads" env:"MINIO_UPLOAD_FILE_THREADS"`
}
