package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hamstah/awstools/common"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	bucket   = kingpin.Flag("bucket", "Name of the bucket").Required().String()
	key      = kingpin.Flag("key", "Key of the uploaded file").Required().String()
	file     = kingpin.Flag("file", "File to upload").Required().File()
	acl      = kingpin.Flag("acl", "ACL of the uploaded file").Default("private").String()
	metadata = kingpin.Flag("metadata", "Metadata of the uploaded file (json)").String()
)

func main() {
	kingpin.CommandLine.Name = "s3-upload"
	kingpin.CommandLine.Help = "Upload a file to S3."
	flags := common.HandleFlags()
	defer (*file).Close()

	session, conf := common.OpenSession(flags)

	s3Client := s3.New(session, conf)
	uploader := s3manager.NewUploaderWithClient(s3Client)

	var parsedMetadata map[string]*string

	if metadata != nil && len(*metadata) != 0 {
		err := json.Unmarshal([]byte(*metadata), &parsedMetadata)
		common.FatalOnErrorW(err, "Invalid metadata")
	}

	uploadInput := &s3manager.UploadInput{
		Bucket:   bucket,
		Key:      key,
		Body:     *file,
		ACL:      acl,
		Metadata: parsedMetadata,
	}

	res, err := uploader.Upload(uploadInput)

	common.FatalOnError(err)

	fmt.Println(res.Location)
}
