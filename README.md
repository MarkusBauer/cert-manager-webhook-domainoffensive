# ACME webhook for Domain-Offensive (do.de)

This solver can be used when you want to use cert-manager with do.de DNS API (documentation in customer portal).

TODO usage


### Running the test suite

First, edit [domainoffensive-secret.yml](testdata/domainoffensive-solver/domainoffensive-secret.yml) and insert your token.

You can run the test suite with:

```bash
$ TEST_ZONE_NAME=example.com. make test
```
