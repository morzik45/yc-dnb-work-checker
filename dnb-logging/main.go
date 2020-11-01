package dnb_logging

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"math/rand"
	"os"
	"time"
)

func GetRandInt() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Int()
}

func SaveError(myErr string) {
	bucket := os.Getenv("DB_NAME")
	filename := fmt.Sprintf("%s%d%s", time.Now().Format(time.RFC3339), GetRandInt(), ".txt")

	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String("ru-central1"),
		Endpoint: aws.String("https://storage.yandexcloud.net"),
	},
	)

	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fmt.Sprintf("logs/work-checker/%s", filename)),
		Body:   bytes.NewReader([]byte(myErr)),
	})
	if err != nil {
		err := fmt.Errorf("unable to upload %q to %q, %v", filename, bucket, err)
		fmt.Println(err.Error())
	}
	fmt.Printf("Successfully uploaded %q to %q\n", filename, bucket)
}
