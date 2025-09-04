package main

import (
	"os"
	"testing"

	// "github.com/MarkusBauer/domain-offensive-webhook"
	acmetest "github.com/cert-manager/cert-manager/test/acme"
)

var (
	zone = os.Getenv("TEST_ZONE_NAME")
)

func TestRunsSuite(t *testing.T) {
	// The manifest path should contain a file named config.json that is a
	// snippet of valid configuration that should be included on the
	// ChallengeRequest passed as part of the test cases.
	//

	// Uncomment the below fixture when implementing your custom DNS provider
	//fixture := acmetest.NewFixture(&domainoffensiveDNSProviderSolver{},
	//	acmetest.SetResolvedZone(zone),
	//	acmetest.SetAllowAmbientCredentials(false),
	//	acmetest.SetManifestPath("testdata/domainoffensive-solver"),
	//	acmetest.SetBinariesPath("_test/kubebuilder/bin"),
	//)
	solver := New()
	fixture := acmetest.NewFixture(solver,
		acmetest.SetResolvedZone("mk-bauer.de."),
		acmetest.SetManifestPath("testdata/domainoffensive-solver"),
		acmetest.SetDNSServer("192.174.68.21:53"), // ns1.domainoffensive.de
		acmetest.SetUseAuthoritative(false),
	)
	//need to uncomment RunConformance and delete runBasic and runExtended once https://github.com/cert-manager/cert-manager/pull/4835 is merged
	//fixture.RunConformance(t)
	fixture.RunBasic(t)
	fixture.RunExtended(t)

}
