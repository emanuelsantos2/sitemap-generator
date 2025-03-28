// utils/s3.go
package utils

import (
    "bytes"
    "context"
    "os"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/credentials"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

func UploadToS3(fileData []byte, bucket, key, region, endpoint, contentType string) error {
    ctx := context.Background()
    
    cfg, err := config.LoadDefaultConfig(ctx,
        config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
            os.Getenv("AWS_ACCESS_KEY_ID"),
            os.Getenv("AWS_SECRET_ACCESS_KEY"),
            "",
        )),
        config.WithRegion(region),
    )
    if err != nil {
        return err
    }

    client := s3.NewFromConfig(cfg, func(o *s3.Options) {
        if endpoint != "" {
            o.BaseEndpoint = aws.String(endpoint)
        }
    })

    _, err = client.PutObject(ctx, &s3.PutObjectInput{
        Bucket: aws.String(bucket),
        Key:    aws.String(key),
        Body:   bytes.NewReader(fileData),
        ACL:    "public-read",
        ContentType: aws.String(contentType),
    })
    
    return err
}
