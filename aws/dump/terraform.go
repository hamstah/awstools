package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hamstah/awstools/common"
	"github.com/hashicorp/terraform/states"
	"github.com/hashicorp/terraform/states/statefile"
)

type Substitution struct {
	Old string `json:"old"`
	New string `json:"new"`
}

type Options struct {
	PathSubstitutions []Substitution `json:"path_substitutions"`
	Overwrite         bool           `json:"overwrite"`
}

type S3Backend struct {
	Bucket      string   `json:"bucket"`
	Keys        []string `json:"keys"`
	Region      string   `json:"region"`
	RoleARN     string   `json:"role_arn"`
	ExternalID  string   `json:"external_id"`
	SessionName string   `json:"session_name"`
}

func (s3Backend *S3Backend) Download(destination string, options *Options) (map[string]string, error) {
	sess, conf := common.OpenSession(&common.SessionFlags{
		RoleArn:         &s3Backend.RoleARN,
		RoleExternalID:  &s3Backend.ExternalID,
		Region:          &s3Backend.Region,
		RoleSessionName: &s3Backend.SessionName,

		MFASerialNumber: aws.String(""),
		MFATokenCode:    aws.String(""),
	})

	filenames := make(map[string]string, len(s3Backend.Keys))
	objects := make([]s3manager.BatchDownloadObject, 0, len(s3Backend.Keys))
	for _, key := range s3Backend.Keys {

		transformed := key
		for _, substitution := range options.PathSubstitutions {
			transformed = strings.Replace(transformed, substitution.Old, substitution.New, -1)
		}

		dir, _ := filepath.Split(transformed)

		err := os.MkdirAll(filepath.Join(destination, s3Backend.Bucket, dir), os.ModePerm)
		if err != nil {
			return nil, err
		}

		filename := filepath.Join(destination, s3Backend.Bucket, dir, transformed)
		filenames[filename] = fmt.Sprintf("arn:aws:s3:::%s/%s", s3Backend.Bucket, key)

		if _, err := os.Stat(filename); !os.IsNotExist(err) && !options.Overwrite {
			// file already exists
			continue
		}

		file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			return nil, err
		}

		objects = append(objects, s3manager.BatchDownloadObject{
			Object: &s3.GetObjectInput{
				Bucket: aws.String(s3Backend.Bucket),
				Key:    aws.String(key),
			},
			Writer: file,
		})
	}

	if len(objects) > 0 {
		client := s3.New(sess, conf)
		manager := s3manager.NewDownloaderWithClient(client)
		iter := &s3manager.DownloadObjectsIterator{Objects: objects}
		if err := manager.DownloadWithIterator(aws.BackgroundContext(), iter); err != nil {
			return nil, err
		}
	}

	return filenames, nil
}

type TerraformBackends struct {
	Destination string       `json:"destination"`
	Options     *Options     `json:"options"`
	S3          []*S3Backend `json:"s3"`

	StateFilenames map[string]string
}

func (t *TerraformBackends) Pull() error {
	t.StateFilenames = map[string]string{}
	for _, backend := range t.S3 {
		filenames, err := backend.Download(t.Destination, t.Options)
		if err != nil {
			return err
		}
		for filename, s3 := range filenames {
			t.StateFilenames[filename] = s3
		}
	}
	return nil
}

type ResourceMap map[string]string

func (t *TerraformBackends) Load() (ResourceMap, error) {
	managed := ResourceMap{}

	for filename, s3 := range t.StateFilenames {
		resources, err := LoadStateFromFile(filename)
		if err != nil {
			fmt.Println("Failed to load state", filename, err)
			continue
		}
		for _, resource := range resources {
			managed[resource.UniqueID()] = s3
		}
	}

	return managed, nil
}

func NewTerraformBackends(filename string) (*TerraformBackends, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	result := &TerraformBackends{
		Options: &Options{
			PathSubstitutions: []Substitution{},
			Overwrite:         false,
		},
		S3: []*S3Backend{},
	}
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}

	if len(result.Destination) == 0 {
		return nil, errors.New("Destination field is empty")
	}

	if len(result.S3) == 0 {
		return nil, errors.New("s3 field is empty")
	}

	return result, nil
}

func LoadStateFromFile(filename string) ([]*Resource, error) {
	output := []*Resource{}
	reader, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	state, err := statefile.Read(reader)
	if err != nil {
		return nil, err
	}

	filter := &states.Filter{State: state.State}
	results, err := filter.Filter()
	for _, result := range results {
		switch result.Value.(type) {
		case *states.Resource:
			// process
		default:
			continue
		}

		resource := result.Value.(*states.Resource)
		for _, resourceInstance := range resource.Instances {

			if resourceInstance.Current == nil {
				continue
			}

			attr := resourceInstance.Current.AttrsFlat

			additional := &Resource{
				ID: attr["id"],
			}

			if attr["arn"] == "" {
				switch resource.Addr.Type {
				case "aws_iam_access_key":
				case "aws_route53_record":
				case "aws_route53_zone":
				default:
					continue
				}
			} else {
				additional.ARN = attr["arn"]
			}

			output = append(output, additional)
		}

	}

	return output, nil
}
