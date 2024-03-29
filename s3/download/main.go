package main

import (
	"os"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hamstah/awstools/common"
)

var (
	bucket   = kingpin.Flag("bucket", "Name of the bucket").Required().String()
	key      = kingpin.Flag("key", "Key to download").Required().String()
	filename = kingpin.Flag("filename", "Output filename").Required().String()
)

func main() {
	kingpin.CommandLine.Name = "s3-download"
	kingpin.CommandLine.Help = "Download a file from S3."
	flags := common.HandleFlags()

	session, conf := common.OpenSession(flags)

	s3Client := s3.New(session, conf)
	downloader := s3manager.NewDownloaderWithClient(s3Client)

	f, err := os.Create(*filename)
	common.FatalOnError(err)
	defer f.Close()

	_, err = downloader.Download(f, &s3.GetObjectInput{
		Bucket: bucket,
		Key:    key,
	})

	if err != nil {
		f.Close()
		os.Remove(*filename)
		common.Fatalln(err.Error())
	}
}
