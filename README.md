# vault-policy-backup

Using this CLI tool, you can backup Vault Policies from a Vault instance to your local machine! :D

Note: The tool is written in Golang and uses Vault Official Golang API. The official Vault API documentation is here - https://pkg.go.dev/github.com/hashicorp/vault/api

Note: The tool needs Vault credentials of a user/account that has access to Vault, to read the Vault Policies

Note: We have tested this only with some versions of Vault (like v1.15.x). So beware to test this in a testing environment with whatever version of Vault you are using, before using this in critical environments like production! Also, ensure that the testing environment is as close to your production environment as possible so that your testing makes sense

Note: This does NOT backup the `root` Vault Policy as Vault does not support updating it / changing it - which would happen during the restore process - that is, reading from the Vault will succeed as part of backup, but later when we are writing to Vault as part of restore process, it will fail / throw error. So, no point in backing up an empty `root` policy to later try to restore it and get an error that `root` policy cannot be updated / changed. Also, `root` Vault Policy is just an empty policy, with no content. I believe it's just a placeholder policy which is assumed to have all the access to Vault

## Building

```bash
CGO_ENABLED=0 go build -v
```

or

```bash
make
```

## Usage

```bash
$ ./vault-policy-backup --help
usage: vault-policy-backup <vault-policy-name>

examples:

# show help
vault-policy-backup -h

# show help
vault-policy-backup --help

# backs up all vault policies from vault
# except the root policy
vault-policy-backup

# backs up allow_read policy from vault.
# it will throw an error if it does not exist
vault-policy-backup allow_read

# quietly backup all vault policies.
# this will just show dots (.) for progress
vault-policy-backup --quiet
```

# Demo

## Demo 1

The Vault we have here is a secured Vault with HTTPS API enabled and a big token for root. It has some Vault Policies configured.

I'm using the Vault Root Token here for full access to the Vault. But you don't need Vault Root Token. You just need any Vault Token / Credentials that has enough access to read Vault Policies in the Vault

```bash
$ export VAULT_ADDR='https://127.0.0.1:8200'
$ export VAULT_TOKEN="some-big-token-here"
$ export VAULT_CACERT=$HOME/vault-ca.crt

$ vault status
Key                     Value
---                     -----
Seal Type               shamir
Initialized             true
Sealed                  false
Total Shares            5
Threshold               3
Version                 1.15.6
Build Date              2024-02-28T17:07:34Z
Storage Type            raft
Cluster Name            vault-cluster-9f170feb
Cluster ID              151e903e-e1e7-541e-d089-ce8db2da0a34
HA Enabled              true
HA Cluster              https://karuppiah-vault-0:8201
HA Mode                 active
Active Since            2024-04-27T23:15:36.130464099Z
Raft Committed Index    99476
Raft Applied Index      99476

$ vault policy list
allow_secrets
allow_stage_kv_secrets
allow_test_kv_secrets
default
root

$ vault policy read allow_secrets
path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

$ vault policy read allow_stage_kv_secrets
# KV v2 secrets engine mount path is "stage-kv"
path "stage-kv/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

$ vault policy read allow_test_kv_secrets
# KV v2 secrets engine mount path is "test-kv"
path "test-kv/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

$ vault policy read default
# Allow tokens to look up their own properties
path "auth/token/lookup-self" {
    capabilities = ["read"]
}

# Allow tokens to renew themselves
path "auth/token/renew-self" {
    capabilities = ["update"]
}

# Allow tokens to revoke themselves
path "auth/token/revoke-self" {
    capabilities = ["update"]
}

# Allow a token to look up its own capabilities on a path
path "sys/capabilities-self" {
    capabilities = ["update"]
}

# Allow a token to look up its own entity by id or name
path "identity/entity/id/{{identity.entity.id}}" {
  capabilities = ["read"]
}
path "identity/entity/name/{{identity.entity.name}}" {
  capabilities = ["read"]
}


# Allow a token to look up its resultant ACL from all policies. This is useful
# for UIs. It is an internal path because the format may change at any time
# based on how the internal ACL features and capabilities change.
path "sys/internal/ui/resultant-acl" {
    capabilities = ["read"]
}

# Allow a token to renew a lease via lease_id in the request body; old path for
# old clients, new path for newer
path "sys/renew" {
    capabilities = ["update"]
}
path "sys/leases/renew" {
    capabilities = ["update"]
}

# Allow looking up lease properties. This requires knowing the lease ID ahead
# of time and does not divulge any sensitive information.
path "sys/leases/lookup" {
    capabilities = ["update"]
}

# Allow a token to manage its own cubbyhole
path "cubbyhole/*" {
    capabilities = ["create", "read", "update", "delete", "list"]
}

# Allow a token to wrap arbitrary values in a response-wrapping token
path "sys/wrapping/wrap" {
    capabilities = ["update"]
}

# Allow a token to look up the creation time and TTL of a given
# response-wrapping token
path "sys/wrapping/lookup" {
    capabilities = ["update"]
}

# Allow a token to unwrap a response-wrapping token. This is a convenience to
# avoid client token swapping since this is also part of the response wrapping
# policy.
path "sys/wrapping/unwrap" {
    capabilities = ["update"]
}

# Allow general purpose tools
path "sys/tools/hash" {
    capabilities = ["update"]
}
path "sys/tools/hash/*" {
    capabilities = ["update"]
}

# Allow checking the status of a Control Group request if the user has the
# accessor
path "sys/control-group/request" {
    capabilities = ["update"]
}

# Allow a token to make requests to the Authorization Endpoint for OIDC providers.
path "identity/oidc/provider/+/authorize" {
    capabilities = ["read", "update"]
}

# root policy cannot be read using `vault policy read`
$ vault policy read root
No policy named: root

# root policy can be read using `vault read`
# but as you can see it has no data really
$ vault read sys/policies/acl/root
Key       Value
---       -----
name      root
policy    n/a

# the below is just a comparison to show how the output
# for `vault read` would look like if the policy has
# some content in it
$ vault read sys/policies/acl/allow_test_kv_secrets
Key       Value
---       -----
name      allow_test_kv_secrets
policy    # KV v2 secrets engine mount path is "test-kv"
path "test-kv/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

```

We can see that it has three policies already configured, named `allow_secrets`, `allow_stage_kv_secrets` and `allow_test_kv_secrets` and we can also see the built-in `default` and `root` Vault Policies.

First, let's take a backup of one of the Vault policies in this Vault

```bash
$ export VAULT_ADDR='https://127.0.0.1:8200'
$ export VAULT_TOKEN="some-big-token-here"
$ export VAULT_CACERT=$HOME/vault-ca.crt

$ vault policy read allow_secrets
path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

$ ./vault-policy-backup allow_secrets

backing up the following vault policies in vault: [allow_secrets]

backing up `allow_secrets` vault policy

`allow_secrets` vault policy rules: path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

$ cat vault_policy_backup.json 
{"policies":[{"name":"allow_secrets","policy":"path \"secret/*\" {\n  capabilities = [\"create\", \"read\", \"update\", \"delete\", \"list\"]\n}\n"}]}

$ cat vault_policy_backup.json | jq
{
  "policies": [
    {
      "name": "allow_secrets",
      "policy": "path \"secret/*\" {\n  capabilities = [\"create\", \"read\", \"update\", \"delete\", \"list\"]\n}\n"
    }
  ]
}
```

As we can see, `allow_secrets` Vault Policy is backed up as a JSON file :)

Also, if you want a quiet backup, you can use the `--quiet` flag :D

Now, let's take a backup of all the Vault policies in this Vault :)

```bash
$ export VAULT_ADDR='https://127.0.0.1:8200'
$ export VAULT_TOKEN="some-big-token-here"
$ export VAULT_CACERT=$HOME/vault-ca.crt

$ ./vault-policy-backup

backing up the following vault policies in vault: [allow_secrets allow_stage_kv_secrets allow_test_kv_secrets default root]

backing up `allow_secrets` vault policy

`allow_secrets` vault policy rules: path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}


backing up `allow_stage_kv_secrets` vault policy

`allow_stage_kv_secrets` vault policy rules: # KV v2 secrets engine mount path is "stage-kv"
path "stage-kv/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}


backing up `allow_test_kv_secrets` vault policy

`allow_test_kv_secrets` vault policy rules: # KV v2 secrets engine mount path is "test-kv"
path "test-kv/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}


backing up `default` vault policy

`default` vault policy rules: 
# Allow tokens to look up their own properties
path "auth/token/lookup-self" {
    capabilities = ["read"]
}

# Allow tokens to renew themselves
path "auth/token/renew-self" {
    capabilities = ["update"]
}

# Allow tokens to revoke themselves
path "auth/token/revoke-self" {
    capabilities = ["update"]
}

# Allow a token to look up its own capabilities on a path
path "sys/capabilities-self" {
    capabilities = ["update"]
}

# Allow a token to look up its own entity by id or name
path "identity/entity/id/{{identity.entity.id}}" {
  capabilities = ["read"]
}
path "identity/entity/name/{{identity.entity.name}}" {
  capabilities = ["read"]
}


# Allow a token to look up its resultant ACL from all policies. This is useful
# for UIs. It is an internal path because the format may change at any time
# based on how the internal ACL features and capabilities change.
path "sys/internal/ui/resultant-acl" {
    capabilities = ["read"]
}

# Allow a token to renew a lease via lease_id in the request body; old path for
# old clients, new path for newer
path "sys/renew" {
    capabilities = ["update"]
}
path "sys/leases/renew" {
    capabilities = ["update"]
}

# Allow looking up lease properties. This requires knowing the lease ID ahead
# of time and does not divulge any sensitive information.
path "sys/leases/lookup" {
    capabilities = ["update"]
}

# Allow a token to manage its own cubbyhole
path "cubbyhole/*" {
    capabilities = ["create", "read", "update", "delete", "list"]
}

# Allow a token to wrap arbitrary values in a response-wrapping token
path "sys/wrapping/wrap" {
    capabilities = ["update"]
}

# Allow a token to look up the creation time and TTL of a given
# response-wrapping token
path "sys/wrapping/lookup" {
    capabilities = ["update"]
}

# Allow a token to unwrap a response-wrapping token. This is a convenience to
# avoid client token swapping since this is also part of the response wrapping
# policy.
path "sys/wrapping/unwrap" {
    capabilities = ["update"]
}

# Allow general purpose tools
path "sys/tools/hash" {
    capabilities = ["update"]
}
path "sys/tools/hash/*" {
    capabilities = ["update"]
}

# Allow checking the status of a Control Group request if the user has the
# accessor
path "sys/control-group/request" {
    capabilities = ["update"]
}

# Allow a token to make requests to the Authorization Endpoint for OIDC providers.
path "identity/oidc/provider/+/authorize" {
    capabilities = ["read", "update"]
}

$ cat vault_policy_backup.json 
{"policies":[{"name":"allow_secrets","policy":"path \"secret/*\" {\n  capabilities = [\"create\", \"read\", \"update\", \"delete\", \"list\"]\n}\n"},{"name":"allow_stage_kv_secrets","policy":"# KV v2 secrets engine mount path is \"stage-kv\"\npath \"stage-kv/*\" {\n  capabilities = [\"create\", \"read\", \"update\", \"delete\", \"list\"]\n}\n"},{"name":"allow_test_kv_secrets","policy":"# KV v2 secrets engine mount path is \"test-kv\"\npath \"test-kv/*\" {\n  capabilities = [\"create\", \"read\", \"update\", \"delete\", \"list\"]\n}\n"},{"name":"default","policy":"\n# Allow tokens to look up their own properties\npath \"auth/token/lookup-self\" {\n    capabilities = [\"read\"]\n}\n\n# Allow tokens to renew themselves\npath \"auth/token/renew-self\" {\n    capabilities = [\"update\"]\n}\n\n# Allow tokens to revoke themselves\npath \"auth/token/revoke-self\" {\n    capabilities = [\"update\"]\n}\n\n# Allow a token to look up its own capabilities on a path\npath \"sys/capabilities-self\" {\n    capabilities = [\"update\"]\n}\n\n# Allow a token to look up its own entity by id or name\npath \"identity/entity/id/{{identity.entity.id}}\" {\n  capabilities = [\"read\"]\n}\npath \"identity/entity/name/{{identity.entity.name}}\" {\n  capabilities = [\"read\"]\n}\n\n\n# Allow a token to look up its resultant ACL from all policies. This is useful\n# for UIs. It is an internal path because the format may change at any time\n# based on how the internal ACL features and capabilities change.\npath \"sys/internal/ui/resultant-acl\" {\n    capabilities = [\"read\"]\n}\n\n# Allow a token to renew a lease via lease_id in the request body; old path for\n# old clients, new path for newer\npath \"sys/renew\" {\n    capabilities = [\"update\"]\n}\npath \"sys/leases/renew\" {\n    capabilities = [\"update\"]\n}\n\n# Allow looking up lease properties. This requires knowing the lease ID ahead\n# of time and does not divulge any sensitive information.\npath \"sys/leases/lookup\" {\n    capabilities = [\"update\"]\n}\n\n# Allow a token to manage its own cubbyhole\npath \"cubbyhole/*\" {\n    capabilities = [\"create\", \"read\", \"update\", \"delete\", \"list\"]\n}\n\n# Allow a token to wrap arbitrary values in a response-wrapping token\npath \"sys/wrapping/wrap\" {\n    capabilities = [\"update\"]\n}\n\n# Allow a token to look up the creation time and TTL of a given\n# response-wrapping token\npath \"sys/wrapping/lookup\" {\n    capabilities = [\"update\"]\n}\n\n# Allow a token to unwrap a response-wrapping token. This is a convenience to\n# avoid client token swapping since this is also part of the response wrapping\n# policy.\npath \"sys/wrapping/unwrap\" {\n    capabilities = [\"update\"]\n}\n\n# Allow general purpose tools\npath \"sys/tools/hash\" {\n    capabilities = [\"update\"]\n}\npath \"sys/tools/hash/*\" {\n    capabilities = [\"update\"]\n}\n\n# Allow checking the status of a Control Group request if the user has the\n# accessor\npath \"sys/control-group/request\" {\n    capabilities = [\"update\"]\n}\n\n# Allow a token to make requests to the Authorization Endpoint for OIDC providers.\npath \"identity/oidc/provider/+/authorize\" {\n    capabilities = [\"read\", \"update\"]\n}\n"}]}

$ cat vault_policy_backup.json | jq
{
  "policies": [
    {
      "name": "allow_secrets",
      "policy": "path \"secret/*\" {\n  capabilities = [\"create\", \"read\", \"update\", \"delete\", \"list\"]\n}\n"
    },
    {
      "name": "allow_stage_kv_secrets",
      "policy": "# KV v2 secrets engine mount path is \"stage-kv\"\npath \"stage-kv/*\" {\n  capabilities = [\"create\", \"read\", \"update\", \"delete\", \"list\"]\n}\n"
    },
    {
      "name": "allow_test_kv_secrets",
      "policy": "# KV v2 secrets engine mount path is \"test-kv\"\npath \"test-kv/*\" {\n  capabilities = [\"create\", \"read\", \"update\", \"delete\", \"list\"]\n}\n"
    },
    {
      "name": "default",
      "policy": "\n# Allow tokens to look up their own properties\npath \"auth/token/lookup-self\" {\n    capabilities = [\"read\"]\n}\n\n# Allow tokens to renew themselves\npath \"auth/token/renew-self\" {\n    capabilities = [\"update\"]\n}\n\n# Allow tokens to revoke themselves\npath \"auth/token/revoke-self\" {\n    capabilities = [\"update\"]\n}\n\n# Allow a token to look up its own capabilities on a path\npath \"sys/capabilities-self\" {\n    capabilities = [\"update\"]\n}\n\n# Allow a token to look up its own entity by id or name\npath \"identity/entity/id/{{identity.entity.id}}\" {\n  capabilities = [\"read\"]\n}\npath \"identity/entity/name/{{identity.entity.name}}\" {\n  capabilities = [\"read\"]\n}\n\n\n# Allow a token to look up its resultant ACL from all policies. This is useful\n# for UIs. It is an internal path because the format may change at any time\n# based on how the internal ACL features and capabilities change.\npath \"sys/internal/ui/resultant-acl\" {\n    capabilities = [\"read\"]\n}\n\n# Allow a token to renew a lease via lease_id in the request body; old path for\n# old clients, new path for newer\npath \"sys/renew\" {\n    capabilities = [\"update\"]\n}\npath \"sys/leases/renew\" {\n    capabilities = [\"update\"]\n}\n\n# Allow looking up lease properties. This requires knowing the lease ID ahead\n# of time and does not divulge any sensitive information.\npath \"sys/leases/lookup\" {\n    capabilities = [\"update\"]\n}\n\n# Allow a token to manage its own cubbyhole\npath \"cubbyhole/*\" {\n    capabilities = [\"create\", \"read\", \"update\", \"delete\", \"list\"]\n}\n\n# Allow a token to wrap arbitrary values in a response-wrapping token\npath \"sys/wrapping/wrap\" {\n    capabilities = [\"update\"]\n}\n\n# Allow a token to look up the creation time and TTL of a given\n# response-wrapping token\npath \"sys/wrapping/lookup\" {\n    capabilities = [\"update\"]\n}\n\n# Allow a token to unwrap a response-wrapping token. This is a convenience to\n# avoid client token swapping since this is also part of the response wrapping\n# policy.\npath \"sys/wrapping/unwrap\" {\n    capabilities = [\"update\"]\n}\n\n# Allow general purpose tools\npath \"sys/tools/hash\" {\n    capabilities = [\"update\"]\n}\npath \"sys/tools/hash/*\" {\n    capabilities = [\"update\"]\n}\n\n# Allow checking the status of a Control Group request if the user has the\n# accessor\npath \"sys/control-group/request\" {\n    capabilities = [\"update\"]\n}\n\n# Allow a token to make requests to the Authorization Endpoint for OIDC providers.\npath \"identity/oidc/provider/+/authorize\" {\n    capabilities = [\"read\", \"update\"]\n}\n"
    }
  ]
}
```

Note: If you try to backup the `root` Vault Policy, nothing will be backed up, as there's nothing to read from `root` policy, it's empty -

```bash
$ export VAULT_ADDR='https://127.0.0.1:8200'
$ export VAULT_TOKEN="some-big-token-here"
$ export VAULT_CACERT=$HOME/vault-ca.crt

$ ./vault-policy-backup root

backing up the following vault policies in vault: [root]

$ cat vault_policy_backup.json 
{"policies":[]}
```

You can also notice that the `root` Vault Policy is empty if you try to read the policy using Vault CLI, but this is only possible if you are using `vault read` command, as `vault policy read` command does **NOT** workout here, probably because they (Vault CLI developers) put some check there and removed the `root` policy, hence the `No policy named: root` error as seen below -

```bash
# straightforward way to list
$ vault policy list
allow_secrets
allow_stage_kv_secrets
allow_test_kv_secrets
default
root

# straightforward way to read a policy generally
$ vault policy read allow_secrets
path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# straightforward way to read but it doesn't work for root policy
$ vault policy read root
No policy named: root

# a bit complicated way to list - needs knowledge about list API and the API
# path / endpoint
$ vault list sys/policies/acl
Keys
----
allow_secrets
allow_stage_kv_secrets
allow_test_kv_secrets
default
root

# You can also put an extra forward slash ("/") at the end of the API path like
# the below, near `acl`
$ vault list sys/policies/acl/
Keys
----
allow_secrets
allow_stage_kv_secrets
allow_test_kv_secrets
default
root

# a bit complicated way to read - needs knowledge about read API and the API
# path / endpoint
$ vault read sys/policies/acl/allow_secrets
Key       Value
---       -----
name      allow_secrets
policy    path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# a bit complicated way to read, and it works for root policy, showing that
# the `root` policy is empty
$ vault read sys/policies/acl/root
Key       Value
---       -----
name      root
policy    n/a
```

Note: If the Vault Token / Credentials used for the Vault is not valid / wrong / does not have enough access, then the tool throws errors similar to this -

```bash
$ export VAULT_ADDR='https://127.0.0.1:8200'
$ export VAULT_TOKEN="some-big-WRONG-token-here"
$ export VAULT_CACERT=$HOME/vault-ca.crt

$ ./vault-policy-backup allow_secrets

backing up the following vault policies in vault: [allow_secrets]
error reading 'allow_secrets' vault policy from vault: Error making API request.

URL: GET https://127.0.0.1:8200/v1/sys/policies/acl/allow_secrets
Code: 403. Errors:

* permission denied

$ ./vault-policy-backup 
error listing vault policies: Error making API request.

URL: GET https://127.0.0.1:8200/v1/sys/policies/acl?list=true
Code: 403. Errors:

* permission denied
```

## Future Ideas

Talking about future ideas, here are some of the ideas for the future -
- Ability to give a specific set of policies alone to be backed up from Vault. As of now it's either one or all policies, in one command. I want to allow users to just run one command and back up N number of policies. Maybe take a file as input, say YAML file, or JSON file, with all the policies to be backed up - and just run one command to do that. The file would say what're the Vault Policy names to be backed up. Something like -

```json
[
  "policies": [
    "allow_secrets",
    "allow_stage_kv_secrets",
    "allow_test_kv_secrets",
    "default"
  ]
]
```

or by providing the names of Vault policies as arguments to the CLI, or provide the ability to use either of the two or even both

- Give warning/information to user about how backup of root policy is not taken, so that they don't have to read the docs so much to understand this information. Reasoning - This is because - empty `root` Vault Policy - nothing to backup. Also, Vault does NOT support updating it / changing it - which would happen during the restore process - that is, reading from the Vault will succeed as part of backup, but later when we are writing to Vault as part of restore process, it will fail / throw error. So, no point in backing up an empty `root` policy to later try to restore it and get an error that `root` policy cannot be updated / changed

- Provide ability to give a specific file name for storing the vault policy backup json and not just use the default hard coded `vault_policy_backup.json` file name

- Get rid of `root` Vault policy from the output that says `backing up the following vault policies in vault` as it can be confusing for the user to see that the `root` Vauly policy is being backed up but in reality it's empty policy and does not have any content and does **NOT** get backed up by the tool and that's what we say too in the docs. It's a major mismatch in the tool's behaviour / display information and the tool's documentation, especially it's reality

- Consider changing the structure of the Vault Policy Backup JSON file

Why change the structure of the Vault Policy Backup JSON file? It's because current file - someone can use a file like the below for the restore -

```json
{
  "policies": [
    {
      "name": "allow_secrets",
      "policy": "path \"secret/*\" {\n  capabilities = [\"create\", \"read\", \"update\", \"delete\", \"list\"]\n}\n"
    },
    {
      "name": "allow_secrets",
      "policy": "# KV v2 secrets engine mount path is \"stage-kv\"\npath \"stage-kv/*\" {\n  capabilities = [\"create\", \"read\", \"update\", \"delete\", \"list\"]\n}\n"
    }
  ]
}
```

Notice how the above two policies have same name but different content. This can be weird and confusing. So, it's better to change the structure of the Vault Policy Backup JSON file. This is because, though the tool I have in this repository would never / would probably never create a content like the above, anyone can create a file like the above and use it for the restore process - causing the tool to restore the `allow_secrets` policy twice, but each time with different content and only the last one will be taken up and written to Vault finally and will be present in Vault finally

One good thing about the above JSON file structure is - it is clear as to what's the policy name as it's labelled as `name` and what's the policy content as it's labelled as `policy`

Consider changing the structure of the Vault Policy Backup JSON file to something like this -

```json
{
  "policies": {
    "allow_secrets": {
      "name": "allow_secrets",
      "policy": "path \"secret/*\" {\n  capabilities = [\"create\", \"read\", \"update\", \"delete\", \"list\"]\n}\n"
    }
  }
}
```


Why not the above? The same problem as before comes up, but in a different manner. There are two places where the Vault Policy name is used. One is in the `name` field and the other is in the policy's key. That's like two sources of truth. Which one do we use to store the Vault Policy name? If the `name` field is used, then it again brings back the problem of two policies having different keys but same `name` and different `policy` values, causing problems and confusion

So, we could do something like this -

```json
{
  "policies": {
    "allow_secrets": {
      "policy": "path \"secret/*\" {\n  capabilities = [\"create\", \"read\", \"update\", \"delete\", \"list\"]\n}\n"
    }
  }
}
```

Why not the above? However there's not much inside a Policy except the Policy content in Hashicorp Configuration Language (HCL) format. So, why just have one field inside the policy instead of just directly putting it in the value of the policy?

Another problematic thing about the above JSON file structure is - it is not clear as to what's `allow_secrets`, which is the policy name as it's NOT labelled as `name` but it's clear as to what's the policy content as it's labelled as `policy`

So, maybe do something like this -

```json
{
  "policies": {
    "allow_secrets": "path \"secret/*\" {\n  capabilities = [\"create\", \"read\", \"update\", \"delete\", \"list\"]\n}\n"
  }
}
```

Why? This way, it's not possible for someone to mention that same policy name as policy name comes as a key and the only value it has is a string which is the policy content

Two problematic things about the above JSON file structure is - it is not clear as to what's `allow_secrets`, which is the policy name as it's NOT labelled as `name` and it's not clear as to what's the policy content as it's NOT labelled as `policy`. But that's a tradeoff we have to take to ensure that people don't face the above mentioned problems
