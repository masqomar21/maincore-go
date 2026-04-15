package services

import (
	"context"
	"log"
	"mime/multipart"

	"maincore_go/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var S3Client *s3.Client

func InitS3() {
	bucket := config.AppConfig.S3Bucket
	if bucket == "" {
		log.Println("S3 Bucket not configured, skipping S3 initialization")
		return
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(config.AppConfig.S3Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			config.AppConfig.S3AccessKeyID,
			config.AppConfig.S3SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	S3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		if config.AppConfig.S3Endpoint != "" {
			o.BaseEndpoint = aws.String(config.AppConfig.S3Endpoint)
		}
		o.UsePathStyle = config.AppConfig.S3ForcePathStyle
	})

	log.Println("S3 Client initialized")
}

func UploadFileToS3(ctx context.Context, file multipart.File, header *multipart.FileHeader, path string) (string, error) {
	key := path + "/" + header.Filename

	_, err := S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(config.AppConfig.S3Bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(header.Header.Get("Content-Type")),
	})

	if err != nil {
		return "", err
	}

	return key, nil
}

func DeleteFileFromS3(ctx context.Context, key string) error {
	_, err := S3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(config.AppConfig.S3Bucket),
		Key:    aws.String(key),
	})
	return err
}
