# ACME webhook for Domain-Offensive (do.de)

This solver can be used when you want to use cert-manager with do.de DNS API (documentation in customer portal).

See [https://github.com/MarkusBauer/cert-manager-webhook-domainoffensive](https://github.com/MarkusBauer/cert-manager-webhook-domainoffensive) for detailed instructions.


## Installation
Install the webhook with Helm:
```shell
helm repo add cert-manager-webhook-domainoffensive https://markusbauer.github.io/cert-manager-webhook-domainoffensive/
helm install --namespace cert-manager cert-manager-webhook-domainoffensive cert-manager-webhook-domainoffensive/cert-manager-webhook-domainoffensive
```

See [https://github.com/MarkusBauer/cert-manager-webhook-domainoffensive](https://github.com/MarkusBauer/cert-manager-webhook-domainoffensive) for detailed instructions.
