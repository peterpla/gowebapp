package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	cloudkms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/storage"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

// loadFlagsAndConfig reads the configuration file from Cloud Storage and decrypts it using Cloud KMS.
//
// (gdeploy.sh deploys the app to Google App Engine, encrypting the local
// configuration file using Cloud KMS and writing it to Cloud Storage.)
func loadFlagsAndConfig(cfg *config) error {
	// log.Printf("Entering, cfg: %+v", cfg)

	// ***** ***** process command line flags ***** *****
	// appname --port=8080 --v --help
	pflag.IntVar(&cfg.port, "port", 8080, "--port=8080 to listen on port :8080")
	pflag.BoolVar(&cfg.verbose, "v", false, "--v to enable verbose output")
	pflag.BoolVar(&cfg.help, "help", false, "")
	pflag.Parse()
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		log.Fatalf("error from viper.BindPFlags: %v", err)
	}

	if cfg.help {
		fmt.Fprintf(os.Stdout, "%s\n", helpText)
		os.Exit(0)
	}
	// log.Printf("After pflag.Parse(), cfg: %+v", cfg)

	/* initialize Viper, precedence:
	 * - explicit call to Set
	 * - flag
	 * - env
	 * - config
	 * - key/value store
	 * - default
	 */

	// ***** ***** read (encrypted) configuration file from Cloud Storage ***** *****
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	// retrieve env vars needed to open the config file
	cfg.encryptedBucket = os.Getenv("ENCRYPTED_BUCKET")
	cfg.storageLocation = os.Getenv("STORAGE_LOCATION")
	cfg.configFile = os.Getenv("CONFIG_FILE")
	configFileEncrypted := cfg.configFile + ".enc"
	// log.Printf("After os.Getenv() GCS env vars, cfg: %+v", cfg)

	r, err := client.Bucket(cfg.encryptedBucket).Object(configFileEncrypted).NewReader(ctx)
	if err != nil {
		return err
	}
	defer r.Close()

	// read the encrypted file contants
	cfgEncoded, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	// retrieve env vars needed to decrypt the config file contents
	cfg.projectID = os.Getenv("PROJECT_ID")     // "elated-practice-224603"
	cfg.kmsLocation = os.Getenv("KMS_LOCATION") // "us-west2"
	cfg.kmsKeyRing = os.Getenv("KMS_KEYRING")   // "devkeyring"
	cfg.kmsKey = os.Getenv("KMS_KEY")           // "config"
	// log.Printf("After os.Getenv() KMS env vars, cfg: %+v", cfg)

	keyName := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s",
		cfg.projectID, cfg.kmsLocation, cfg.kmsKeyRing, cfg.kmsKey)

	// decrypt using Cloud KMS:
	//    https://github.com/GoogleCloudPlatform/golang-samples/blob/master/kms/kms_decrypt.go
	//    https://medium.com/google-cloud/gcs-kms-and-wrapped-secrets-e5bde6b0c859
	kmsClient, err := cloudkms.NewKeyManagementClient(ctx)
	if err != nil {
		return err
	}
	dresp, err := kmsClient.Decrypt(ctx,
		&kmspb.DecryptRequest{
			Name:       keyName,
			Ciphertext: cfgEncoded,
		})
	if err != nil {
		return err
	}
	// dresp.Plaintext is the decrypted config file contents
	// log.Printf("Config: %s", dresp.Plaintext)

	// pass (decrypted) configuration file to Viper:
	//    https://github.com/spf13/viper
	viper.SetConfigType("yaml")
	err = viper.ReadConfig(bytes.NewBuffer(dresp.Plaintext))
	if err != nil {
		return err
	}

	cfg.appName = viper.GetString("AppName")
	cfg.description = viper.GetString("Description")
	cfg.version = viper.GetString("Version")
	// log.Printf("After ReadConfig(), cfg: %+v", cfg)

	// bind env vars to Viper
	if err := viper.BindEnv("projectID", "PROJECT_ID"); err != nil {
		log.Fatalf("error from viper.BindEnv: %v", err)
	}
	if err := viper.BindEnv("storageLocation", "STORAGE_LOCATION"); err != nil {
		log.Fatalf("error from viper.BindEnv: %v", err)
	}
	if err := viper.BindEnv("kmsKey", "KMS_KEY"); err != nil {
		log.Fatalf("error from viper.BindEnv: %v", err)
	}
	if err := viper.BindEnv("kmsKeyRing", "KMS_KEYRING"); err != nil {
		log.Fatalf("error from viper.BindEnv: %v", err)
	}
	if err := viper.BindEnv("kmsLocation", "KMS_LOCATION"); err != nil {
		log.Fatalf("error from viper.BindEnv: %v", err)
	}
	viper.AutomaticEnv()
	// unmarshall all bound flags and env vars to cfg
	err = viper.Unmarshal(cfg)
	if err != nil {
		return err
	}

	// log.Printf("Exiting, after BindEnv() and Unmarshal(), cfg: %+v\n", cfg)
	return nil
}

type config struct {
	appName         string
	configFile      string
	description     string
	encryptedBucket string
	kmsKey          string
	kmsKeyRing      string
	kmsLocation     string
	port            int
	projectID       string
	storageLocation string
	verbose         bool
	version         string
	help            bool
}

var helpText = `
appname
--help to display this usage info
--port=80 to listen on port :80 (default is :8080)
--v to enable verbose output
`
