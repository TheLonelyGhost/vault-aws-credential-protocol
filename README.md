# Vault-vended AWS credentials

Using the [`credential_process` feature](https://docs.aws.amazon.com/sdkref/latest/guide/feature-process-credentials.html) supported by modern versions of the AWS SDK, we are able to use the existing `VAULT_ADDR` and `VAULT_TOKEN` values to bridge the gap between AWS credentials vended by HashiCorp Vault and the various consuming applications.

## Examples

```
# File: ~/.aws/config

[profile vault-dev]
# When talking to Vault with `export VAULT_ADDR=https://vault.example.com:8200`,
# be sure to authenticate and set `export VAULT_TOKEN='<your token>'` as well.
credential_process = /usr/local/bin/vault-aws-credential-process --mount 'my-namespace/aws/dev' --role 'super-admin'
region = us-east-2

[profile vault-uat]
# Sometimes you already have a way of managing Vault authentication, such as
# Vault Proxy. In those cases, set the associated environment variables and
# configure it the same way. Be sure to also skip the caching measures this
# process puts in place and defer entirely to the caching Vault Proxy does for
# you.
credential_process = /usr/bin/env VAULT_ADDR=http://127.0.0.1:8200 /usr/local/bin/vault-aws-credential-process --mount 'my-namespace/aws/uat' --role 'devops-admin' --no-cache
region = us-west-2
```
