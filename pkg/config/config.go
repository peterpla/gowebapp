package config

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

// LoadFlagsAndConfig reads the configuration file from Cloud Storage and decrypts it using Cloud KMS.
//
// (gdeploy.sh deploys the app to Google App Engine, encrypting the local
// configuration file using Cloud KMS and writing it to Cloud Storage.)
func LoadFlagsAndConfig(cfg *Config) error {
	// log.Printf("Entering, cfg: %+v", cfg)

	// ***** ***** process command line flags ***** *****
	// appname --port=8080 --v --help
	pflag.IntVar(&cfg.Port, "port", 8080, "--port=8080 to listen on port :8080")
	pflag.BoolVar(&cfg.Verbose, "v", false, "--v to enable verbose output")
	pflag.BoolVar(&cfg.Help, "help", false, "")
	pflag.Parse()
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		log.Fatalf("error from viper.BindPFlags: %v", err)
	}

	if cfg.Help {
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
	cfg.EncryptedBucket = os.Getenv("ENCRYPTED_BUCKET")
	cfg.StorageLocation = os.Getenv("STORAGE_LOCATION")
	cfg.ConfigFile = os.Getenv("CONFIG_FILE")
	configFileEncrypted := cfg.ConfigFile + ".enc"
	// log.Printf("After os.Getenv() GCS env vars, cfg: %+v", cfg)

	r, err := client.Bucket(cfg.EncryptedBucket).Object(configFileEncrypted).NewReader(ctx)
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
	cfg.ProjectID = os.Getenv("PROJECT_ID")
	cfg.KmsLocation = os.Getenv("KMS_LOCATION")
	cfg.KmsKeyRing = os.Getenv("KMS_KEYRING")
	cfg.KmsKey = os.Getenv("KMS_KEY")

	// env vars for Cloud Tasks
	cfg.TasksLocation = os.Getenv("TASKS_LOCATION")

	// log.Printf("After os.Getenv() KMS env vars, cfg: %+v", cfg)

	keyName := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s",
		cfg.ProjectID, cfg.KmsLocation, cfg.KmsKeyRing, cfg.KmsKey)

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

	cfg.AppName = viper.GetString("AppName")
	cfg.Description = viper.GetString("Description")
	cfg.Version = viper.GetString("Version")
	// log.Printf("After ReadConfig(), cfg: %+v", cfg)

	// bind env vars to Viper
	type binding struct {
		structField string
		envVar      string
	}

	// port number used by each service
	cfg.TaskDefaultPort = os.Getenv("TASK_DEFAULT_PORT")
	cfg.TaskInitialRequestPort = os.Getenv("TASK_INITIAL_REQUEST_PORT")
	cfg.TaskServiceDispatchPort = os.Getenv("TASK_SERVICE_DISPATCH_PORT")
	// queue name used by each service
	cfg.TaskDefaultWriteToQ = os.Getenv("TASK_DEFAULT_WRITE_TO_Q")
	cfg.TaskInitialRequestWriteToQ = os.Getenv("TASK_INITIAL_REQUEST_WRITE_TO_Q")
	cfg.TaskServiceDispatchWriteToQ = os.Getenv("TASK_SERVICE_DISPATCH_WRITE_TO_Q")
	// service name of each service
	cfg.TaskDefaultSvc = os.Getenv("TASK_DEFAULT_SVC")
	cfg.TaskInitialRequestSvc = os.Getenv("TASK_INITIAL_REQUEST_SVC")
	cfg.TaskServiceDispatchSvc = os.Getenv("TASK_SERVICE_DISPATCH_SVC")

	bindings := []binding{
		{structField: "ProjectID", envVar: "PROJECT_ID"},
		{structField: "StorageLocation", envVar: "STORAGE_LOCATION"},
		{structField: "KmsKey", envVar: "KMS_KEY"},
		{structField: "KmsKeyRing", envVar: "KMS_KEYRING"},
		{structField: "KmsLocation", envVar: "KMS_LOCATION"},
		{structField: "TasksLocation", envVar: "TASKS_LOCATION"},
		{structField: "TaskDefaultPort", envVar: "TASK_DEFAULT_PORT"},
		{structField: "TaskInitialRequestPort", envVar: "TASK_INITIAL_REQUEST_PORT"},
		{structField: "TaskDefaultWriteToQ", envVar: "TASK_DEFAULT_WRITE_TO_Q"},
		{structField: "TaskInitialRequestWriteToQ", envVar: "TASK_INITIAL_REQUEST_WRITE_TO_Q"},
		{structField: "TaskDefaultSvc", envVar: "TASK_DEFAULT_SVC"},
		{structField: "TaskInitialRequestSvc", envVar: "TASK_INITIAL_REQUEST_SVC"},
		{structField: "TaskServiceDispatchPort", envVar: "TASK_SERVICE_DISPATCH_PORT"},
		{structField: "TaskServiceDispatchWriteToQ", envVar: "TASK_SERVICE_DISPATCH_WRITE_TO_Q"},
		{structField: "TaskServiceDispatchSvc", envVar: "TASK_SERVICE_DISPATCH_SVC"},
	}

	for _, b := range bindings {
		// log.Printf("LoadFlagsAndConfig, viper.BindEnv(%q,%q)\n", b.structField, b.envVar)
		if err := viper.BindEnv(b.structField, b.envVar); err != nil {
			log.Fatalf("error from viper.BindEnv: %v", err)
		}
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

type Config struct {
	AppName     string
	ConfigFile  string
	Description string
	// Key Management Service for encrypted config
	EncryptedBucket string
	KmsKey          string
	KmsKeyRing      string
	KmsLocation     string
	//
	Port            int
	ProjectID       string
	StorageLocation string
	TasksLocation   string
	// port number used by each service
	TaskDefaultPort         string
	TaskInitialRequestPort        string
	TaskServiceDispatchPort string
	// queue name used by each services
	TaskDefaultWriteToQ         string
	TaskInitialRequestWriteToQ  string
	TaskServiceDispatchWriteToQ string
	// service name of each service
	TaskDefaultSvc         string
	TaskInitialRequestSvc        string
	TaskServiceDispatchSvc string
	Verbose                 bool
	Version                 string
	Help                    bool
}

var helpText = `
appname
--help to display this usage info
--port=80 to listen on port :80 (default is :8080)
--v to enable verbose output
`
