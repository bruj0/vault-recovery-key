package main

import (
	"bufio"
	"fmt"
	"os"

	proto "github.com/golang/protobuf/proto"
	context "golang.org/x/net/context"

	wrapping "github.com/hashicorp/go-kms-wrapping"
	"github.com/hashicorp/go-kms-wrapping/wrappers/gcpckms"

	"encoding/base64"

	log "github.com/sirupsen/logrus"
)

const (
	// These values need to match the values from the hc-value-testing project
	gcpckmsProjectID  = "rodrigo-support"
	gcpckmsLocationID = "global"
	gcpckmsKeyRing    = "vault"
	gcpckmsCryptoKey  = "vault-unsealer"
)

func main() {

	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.DebugLevel)
	log.Info("Starting")

	checkAndSetEnvVars()
	config := map[string]string{
		"credentials": os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
	}

	// Do an error check before env vars are set
	s := gcpckms.NewWrapper(nil)
	_, err := s.SetConfig(config)
	if err != nil {
		log.Fatalf("config error %s", err)
	}

	env, _ := ReadBin("key.enc")

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
	log.Debugf("blobInfo=%v", blobInfo)

	pt, err := s.Decrypt(context.Background(), blobInfo, nil)
	if err != nil {
		log.Errorf("failed to decrypt encrypted stored keys: %s", err)
		return
	}
	log.Debugf("HEX=%#X", pt)
	encoded := base64.StdEncoding.EncodeToString([]byte(pt))
	fmt.Printf("BASE64=%s", encoded)
	// Now test for cases where CKMS values are provided

}
func ReadBin(filename string) ([]byte, error) {
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

func checkAndSetEnvVars() {

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
