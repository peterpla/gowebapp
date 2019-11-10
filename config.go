package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"

	cloudkms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/storage"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

// Cfg holds the app-wide configuration
var Cfg config

// loadFlagsAndConfig reads the configuration file from Cloud Storage and decrypts it using Cloud KMS.
//
// (gdeploy.sh deploys the app to Google App Engine, encrypting the local
// configuration file using Cloud KMS and writing it to Cloud Storage.)
func loadFlagsAndConfig(Cfg *config) error {
	// log.Printf("Entering, Cfg: %+v", Cfg)

	// ***** ***** process command line flags ***** *****
	// appname --port=8080 --v --help
	pflag.IntVar(&Cfg.port, "port", 8080, "--port=8080 to listen on port :8080")
	pflag.BoolVar(&Cfg.verbose, "v", false, "--v to enable verbose output")
	pflag.BoolVar(&Cfg.help, "help", false, "")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	if Cfg.help {
		fmt.Fprintf(os.Stdout, "%s\n", helpText)
		os.Exit(0)
	}
	// log.Printf("After pflag.Parse(), Cfg: %+v", Cfg)

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
	Cfg.encryptedBucket = os.Getenv("ENCRYPTED_BUCKET")
	Cfg.storageLocation = os.Getenv("STORAGE_LOCATION")
	Cfg.configFile = os.Getenv("CONFIG_FILE")
	configFileEncrypted := Cfg.configFile + ".enc"
	// log.Printf("After os.Getenv() GCS env vars, Cfg: %+v", Cfg)

	r, err := client.Bucket(Cfg.encryptedBucket).Object(configFileEncrypted).NewReader(ctx)
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
	Cfg.projectID = os.Getenv("PROJECT_ID")     // "elated-practice-224603"
	Cfg.kmsLocation = os.Getenv("KMS_LOCATION") // "us-west2"
	Cfg.kmsKeyRing = os.Getenv("KMS_KEYRING")   // "devkeyring"
	Cfg.kmsKey = os.Getenv("KMS_KEY")           // "config"
	// log.Printf("After os.Getenv() KMS env vars, Cfg: %+v", Cfg)

	keyName := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s",
		Cfg.projectID, Cfg.kmsLocation, Cfg.kmsKeyRing, Cfg.kmsKey)

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

	Cfg.appName = viper.GetString("AppName")
	Cfg.description = viper.GetString("Description")
	Cfg.version = viper.GetString("Version")
	// log.Printf("After ReadConfig(), Cfg: %+v", Cfg)

	// bind env vars to Viper
	viper.BindEnv("projectID", "PROJECT_ID")
	viper.BindEnv("storageLocation", "STORAGE_LOCATION")
	viper.BindEnv("kmsKey", "KMS_KEY")
	viper.BindEnv("kmsKeyRing", "KMS_KEYRING")
	viper.BindEnv("kmsLocation", "KMS_LOCATION")
	viper.AutomaticEnv()
	// unmarshall all bound flags and env vars to Cfg
	err = viper.Unmarshal(&Cfg)
	if err != nil {
		return err
	}

	// log.Printf("Exiting, after BindEnv() and Unmarshal(), Cfg: %+v\n", Cfg)
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
