package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/elazarl/goproxy"
	"github.com/hamstah/awstools/common"
	"golang.org/x/net/proxy"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	socksProxy = kingpin.Flag("socks-proxy", "Socks proxy host:port to use.").String()
	bind       = kingpin.Flag("bind", "Address to bind to").Default(":8080").String()
)

func redirect(r *http.Request, location, body string) *http.Response {
	resp := &http.Response{}
	resp.Request = r
	resp.TransferEncoding = r.TransferEncoding
	resp.Header = make(http.Header)
	resp.Header.Add("Content-Type", goproxy.ContentTypeText)
	resp.Header.Add("Location", location)
	resp.StatusCode = 302
	buf := bytes.NewBufferString(body)
	resp.ContentLength = int64(buf.Len())
	resp.Body = ioutil.NopCloser(buf)
	return resp
}

func forbidden(r *http.Request, body string) *http.Response {
	resp := &http.Response{}
	resp.Request = r
	resp.TransferEncoding = r.TransferEncoding
	resp.Header = make(http.Header)
	resp.Header.Add("Content-Type", goproxy.ContentTypeText)
	resp.StatusCode = 403
	buf := bytes.NewBufferString(body)
	resp.ContentLength = int64(buf.Len())
	resp.Body = ioutil.NopCloser(buf)
	return resp
}

func setupSocks(p *goproxy.ProxyHttpServer, socksProxy string) error {
	dialer, err := proxy.SOCKS5("tcp", socksProxy, nil, proxy.Direct)
	if err != nil {
		return err
	}
	p.ConnectDial = dialer.Dial
	p.Tr.Dial = dialer.Dial
	return nil
}

func main() {
	kingpin.CommandLine.Name = "iam-auth-proxy"
	kingpin.CommandLine.Help = "Proxy to generate IAM auth token"
	flags := common.HandleFlags()

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = false

	if *socksProxy != "" {
		err := setupSocks(proxy, *socksProxy)
		common.FatalOnError(err)
	}

	authHeaderRegex, err := regexp.Compile("IAM realm=\"([^\"]+)\"")
	common.FatalOnError(err)

	proxy.OnResponse().DoFunc(
		func(r *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
			if r == nil {
				return r
			}
			if r.StatusCode == 401 {
				authHeaders := r.Header["Www-Authenticate"]
				if len(authHeaders) != 1 {
					return r
				}

				authHeader := authHeaders[0]
				realm := authHeaderRegex.FindStringSubmatch(authHeader)
				if len(realm) != 2 {
					// does not look like the auth scheme we support
					return r
				}

				// get the KMS key Id
				kmsKeyIdHeader := r.Header["Iam-Auth-Kms-Key-Id"]
				if len(kmsKeyIdHeader) == 0 {
					return forbidden(r.Request, "Invalid auth headers returned by the server: Misssing KMS key ID")
				}
				kmsKeyId := kmsKeyIdHeader[0]

				// get the KMS encryption context
				kmsEncryptionContextHeader := r.Header["Iam-Auth-Kms-Encryption-Context"]
				if len(kmsEncryptionContextHeader) == 0 {
					return forbidden(r.Request, "Invalid auth headers returned by the server: Missing KMS encryption context")
				}
				kmsEncryptionContextBase64 := kmsEncryptionContextHeader[0]

				jsonEncryptionContext, err := base64.StdEncoding.DecodeString(kmsEncryptionContextBase64)
				if len(kmsEncryptionContextHeader) == 0 {
					return forbidden(r.Request, "Invalid auth headers returned by the server: Can't decode KMS encryption context")
				}

				encryptionContext := map[string]string{}
				err = json.Unmarshal(jsonEncryptionContext, &encryptionContext)
				if err != nil {
					log.Println(err)
					return forbidden(r.Request, "Invalid auth headers returned by the server: Can't decode KMS encryption context")
				}

				session, conf := common.OpenSession(flags)
				stsClient := sts.New(session, conf)

				identity, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
				if err != nil {
					log.Println(err)
					return forbidden(r.Request, "Could not fetch IAM identity to authenticate")
				}

				if *identity.Account != realm[1] {
					return forbidden(r.Request, fmt.Sprintf("The IAM identity does not match the server realm (expected: %s)", realm[1]))
				}

				tokenStsClient := stsClient
				if *flags.RoleArn != "" || *flags.MFASerialNumber != "" {
					// get the session token without the session
					tokenStsClient = sts.New(common.NewSession(*flags.Region))
				}

				creds, err := tokenStsClient.GetSessionToken(&sts.GetSessionTokenInput{
					DurationSeconds: aws.Int64(900),
				})
				if err != nil {
					log.Println(err)
					return forbidden(r.Request, "Could not get a session token")
				}

				serialized, err := json.Marshal(creds.Credentials)
				if err != nil {
					log.Println(err)
					return forbidden(r.Request, "Could not get a session token")
				}

				awsEncryptionContext := map[string]*string{}
				for key, value := range encryptionContext {
					awsEncryptionContext[key] = aws.String(value)
				}

				kmsClient := kms.New(session, conf)
				kmsRes, err := kmsClient.Encrypt(&kms.EncryptInput{
					Plaintext:         []byte(serialized),
					KeyId:             aws.String(kmsKeyId),
					EncryptionContext: awsEncryptionContext,
				})
				if err != nil {
					log.Println(err)
					return forbidden(r.Request, "Could not get encrypt token")
				}
				str := base64.StdEncoding.EncodeToString(kmsRes.CiphertextBlob)

				returnURL := url.URL{}
				returnURL.Path = r.Request.URL.Path
				returnURL.RawQuery = r.Request.URL.RawQuery

				newURL := url.URL{}
				newURL.Path = "/auth"
				newURL.Host = r.Request.Host
				parameters := url.Values{}
				parameters.Add("token", str)
				parameters.Add("return_url", returnURL.String())
				newURL.RawQuery = parameters.Encode()
				return redirect(r.Request, newURL.String(), "")
			}

			return r
		})
	log.Fatal(http.ListenAndServe(*bind, proxy))
}
