package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func main() {

	bucket := flag.String("bucket", "", "Bucket name")
	key := flag.String("key", "", "Key to download")
	filename := flag.String("filename", "", "Filename")
	flag.Parse()

	if len(*bucket) < 1 {
		fmt.Println("Missing bucket name")
		os.Exit(1)
	}

	if len(*key) < 1 {
		fmt.Println("Missing key")
		os.Exit(1)
	}

	if len(*filename) < 1 {
		fmt.Println("Missing filename")
		os.Exit(1)
	}

	sess, err := session.NewSession()
	s3Svc := s3.New(sess)
	downloader := s3manager.NewDownloaderWithClient(s3Svc)


	f, err := os.Create(*filename)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}

	_, err = downloader.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(*bucket),
		Key:    aws.String(*key),
	})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
}
