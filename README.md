# Vault-vended AWS credentials

This tool bridges the gap between HashiCorp Vault's AWS secrets engine and applications that may not be "Vault-aware". Using the [`credential_process` feature](https://docs.aws.amazon.com/sdkref/latest/guide/feature-process-credentials.html), supported by modern versions of the AWS SDK, we are able to transparently refresh short-lived AWS credentials vended by Vault for use by those applications implementing a supported AWS SDK.

## Value proposition

AWS supports various forms of identity federation, primarily through the `sts:AssumeRole` API call and similar (e.g., `sts:AssumeRoleWithWebIdentity` or `sts:AssumeRoleWithSAML`). The difficulty is that revocation of these federated authorizations is difficult and, as such, many organizations set limits on how long of a session is supported before one needs to re-authenticate.

If I am in one such organization, implementing restrictive security policies around AssumeRole time limits, I may need to run a command that lasts longer than 1 hour (the default limit in a new AWS Account). Perhaps that command starts with `t` and ends with `erraform`. At the most restrictive, the TTL for these AssumeRole operations can be set no lower than 15 minutes. Rather than extending that limit, wouldn't it be great if we could seamlessly rauthenticate midway through running one of these long-running commands?

Enter: `credential_process`!

Modern AWS SDKs support this feature which executes a defined shell function and reads the JSON that is printed by it as the source of AWS credentials. It supports AssumeRole formatted ones (with a session token), IAM User credentials, and more. It also supports passing along the expiration time for the session so the application consuming this authorization can determine when it needs to reauthenticate.

Since few applications build in that logic to detect and reauthenticate, many of them request credentials every time they need to make an AWS API call. This is where caching comes in (see below).

By relying on the `vault-aws-credential-process` executable to handle caching and communication with Vault, we are able to effectively remain authenticated to AWS for as long as the Vault Token is authorized to communicate with Vault. In reality, the tool is periodically reauthenticating to AWS in the background. Security personell are able to rapidly respond and revoke Vault Tokens if abuse is detected, as with any other incident response scenario with Vault-brokered credentials.

## Authentication

This tool does not authenticate itself to Vault. Instead, it relies on outside mechanisms to authenticate and uses the authorization (i.e., Vault token) to communicate with the HashiCorp Vault service.

While it will technically operate without one, the `vault-aws-credential-process` executable works best when it communicates with HashiCorp Vault via an authenticating and caching proxy (i.e., [Vault Proxy](https://developer.hashicorp.com/vault/docs/agent-and-proxy/proxy)).

## Caching

There is a barebones caching mechanism built into the tool, with minimal security measures based on the filesystem ACLs. Some level of caching is required, per AWS's stated caveats with the protocol, since the AWS SDK will not do any kind of intelligent caching on its own. ==It is vastly preferred to use a different mechanism to authenticate and cache responses, such as [Vault Proxy](https://developer.hashicorp.com/vault/docs/agent-and-proxy/proxy), where possible.==

By default, the `vault-aws-credential-process` executable caches to a directory by the same name in the user's `XDG_CACHE_HOME`. The `$XDG_CACHE_HOME` environment variable may be overridden to change the location, or left untouched with the following default values:

| Platform | XDG_CACHE_HOME location |
|:---------|:------------------------|
| Windows  | `%LOCALAPPDATA%\cache`  |
| MacOS    | `~/Library/Caches`      |
| Linux    | `~/.cache`              |

> [!note]-
> It is recommended to periodically clear out the cache from the `$XDG_CACHE_HOME/vault-aws-credential-process`
> directory since this tool will not clear it out on its own.

## Usage

Within your [AWS config file](https://docs.aws.amazon.com/sdkref/latest/guide/file-location.html) (create it if it does not yet exist), we only need to add a new `[profile FOO]` section, where `FOO` is the name of the profile we will reference later. Add the fully qualified path to the executable and any command line arguments to it.

For example:

```ini
# File: ~/.aws/config

[profile vault-dev]
credential_process = /usr/local/bin/vault-aws-credential-process --mount 'aws/dev' --role 'super-admin'
region = us-east-2
```

If the AWS secrets engine is located in the root namespace and mounted at `aws/dev/`, with a role named `super-admin` created on the secrets engine, the program will communicate with Vault using the values defined in `VAULT_ADDR` (or `VAULT_AGENT_ADDR`) and token from `VAULT_TOKEN`. If the secrets engine was created in a namespace in Vault Enterprise, such as `my-namespace`, prefix the mount path with that namespace like so: `--mount my-namespace/aws/dev`.

For communicating with Vault Proxy, consider it using auto-auth and injecting its token into any requests received. Then consider setting `VAULT_AGENT_ADDR=http://127.0.0.1:8200` before running the program that communicates with AWS. No need for `VAULT_TOKEN` anymore! At this point, the effective AWS credentials will remain available continuously until Vault Proxy is no longer able to reauthenticate with the Vault service (forever?).

Here's what that may look like:

```ini
# File: ~/.aws/config

[profile vault-uat]
credential_process = /usr/local/bin/vault-aws-credential-process --mount 'aws/dev' --role 'super-admin' --no-cache
region = us-west-2
```

Note the `--no-cache` flag at the end, directing `vault-aws-credential-process` to forego caching and defer entirely to Vault Proxy's caching mechanisms.

```hcl
# File: /etc/vault.d/proxy.hcl
vault {
    address = "https://vault.example.com:8200"
}

auto_auth {
    method {
        type = "approle"
        config = {
            role_id_file_path = "/var/run/secrets/role-id"
            secret_id_file_path = "/var/run/secrets/secret-id"
        }
    }
}

api_proxy {
    use_auto_auth_token = true
}

cache {
    cache_static_secrets = true
}

listener "tcp" {
    address = "127.0.0.1:8200"
    tls_disable = true
}
```
