# ACME webhook for Domain-Offensive (do.de)

This solver can be used when you want to use cert-manager with do.de DNS API (documentation in customer portal).


## Installation
Install the webhook with Helm:
```shell
helm repo add cert-manager-webhook-domainoffensive https://markusbauer.github.io/cert-manager-webhook-domainoffensive/
helm install --namespace cert-manager cert-manager-webhook-domainoffensive cert-manager-webhook-domainoffensive/cert-manager-webhook-domainoffensive
```

Create the necessary secret with your token from [my.do.de](https://my.do.de):
```shell
kubectl -n cert-manager create secret generic dode-token --from-literal=token="..."
```


## Usage (examples)
Create a `Issuer` like this:
```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    # The ACME server URL  # TODO change to a production URL when everything is ready
    server: https://acme-staging-v02.api.letsencrypt.org/directory

    # Email address used for ACME registration
    email: mail@example.com # TODO REPLACE THIS WITH YOUR EMAIL!!!

    # Name of a secret used to store the ACME account private key
    privateKeySecretRef:
      name: letsencrypt-staging

    solvers:
      - dns01:
          webhook:
            groupName: do.cert-manager-webhook.mk-bauer.de  # do NOT change this URL
            solverName: domain-offensive
            config:
              tokenSecretName: dode-token  # your token name (from kubectl command above)
```

Finally, you should be able to create a certificate, for example like this:
```yaml
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: your-domain-wildcard
  namespace: cert-manager
spec:
  dnsNames:  # TODO your domains
    - "your-domain.com"
    - "*.your-domain.com"
  issuerRef:
    kind: Issuer
    name: letsencrypt-staging
  secretName: your-domain-wildcard
```

### Switch from staging to production certificates
The "staging" certificates are good for testing, because they have high rate limits. But they're not trusted by browsers.
If your setup is working, you should switch to Let's Encrypt's production certificates:
- Replace `https://acme-staging-v02.api.letsencrypt.org/directory` with `https://acme-v02.api.letsencrypt.org/directory` in Issuer / ClusterIssuer
- (optional) rename the issuer from `letsencrypt-staging` to `letsencrypt-prod`, including `privateKeySecretRef` and the certificate's `issuerRef`


## Running the test suite

First, edit [domainoffensive-secret.yml](testdata/domainoffensive-solver/domainoffensive-secret.yml) and insert your token.

You can run the test suite with:

```bash
$ TEST_ZONE_NAME=example.com. make test
```
