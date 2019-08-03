package resources

import (
	"github.com/aws/aws-sdk-go/service/acm"
)

var (
	ACMService = Service{
		Name: "acm",
		Reports: map[string]Report{
			"certificates": ACMListCertificates,
		},
	}
)

func ACMListCertificates(session *Session) *ReportResult {
	client := acm.New(session.Session, session.Config)

	result := &ReportResult{}
	result.Error = client.ListCertificatesPages(&acm.ListCertificatesInput{},
		func(page *acm.ListCertificatesOutput, lastPage bool) bool {
			for _, certificate := range page.CertificateSummaryList {
				resource, err := NewResource(*certificate.CertificateArn, certificate)
				if err != nil {
					result.Error = err
					return false
				}
				result.Resources = append(result.Resources, *resource)
			}

			return true
		})

	return result
}
