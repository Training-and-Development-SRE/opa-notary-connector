package notary

import (
	"encoding/hex"
	"testing"

	"github.com/sighupio/opa-notary-connector/config"
	"github.com/sirupsen/logrus"
	"github.com/theupdateframework/notary"
	"github.com/theupdateframework/notary/client"
	"github.com/theupdateframework/notary/tuf/data"
)

func TestRepository(t *testing.T) {
	log := logrus.NewEntry(&logrus.Logger{})
	conf := config.NewConfig()
	signer := &config.Signer{
		Role:      "targets/sighup",
		PublicKey: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURCakNDQWU0Q0NRRHlNeEhRRG5WTWd6QU5CZ2txaGtpRzl3MEJBUXNGQURCRk1Rc3dDUVlEVlFRR0V3SkIKVlRFVE1CRUdBMVVFQ0F3S1UyOXRaUzFUZEdGMFpURWhNQjhHQTFVRUNnd1lTVzUwWlhKdVpYUWdWMmxrWjJsMApjeUJRZEhrZ1RIUmtNQjRYRFRFNU1URXlOVEE1TlRBME9Wb1hEVEl3TVRFeU5EQTVOVEEwT1Zvd1JURUxNQWtHCkExVUVCaE1DUVZVeEV6QVJCZ05WQkFnTUNsTnZiV1V0VTNSaGRHVXhJVEFmQmdOVkJBb01HRWx1ZEdWeWJtVjAKSUZkcFpHZHBkSE1nVUhSNUlFeDBaRENDQVNJd0RRWUpLb1pJaHZjTkFRRUJCUUFEZ2dFUEFEQ0NBUW9DZ2dFQgpBTGcrYW1QNTVESjNaYUdLOTdSSXlLaEc5L1I1TEltVWJjaEpraHcrMFdoZUNtUm1HK1M2SlkvbG9DbXhKcTVOClJ3N0lKYmhYQytHRm13MitsTVlZS1I0QjY0UTZVRkdreVN6cndFMWtzVU5JbXZkL3dCRmtmNUJpb3g3eUFWSTAKZUx0T0V0dGdCZUxRb3JaWi8yQWRNYlpSQjFGQ3craXYvaWs3SDJLcGhJdDg0bWNmOXhoUDI5Wmcvcyt5aHVSUQo5bm5yTnNNRUQrNkZYald1QlI4aFZmanhZcHlPUmdWeUVZSDdJZXhLWkR6ckZjMHZvdlNXbVkvTURJZUozN3VHCnpvcU5SMUxGeEtMblduYzNubXZWUXJpajJ1VENjSWVtYW90MW95Z0ZMMXJFRzR2aGxTTjVhT3YyOGFqNjRJNVMKS09mTUU2MU9PODdaZlBYcmoxNFJLdWtDQXdFQUFUQU5CZ2txaGtpRzl3MEJBUXNGQUFPQ0FRRUFUMEFuek1DNApNdVFoQk5lS1BBV2s2QnllR041ckdHL0hUQStKREhzOHIxb3lRRldpRlJWZ0FmbXNXMlVEV0JvWVN6VGVUdFpUCjMvTDJ1RFg4UmNweEtWb1RoeVRxeDdOY04rZE1lK3BLSHEvbkZqekdrTFR3clI4UkRtQ1RkNXN1SEhlUzZTNm0KOEFFaC9oTVpaaVRqM241czdzRGFuamorYWowcklsNVEyanNOaXFOanUraS9odDcvemJYRnRSV3RFMDJEREpKeQpCL01MODVTcDdEUHZOLzB3SEg0dUxYU0hZRnZ3ODhONEJJMmZjM0FsR1R0QTdDQ0MyZCtiRzNRUDI0UmdDaUpSClhjTit2VWlIc3JncHMxcDMvUkRJQTRQYURGOXdtYXgxM2tqeHZCVTZxSzg4UFdqdzlUbThNNzREWU4wd21FU0kKLytmajJ3aG1oNDg1aGc9PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==",
	}
	conf.Repositories = config.Repositories{
		config.Repository{
			Name: "docker.io/.*",
			//TODO remove namespace logic
			Namespace: "sighup",
			Trust: config.Trust{
				Enabled:     true,
				TrustServer: "sighup.notary.com",
				Signers: []*config.Signer{
					signer,
				},
			},
		},
	}
	// Validate also initialize mutexes in repository and signers => shit
	err := conf.Validate(log)
	if err != nil {
		t.Errorf("Got error %s", err.Error())
		return
	}
	pubKey, err := signer.GetPEM(log)
	if err != nil {
		t.Errorf("Got error %s", err.Error())
		return
	}
	var fakeMetadataGetter AllTargetMetadataByNameGetter = Fake{
		map[string][]client.TargetSignedStruct{
			"latest": {
				client.TargetSignedStruct{
					Role: data.DelegationRole{
						BaseRole: data.BaseRole{
							Name: "targets/sighup",
							Keys: map[string]data.PublicKey{
								(*pubKey).ID(): *pubKey,
							},
						},
					},
					Target: client.Target{
						Hashes: data.Hashes{
							notary.SHA256: []byte("sighup"),
						},
					},
				},
			},
		}}
	var tests = []struct {
		image              string
		repo               *config.Repository
		fakeMetadataGetter *AllTargetMetadataByNameGetter
		expectedSha        string
	}{
		{image: "docker.io/library/alpine", fakeMetadataGetter: &fakeMetadataGetter, repo: &conf.Repositories[0], expectedSha: hex.EncodeToString([]byte("sighup"))},
	}
	for _, tt := range tests {
		t.Run(tt.image, func(t *testing.T) {
			repo, err := NewWithGetter(tt.image, tt.repo, tt.fakeMetadataGetter, log)
			if err != nil {
				t.Errorf("Got error %s", err.Error())
				return
			}

			sha, err := repo.GetSha()
			if err != nil {
				t.Errorf("Got error %s", err.Error())
				return
			}

			if sha != tt.expectedSha {
				t.Errorf("Got %s, expected %s", sha, tt.expectedSha)
				return
			}

		})
	}
}
