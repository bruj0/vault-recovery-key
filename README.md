# Vault recovery key decrypter

This tool will decrypt your recovery keys.
Commonly used when you lost them and want to recreate a root token.

Download the binary from [releases](https://github.com/bruj0/vault-recovery-key/releases/) 
# Disclaimer
*This is not an official HashiCorp tool*

*Use it at your own risk*

# Usage
```
Usage of ./vault-recover-key:
  -enc-key string
    	Path to the encrypted recovery keys from the storage, found at core/_recovery-key (default "key.enc")
  -env string
    	Environment that hosts the KMS: gcpckms,azurekeyvault,transit,awskms (default "gcpckms")
  -shamir-shares int
    	Number of shamir shares to divide the key into (default 1)
  -shamir-threshold int
    	Threshold number of keys needed for shamir creation (default 1)
  -storageType string
        Storage type: file or dynamodb
```
# Limitations
Currently only support GCP, Azure KMS and AWS KMS.

It needs access to the KMS service which your Vault was configured with.

# Environmental variables for GCP
Example, if your KMS setup is: `projects/rodrigo-support/locations/global/keyRings/name-of-keyring-goes-here/cryptoKeys/name-of-vault-unseal-key-from-keyring-goes-here/cryptoKeyVersions/1`

```sh
$ export "GOOGLE_CREDENTIALS" = "service-account.json"
$ export "GOOGLE_PROJECT" = "project-id-goes-here"
$ export "GOOGLE_REGION" = "global"
$ export "GCPCKMS_WRAPPER_KEY_RING" = "name-of-keyring-goes-here"
$ export "GCPCKMS_WRAPPER_CRYPTO_KEY" = "name-of-vault-unseal-key-from-keyring-goes-here"
```

An important note: the relevant KMS environment variables are `GCPCKMS_...` not `GCPKMS_...`. This is a potential cause of the panic issue listed below.

# Environmental variables for AZURE
If your Vault configuration is:

```
seal "azurekeyvault" {
  client_id      = "YOUR-APP-ID"
  client_secret  = "YOUR-APP-PASSWORD"
  tenant_id      = "YOUR-AZURE-TENANT-ID"
  vault_name     = "rodrigo-key-vault"
  key_name       = "generated-key"
}
```

```sh
$ export "AZURE_TENANT_ID" = "YOUR-AZURE-TENANT-ID"
$ export "AZURE_CLIENT_ID" = "YOUR-APP-ID"
$ export "AZURE_CLIENT_SECRET" = "YOUR-APP-PASSWORD"
$ export "VAULT_AZUREKEYVAULT_VAULT_NAME" = "rodrigo-key-vault"
$ export "VAULT_AZUREKEYVAULT_KEY_NAME" = "generated-key"
```

# Environmental variables for AWS KMS

```sh
export AWS_ACCESS_KEY_ID = "YOUR_ACCESS_KEY_ID_HERE"
export AWS_SECRET_ACCESS_KEY = "YOUR_SECRET_ID_HERE"
export AWS_SESSION_TOKEN = "YOUR_SESSION_TOKEN_HERE" (this one is optional, if STS creds are used)
```

Optionally, you can also set `AWSKMS_WRAPPER_KEY_ID` environment variable as well. If not set, the tool will fetch
the key details from the `key.enc`

## Encrypted recovery keys dump
### From a file storage:

```
$ cat core/_recovery-key  | jq -r .Value | base64 -d > key.enc
```

### From Consul
```
$ consul kv get -base64 vault/core/recovery-key  | base64 -d >  consul.key
```

### From MySQL
```mysql
SELECT TO_BASE64(vault_value) FROM vault WHERE vault_key = _binary 'core/recovery-key'
```

```shell
$ echo "<base64 encoded value from the above query>" | base64 > key.enc
```

## Decryption

```log
$ ./vault-recover-key -enc-key key.enc -env azurekeyvault -shamir-shares 5 -shamir-threshold 3
INFO[0000] Starting version 0.2
INFO[0000] Starting with environment azurekeyvault
INFO[0000] Setting up for azurekeyvault
DEBU[0000] blobInfo={
	"ciphertext": "sVi/u3CiFwcfiKajC0qK0+pS/St7/mReTN3mGHXN8l3TyDm/BEtGlL8ZapY+flS8",
	"iv": "09CjA+ImIFBw7yYd",
	"key_info": {
		"KeyID": "3d035268cfd34001b34d739c704ceb1f",
		"WrappedKey": "ZzRYVTNraXctLUx2ZzE2Ny1MbG9nanRvY3g5c3ZzcnBQTlNJeVdxdnFYYkJjVHR6UW14d1ZsaFBpdUVKbFliZW9qQk9UYmY5Q1hNQWpmVlAzVllsUDhtNThreW1qZl9IaFllZzAzNXdidmp3ZGZ2R1ZLV1YtSTZiOHJlVU9PdElsYTZTRmFRa3N2a0Y0cFBITGtwUVFoRG1tRVBHQ0huOXlXcUw0Q01XZWE1SDh6N2lRaGRham10cWgxRXZBS05zSWZwazVFaE9LemxWc1U1cXBQWHNhVmU5OVJiRVE1cV93aE11Y01HbzlQcU1ISGlPWmRzWGp3M25YWUc1RDNxUHRLQ3pmT2s5ZkFPUGhxNTktXzBuZm1LNVZqemtoQWpnMmNyT0F0VjVCemhNb3FNU2NhMXNXdXNpeDlId1FHVGNGTmw0SkdnRXRHb0VjMmhRUEp3MGpn"
	}
}
DEBU[0000] HEX=0X53F336750B4D68C62BCB82CA3D9689D9C4F4261C21968BBCD6803979670C29CC
INFO[0000] Recovery keys
Nemi6Ry62LDhaGqguPxYGTZpvbmueYF+8kgsv4smSXzd
wsxCRl/LixFihsxFUyZkwcuzHFNDVkgTGdghA3Y9kQWk
i7LU05v5yr5+WBD22AAvPFikKF12n24xdl4ge+n9LKKf
D5wepAo7kl9tfrQJAO5ORfCFGduW94GJWknn3sr6hOQU
oQx5/uLfVE7m3B8PhSv+aVYKYxSG0cos2EO8ZOpJwKAc
```

AWS KMS example:

```log
 ./vault-recovery-key -enc-key ../key.enc -env awskms
INFO[0000] Starting version 0.2
INFO[0000] Starting with environment awskms
INFO[0000] Setting up for awskms
DEBU[0000] blobInfo={
	"ciphertext": "ESGKIwFd/uGtnCJJV86/gwE3oCz6Yx74DMVNRTjSE4U+kVL90jiNLYvPJqh7T8pK",
	"iv": "RcvcMe4xZvI15/Gb",
	"key_info": {
		"Mechanism": 1,
		"KeyID": "arn:aws:kms:us-east-1:ACCOUNT_ID_HERE:key/KMS_KEY_ID_HERE",
		"WrappedKey": "AQICAHjy4AAl3/<<<SNIP>>>GLVdp6dcF/ih+NDDnnJastg=="
	}
}
DEBU[0000] HEX=0X4124F68E<<<SNIP>>>291DE7114A89B
DEBU[0000] fetching KMS key details from KeyInfo=arn:aws:kms:us-east-1:ACCOUNT_ID_HERE:key/KMS_KEY_ID_HERE
Recovery key
QST2jpzI44gf/mvssOsBT//FjMIFjOuxKOKR3nEUqJs=%
```

# Troubleshooting
This page lists solutions to problems you might encounter with `vault-recovery-key`

## Issue
### Application crash during runtime when using `gcpckms`:
  ```
  ./vault-recovery-key -enc-key enc.key -env gcpckms -shamir-shares 5 -shamir-threshold 3
INFO[0000] Starting version 0.2
INFO[0000] Starting with environment gcpckms
INFO[0000] Setting up for gcpckms
DEBU[0000] blobInfo={
	"ciphertext": "rWUAXlSnzRTYlA5MxQ8Cdoz32yRD9Bk6BF00oodgukmFjUmP0tR1EhZd6IvP4KkI",
	"iv": "e1YE0TZ0Z0Yfnwdj",
	"key_info": {
		"Mechanism": 1,
		"KeyID": "projects/hc-5d80c603dabb4a669f42e6354a1/locations/global/keyRings/vault-keyring/cryptoKeys/vault-key/cryptoKeyVersions/1",
		"WrappedKey": "CiQA2AIm8C9WyKu9/uUiNUYyng5nK2fKfX0ZDfR2JPupygg3P50SSAB3Uh/JATR2KCPMmXS3e6gkE3UwBXnFr3Bky06Z83lKS/7QOp4bmJXhcckML17F5MdIFyZXmrFLoi1tN44mEROYiE9TQGcUvA=="
	}
}
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x18 pc=0x56f834]

goroutine 1 [running]:
main.main()
	/home/ubuntu/vault-recovery-key/main.go:108 +0x6b4
  ```
## Solution
Check the following variables and their values. Check that the service account has been granted access to the keyring and key. Ensure that you've set the proper environment variables, most notably the key ring and crypto key variables (these env variables start with `GCPCKMS`, not `GCPKMS`). Ensure that you've set the GCP project ID and not the name of the project.

```
$ export "GOOGLE_CREDENTIALS" = "service-account.json"
$ export "GOOGLE_PROJECT" = "project-id-goes-here"
$ export "GOOGLE_REGION" = "global"
$ export "GCPCKMS_WRAPPER_KEY_RING" = "name-of-keyring-goes-here"
$ export "GCPCKMS_WRAPPER_CRYPTO_KEY" = "name-of-vault-unseal-key-from-keyring-goes-here"
```
# Additional information
Tested and verified to work against Vault 1.8.1 and 1.9.3 using `gcpckms` Auto-Unseal.
