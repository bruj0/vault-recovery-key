package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	proto "github.com/golang/protobuf/proto"
	context "golang.org/x/net/context"

	wrapping "github.com/hashicorp/go-kms-wrapping"
	"github.com/hashicorp/go-kms-wrapping/wrappers/awskms"
	"github.com/hashicorp/go-kms-wrapping/wrappers/azurekeyvault"
	"github.com/hashicorp/go-kms-wrapping/wrappers/gcpckms"

	"github.com/hashicorp/vault/shamir"
	log "github.com/sirupsen/logrus"
)

const (
	//GCP
	gcpckmsProjectID  = "rodrigo-support"
	gcpckmsLocationID = "global"
	gcpckmsKeyRing    = "vault"
	gcpckmsCryptoKey  = "vault-unsealer"
	/*Azure
	AZURE_TENANT_ID
	AZURE_CLIENT_ID
	AZURE_CLIENT_SECRET
	AZUREKEYVAULT_WRAPPER_VAULT_NAME
	AZUREKEYVAULT_WRAPPER_KEY_NAME
	*/
	version = "0.2"
)

func main() {

	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.DebugLevel)
	log.Infof("Starting version %s", version)

	cloud := flag.String("env", "gcpckms", "Environment that hosts the KMS: gcpckms,azurekeyvault,transit,awskms")
	encKey := flag.String("enc-key", "key.enc", "Path to the encrypted recovery keys from the storage, found at core/_recovery-key")
	storageType := flag.String("storage-type", "file", "Storage type: file or dynamodb")
	shares := flag.Int("shamir-shares", 1, "Number of shamir shares to divide the key into")
	threshold := flag.Int("shamir-threshold", 1, "Threshold number of keys needed for shamir creation")

	flag.Parse()

	if *cloud == "" || *encKey == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *storageType == "dynamodb" {
		// If storage type is DynamoDB, set cloud to "awskms"
		*cloud = "awskms"
	}

	log.Infof("Starting with environment %s", *cloud)

	env := []byte{}
	var err error
	if *storageType == "dynamodb" {
		// If storage type is DynamoDB, use base64 decoding
		env, err = readBinBase64Decode(*encKey)

	} else {
		// For other storage types, use regular file reading
		env, err = readBin(*encKey)
	}
	if err != nil {
		log.Fatalf("Couldnt read file: %s", err)
		os.Exit(1)
	}

	/* 	message EncryptedBlobInfo {
		// Ciphertext is the encrypted bytes
	    bytes ciphertext = 1;

		// IV is the initialization value used during encryption
	    bytes iv  = 2;

		// HMAC is the bytes of the HMAC, if any
	    bytes hmac = 3;

		// Wrapped can be used by the client to indicate whether Ciphertext
		// actually contains wrapped data or not. This can be useful if you want to
		// reuse the same struct to pass data along before and after wrapping.
	    bool wrapped = 4;

		// KeyInfo contains information about the key that was used to create this value
	    KeyInfo key_info = 5;

		// ValuePath can be used by the client to store information about where the
		// value came from
	    string ValuePath = 6;
	}
	*/
	blobInfo := &wrapping.EncryptedBlobInfo{}
	if err := proto.Unmarshal(env, blobInfo); err != nil {
		log.Errorf("failed to proto decode stored keys: %s", err)
		return
	}
	blobStr := prettyPrint(blobInfo)
	log.Debugf("blobInfo=%s", blobStr)

	var wrapper wrapping.Wrapper

	switch *cloud {
	case "gcpckms":
		wrapper, err = getWrapperGcp()
	case "azurekeyvault":
		wrapper, err = getWrapperAzure()
	case "awskms":
		wrapper, err = getWrapperAws(blobInfo)
	default:
		log.Fatalf("Environment not implemented: %s", *cloud)

	}

	if err != nil {
		log.Fatalf("Couldnt create KMS wrapper: %s", err)
	}

	pt, err := wrapper.Decrypt(context.Background(), blobInfo, nil)
	if err != nil {
		log.Errorf("failed to decrypt encrypted stored keys: %s", err)
		return
	}
	log.Debugf("HEX=%#X", pt)

	if *shares == 1 {
		encoded := base64.StdEncoding.EncodeToString([]byte(pt))
		fmt.Printf("Recovery key\n%s", encoded)
	} else {
		shares, err := shamir.Split(pt, *shares, *threshold)
		if err != nil {
			log.Fatalf("failed to generate barrier shares: %s", err)
		}
		log.Infof("Recovery keys")
		for _, share := range shares {
			fmt.Printf("%s\n", base64.StdEncoding.EncodeToString(share))
		}
	}

}

func getWrapperAws(blob *wrapping.EncryptedBlobInfo) (wrapping.Wrapper, error) {
	log.Infof("Setting up for awskms")
	s := awskms.NewWrapper(nil)
	config := map[string]string{}
	keyInfo := strings.Split(blob.GetKeyInfo().KeyID, ":")
	if len(keyInfo) == 6 {
		log.Debugf("fetching KMS key details from KeyInfo=%s", blob.GetKeyInfo().KeyID)
		config["region"] = keyInfo[3]
		config["kms_key_id"], _ = strings.CutPrefix(keyInfo[5], "key/")
	}

	_, err := s.SetConfig(config)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func getWrapperGcp() (wrapping.Wrapper, error) {
	log.Infof("Setting up for gcpckms")
	gcpCheckAndSetEnvVars()
	config := map[string]string{
		"credentials": os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
	}

	// Do an error check before env vars are set
	s := gcpckms.NewWrapper(nil)
	_, err := s.SetConfig(config)
	if err != nil {
		return nil, err
	}

	return s, nil
}
func getWrapperAzure() (wrapping.Wrapper, error) {
	log.Infof("Setting up for azurekeyvault")

	s := azurekeyvault.NewWrapper(nil)
	_, err := s.SetConfig(nil)
	if err != nil {
		return nil, err
	}
	return s, nil

}
func readBinBase64Decode(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(string(content))
	if err != nil {
		return nil, err
	}

	return decodedBytes, nil
}

func readBin(filename string) ([]byte, error) {
	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}
	defer file.Close()

	stats, statsErr := file.Stat()
	if statsErr != nil {
		return nil, statsErr
	}

	var size int64 = stats.Size()
	bytes := make([]byte, size)

	bufr := bufio.NewReader(file)
	_, err = bufr.Read(bytes)

	return bytes, err
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}
func gcpCheckAndSetEnvVars() {

	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" && os.Getenv(gcpckms.EnvGCPCKMSWrapperCredsPath) == "" {
		log.Fatal("unable to get GCP credentials via environment variables")
	}

	if os.Getenv(gcpckms.EnvGCPCKMSWrapperProject) == "" {
		os.Setenv(gcpckms.EnvGCPCKMSWrapperProject, gcpckmsProjectID)
	}

	if os.Getenv(gcpckms.EnvGCPCKMSWrapperLocation) == "" {
		os.Setenv(gcpckms.EnvGCPCKMSWrapperLocation, gcpckmsLocationID)
	}

	if os.Getenv(gcpckms.EnvVaultGCPCKMSSealKeyRing) == "" {
		os.Setenv(gcpckms.EnvVaultGCPCKMSSealKeyRing, gcpckmsKeyRing)
	}
	if os.Getenv(gcpckms.EnvGCPCKMSWrapperKeyRing) == "" {
		os.Setenv(gcpckms.EnvGCPCKMSWrapperKeyRing, gcpckmsKeyRing)
	}

	if os.Getenv(gcpckms.EnvVaultGCPCKMSSealCryptoKey) == "" {
		os.Setenv(gcpckms.EnvVaultGCPCKMSSealCryptoKey, gcpckmsCryptoKey)
	}
	if os.Getenv(gcpckms.EnvGCPCKMSWrapperCryptoKey) == "" {
		os.Setenv(gcpckms.EnvGCPCKMSWrapperCryptoKey, gcpckmsCryptoKey)
	}
}
