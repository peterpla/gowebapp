package config

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	cloudkms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/storage"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"

	"github.com/peterpla/gowebapp/pkg/adding"
	"github.com/peterpla/gowebapp/pkg/storage/memory"
	"github.com/peterpla/gowebapp/pkg/storage/queue"
)

// GetConfig reads the configuration file from Cloud Storage and decrypts it using Cloud KMS.
//
// (gdeploy.sh deploys the app to Google App Engine, encrypting the local
// configuration file using Cloud KMS and writing it to Cloud Storage.)
func GetConfig(cfg *Config) error {
	// log.Printf("Entering, cfg: %+v", cfg)

	// ***** ***** process command line flags ***** *****
	// appname --v --help
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

	// ***** ***** bind to environment variables ***** *****
	// bind env vars to Viper
	type binding struct {
		structField string
		envVar      string
	}

	bindings := []binding{
		{structField: "EncryptedBucket", envVar: "ENCRYPTED_BUCKET"},
		{structField: "StorageLocation", envVar: "STORAGE_LOCATION"},
		{structField: "ConfigFile", envVar: "CONFIG_FILE"},
		{structField: "ProjectID", envVar: "PROJECT_ID"},
		{structField: "StorageLocation", envVar: "STORAGE_LOCATION"},
		{structField: "KmsKey", envVar: "KMS_KEY"},
		{structField: "KmsKeyRing", envVar: "KMS_KEYRING"},
		{structField: "KmsLocation", envVar: "KMS_LOCATION"},
		{structField: "TasksLocation", envVar: "TASKS_LOCATION"},
		//
		{structField: "TaskDefaultSvcName", envVar: "TASK_DEFAULT_SERVICENAME"},
		{structField: "TaskDefaultWriteToQ", envVar: "TASK_DEFAULT_WRITE_TO_Q"},
		{structField: "TaskDefaultNextSvcToHandleReq", envVar: "TASK_DEFAULT_SVC_TO_HANDLE_REQ"},
		{structField: "TaskDefaultPort", envVar: "TASK_DEFAULT_PORT"},
		//
		{structField: "TaskInitialRequestSvcName", envVar: "TASK_INITIAL_REQUEST_SERVICENAME"},
		{structField: "TaskInitialRequestWriteToQ", envVar: "TASK_INITIAL_REQUEST_WRITE_TO_Q"},
		{structField: "TaskInitialRequestNextSvcToHandleReq", envVar: "TASK_INITIAL_REQUEST_SVC_TO_HANDLE_REQ"},
		{structField: "TaskInitialRequestPort", envVar: "TASK_INITIAL_REQUEST_PORT"},
		//
		{structField: "TaskServiceDispatchSvcName", envVar: "TASK_SERVICE_DISPATCH_SERVICENAME"},
		{structField: "TaskServiceDispatchWriteToQ", envVar: "TASK_SERVICE_DISPATCH_WRITE_TO_Q"},
		{structField: "TaskServiceDispatchNextSvcToHandleReq", envVar: "TASK_SERVICE_DISPATCH_SVC_TO_HANDLE_REQ"},
		{structField: "TaskServiceDispatchPort", envVar: "TASK_SERVICE_DISPATCH_PORT"},
	}

	for _, b := range bindings {
		if err := viper.BindEnv(b.structField, b.envVar); err != nil {
			log.Fatalf("error from viper.BindEnv: %v", err)
		}
	}
	viper.AutomaticEnv()
	// unmarshall all bound flags and env vars into cfg
	err := viper.Unmarshal(cfg)
	if err != nil {
		return err
	}

	// ***** ***** read (encrypted) configuration file from Cloud Storage ***** *****
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	configFileEncrypted := cfg.ConfigFile + ".enc"
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

	// give Viper the (decrypted) configuration file to process
	viper.SetConfigType("yaml")
	err = viper.ReadConfig(bytes.NewBuffer(dresp.Plaintext))
	if err != nil {
		return err
	}

	// add to Config struct the values read from encrypted config file
	cfg.AppName = viper.GetString("AppName")
	cfg.Description = viper.GetString("Description")
	cfg.Version = viper.GetString("Version")

	// set Config struct values based on execution environment
	cfg.IsGAE = false
	cfg.StorageType = Memory
	if os.Getenv("GAE_ENV") != "" {
		cfg.IsGAE = true
		cfg.StorageType = GCTQueue
	}

	switch cfg.StorageType {
	case Memory:
		storage := new(memory.Storage)
		cfg.Adder = adding.NewService(storage)

	case GCTQueue:
		storage := new(queue.GCT)
		cfg.Adder = adding.NewService(storage)

	default:
		panic("unsupported storageType")
	}

	// log.Printf("GetConfig exiting, cfg: %+v\n", cfg)
	return nil
}

// Type defines available storage types implementing Repository interface
type Type int

const (
	// Memory - store data in memory
	Memory Type = iota
	// Cloud Tasks queue - add data to Google Cloud Tasks queue
	GCTQueue
)

type Config struct {
	Adder           adding.Service
	AppName         string
	ConfigFile      string
	Description     string
	IsGAE           bool
	QueueName       string
	Router          http.Handler
	ServiceName     string
	NextServiceName string
	StorageType     Type
	// Key Management Service for encrypted config
	EncryptedBucket string
	KmsKey          string
	KmsKeyRing      string
	KmsLocation     string
	// Google Cloud Platform
	ProjectID       string
	StorageLocation string
	TasksLocation   string
	// port number used by each service
	TaskDefaultPort         string
	TaskInitialRequestPort  string
	TaskServiceDispatchPort string
	// queue name used by each services
	TaskDefaultWriteToQ         string
	TaskInitialRequestWriteToQ  string
	TaskServiceDispatchWriteToQ string
	// service name of each service
	TaskDefaultSvcName         string
	TaskInitialRequestSvcName  string
	TaskServiceDispatchSvcName string
	// next service in the chain to handle requests
	TaskDefaultNextSvcToHandleReq         string
	TaskInitialRequestNextSvcToHandleReq  string
	TaskServiceDispatchNextSvcToHandleReq string
	// miscellaneous
	Verbose bool
	Version string
	Help    bool
}

var helpText = `
appname
--help to display this usage info
--v to enable verbose output
`
