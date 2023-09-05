package common

import (
	"encoding/xml"
	"errors"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/private/protocol/xml/xmlutil"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/hakobe/paranoidhttp"
)

func STSGetIdentityURL(stsClient *sts.STS) (string, error) {
	request, _ := stsClient.GetCallerIdentityRequest(&sts.GetCallerIdentityInput{})
	return request.Presign(10)
}

func STSFetchIdentityURL(identityURL string, maxAge time.Duration) (*sts.GetCallerIdentityOutput, error) {
	url, err := url.Parse(identityURL)
	if err != nil {
		return nil, err
	}

	query := url.Query()

	date, err := time.Parse("20060102T150405Z", query.Get("X-Amz-Date"))
	if err != nil {
		return nil, err
	}

	if url.Host != "sts.amazonaws.com" || query.Get("Action") != "GetCallerIdentity" || url.Scheme != "https" {
		return nil, errors.New("url is not a valid sts:GetCallerIdentity call")
	}

	if time.Now().UTC().Sub(date) > maxAge {
		return nil, errors.New("url signature is too old")
	}

	res, err := paranoidhttp.DefaultClient.Get(identityURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	decoder := xml.NewDecoder(res.Body)
	identity := &sts.GetCallerIdentityOutput{}
	err = xmlutil.UnmarshalXML(identity, decoder, "GetCallerIdentityResult")
	if err != nil {
		return nil, err
	}

	return identity, nil
}
