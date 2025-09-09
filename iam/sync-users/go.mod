module github.com/hamstah/awstools/iam/sync-users

go 1.25.1

require (
	github.com/alecthomas/kingpin/v2 v2.4.0
	github.com/aws/aws-sdk-go v1.55.8
	github.com/hamstah/awstools/common v0.0.0-20250311132610-4c1ba75c7dd5
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/alecthomas/units v0.0.0-20240927000941-0f3dac36c52b // indirect
	github.com/hakobe/paranoidhttp v0.3.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/xhit/go-str2duration/v2 v2.1.0 // indirect
	golang.org/x/crypto v0.42.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
)

replace github.com/hamstah/awstools/common => ../../common
