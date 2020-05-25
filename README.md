# Vault recovery key decrypter

This tool will decrypt your recovery keys.
Commonly used when you lost them and want to recreate a root token.

# Usage
Currently only support GCP KMS. 
It needs access to the KMS service which your Vault was configured with.
Environmental variables that you need to set:

Example, if you KMS setup is: `projects/rodrigo-support/locations/global/keyRings/vault/cryptoKeys/vault-unsealer/cryptoKeyVersions/1`

## Environmental variables
```sh
$ export "GOOGLE_CREDENTIALS" = "service-account.json"
$ export "GOOGLE_PROJECT" = "rodrigo-support"
$ export "GOOGLE_REGION" = "global"
$ export "GCPCKMS_WRAPPER_KEY_RING" = "vault"
$ export "GCPCKMS_WRAPPER_CRYPTO_KEY" = "vault-unsealer"
```

## Encrypted recovery keys dump
From a file storage:

```
$ cat _recovery-key  | jq -r .Value | base64 -d > key.enc
```

## Decryption

```log
$ ./vault-recover-key
INFO[0000] Starting
DEBU[0000] blobInfo=ciphertext:"\xe0ng\x96\x96;\xa6zI\xf3\x18\x1c\x1d\x86c\x0b\xce\"y\xe6P\x9eU\xbc\xe8\xbb\x1c\xbc\xf0݉\x9c\x9di5\x89\xb7'\xff\x00\xb4\xe5\x11$\xc8x"  iv:"j'1\xee\xe4a\xa8\x04\xc1O\xc8B"  key_info:{Mechanism:1  KeyID:"projects/rodrigo-support/locations/global/keyRings/vault/cryptoKeys/vault-unsealer/cryptoKeyVersions/1"  WrappedKey:"\n$\x00\xeb\xdfR\xfb\x1eH\xdc\x1cBZ\xdb\x0b\xdc\\\xda\xf69\xe8\xc9\x17i\x85\xbc\x1f\xafW\xb7\xf23\xdft\xc6\x12I\x00\x0e\x15~\xb9\xe4aq\xc9D\x08;\x80\xbb /\x99\xf2о\x19s\xb6\xb6L\x9c\x90\xefm\xf5\xac\xfc\xe1ɇ8!,L\x0bu\x08Jc\x80\x8c\\B\xa5\xa0ײE\x8d\x90\xf6\xed#\x16\xfbc\x9b\xf9\xeb\x1dҹ?}\x9b\xfc\xe9\xb3"}
DEBU[0000] HEX=0X86D9EC2995DF01B4807F938FD42277CD28E06FFF4AC3E41F55580671F6B38607
BASE64=htnsKZXfAbSAf5OP1CJ3zSjgb/9Kw+QfVVgGcfazhgc=
```