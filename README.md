# Vault recovery key decrypter

This tool will decrypt your recovery keys.
Commonly used when you lost them and want to recreate a root token.
# Disclaimer
*This is not an official HashiCorp tool*

*Use it at your own risk*

# Usage
```
Usage of ./vault-recover-key:
  -enc-key string
    	Path to the encrypted recovery keys from the storage, found at core/_recovery-key (default "key.enc")
  -env string
    	Environment that hosts the KMS: gcpckms,azurekeyvault,transit (default "gcpckms")
  -shamir-shares int
    	Number of shamir shares to divide the key into (default 1)
  -shamir-threshold int
    	Threshold number of keys needed for shamir creation (default 1)
```
# Limitations
Currently only support GCP and Azure KMS.

It needs access to the KMS service which your Vault was configured with.

# Environmental variables for GCP
Example, if you KMS setup is: `projects/rodrigo-support/locations/global/keyRings/vault/cryptoKeys/vault-unsealer/cryptoKeyVersions/1`

```sh
$ export "GOOGLE_CREDENTIALS" = "service-account.json"
$ export "GOOGLE_PROJECT" = "rodrigo-support"
$ export "GOOGLE_REGION" = "global"
$ export "GCPCKMS_WRAPPER_KEY_RING" = "vault"
$ export "GCPCKMS_WRAPPER_CRYPTO_KEY" = "vault-unsealer"
```
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
## Encrypted recovery keys dump
### From a file storage:

```
$ cat core/_recovery-key  | jq -r .Value | base64 -d > key.enc
```

### From Consul
```
$ consul kv get -base64 vault/core/recovery-key  | base64 -d >  consul.key
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
