package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook"
	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"
)

var GroupName = os.Getenv("GROUP_NAME")

const ApiUrl = "https://my.do.de/api/letsencrypt"

func main() {
	if GroupName == "" {
		panic("GROUP_NAME must be specified")
	}

	// This will register our custom DNS provider with the webhook serving
	// library, making it available as an API under the provided GroupName.
	// You can register multiple DNS provider implementations with a single
	// webhook, where the Name() method will be used to disambiguate between
	// the different implementations.
	cmd.RunWebhookServer(GroupName,
		&domainoffensiveDNSProviderSolver{},
	)
}

// domainoffensiveDNSProviderSolver implements the provider-specific logic needed to
// 'present' an ACME challenge TXT record for your own DNS provider.
// To do so, it must implement the `github.com/cert-manager/cert-manager/pkg/acme/webhook.Solver`
// interface.
type domainoffensiveDNSProviderSolver struct {
	client kubernetes.Clientset
}

// domainoffensiveDNSProviderConfig is a structure that is used to decode into when
// solving a DNS01 challenge.
// This information is provided by cert-manager, and may be a reference to
// additional configuration that's needed to solve the challenge for this
// particular certificate or issuer.
// This typically includes references to Secret resources containing DNS
// provider credentials, in cases where a 'multi-tenant' DNS solver is being
// created.
// If you do *not* require per-issuer or per-certificate configuration to be
// provided to your webhook, you can skip decoding altogether in favour of
// using CLI flags or similar to provide configuration.
// You should not include sensitive information here. If credentials need to
// be used by your provider here, you should reference a Kubernetes Secret
// resource and fetch these credentials using a Kubernetes clientset.
type domainoffensiveDNSProviderConfig struct {
	// These fields will be set by users in the
	// `issuer.spec.acme.dns01.providers.webhook.config` field.

	TokenSecretRef string `json:"tokenSecretName"`
}

// Name is used as the name for this DNS solver when referencing it on the ACME
// Issuer resource.
// This should be unique **within the group name**, i.e. you can have two
// solvers configured with the same Name() **so long as they do not co-exist
// within a single webhook deployment**.
// For example, `cloudflare` may be used as the name of a solver.
func (c *domainoffensiveDNSProviderSolver) Name() string {
	return "domain-offensive"
}

// Present is responsible for actually presenting the DNS record with the
// DNS provider.
// This method should tolerate being called multiple times with the same value.
// cert-manager itself will later perform a self check to ensure that the
// solver has correctly configured the DNS provider.
func (c *domainoffensiveDNSProviderSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	token, err := c.getToken(ch)
	if err != nil {
		return err
	}

	// set a record in the DNS provider's console
	params := url.Values{}
	params.Add("token", token)
	params.Add("domain", strings.TrimRight(ch.ResolvedFQDN, "."))
	params.Add("value", ch.Key)
	err = c.doApiRequest(params)
	if err != nil {
		return err
	}

	return nil
}

// CleanUp should delete the relevant TXT record from the DNS provider console.
// If multiple TXT records exist with the same record name (e.g.
// _acme-challenge.example.com) then **only** the record with the same `key`
// value provided on the ChallengeRequest should be cleaned up.
// This is in order to facilitate multiple DNS validations for the same domain
// concurrently.
func (c *domainoffensiveDNSProviderSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	token, err := c.getToken(ch)
	if err != nil {
		return err
	}

	// delete a record from the DNS provider's console
	params := url.Values{}
	params.Add("token", token)
	params.Add("domain", strings.TrimRight(ch.ResolvedFQDN, "."))
	params.Add("value", ch.Key)
	params.Add("action", "delete")
	err = c.doApiRequest(params)
	if err != nil {
		return err
	}
	return nil
}

func (c *domainoffensiveDNSProviderSolver) getToken(ch *v1alpha1.ChallengeRequest) (string, error) {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return "", err
	}

	secretName := cfg.TokenSecretRef
	sec, err := c.client.CoreV1().Secrets(ch.ResourceNamespace).Get(context.TODO(), secretName, metav1.GetOptions{})

	if err != nil {
		return "", fmt.Errorf("unable to get secret `%s/%s`; %v", secretName, ch.ResourceNamespace, err)
	}

	data, ok := sec.Data["token"]
	if !ok {
		return "", fmt.Errorf("key %q not found in secret data of %q", "token", secretName)
	}
	return string(data), nil
}

func (c *domainoffensiveDNSProviderSolver) doApiRequest(params url.Values) error {
	response, err := http.Get(ApiUrl + "?" + params.Encode())
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode > 299 {
		return fmt.Errorf("request failed with status code %d", response.StatusCode)
	}

	var jsonResponse map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&jsonResponse)
	if err != nil {
		return err
	}
	if jsonResponse["success"] != true {
		return fmt.Errorf("operation not successful: %v", jsonResponse)
	}
	return nil
}

// Initialize will be called when the webhook first starts.
// This method can be used to instantiate the webhook, i.e. initialising
// connections or warming up caches.
// Typically, the kubeClientConfig parameter is used to build a Kubernetes
// client that can be used to fetch resources from the Kubernetes API, e.g.
// Secret resources containing credentials used to authenticate with DNS
// provider accounts.
// The stopCh can be used to handle early termination of the webhook, in cases
// where a SIGTERM or similar signal is sent to the webhook process.
func (c *domainoffensiveDNSProviderSolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	cl, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return err
	}
	c.client = *cl
	return nil
}

// loadConfig is a small helper function that decodes JSON configuration into
// the typed config struct.
func loadConfig(cfgJSON *extapi.JSON) (domainoffensiveDNSProviderConfig, error) {
	cfg := domainoffensiveDNSProviderConfig{}
	// handle the 'base case' where no configuration has been provided
	if cfgJSON == nil {
		return cfg, nil
	}
	if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
		return cfg, fmt.Errorf("error decoding solver config: %v", err)
	}

	return cfg, nil
}

func New() webhook.Solver {
	return &domainoffensiveDNSProviderSolver{}
}
