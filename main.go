package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	cloudkms "cloud.google.com/go/kms/apiv1" // replaces deprecated google.golang.org/api/cloudkms/v1
	"cloud.google.com/go/storage"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

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

// Cfg holds the app-wide configuration
var Cfg config

var helpText = `
appname
--help to display this usage info
--port=80 to listen on port :80 (default is :8080)
--v to enable verbose output
`

func main() {
	if err := loadFlagsAndConfig(&Cfg); err != nil {
		log.Fatalf("Error loading flags and configuration: %v", err)
	}
	// log.Printf("config file: %q, port: %d, verbose: %t\n", Cfg.configFile, Cfg.port, Cfg.verbose)

	h := NewHome()
	h.registerRoutes()
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/Users/peterplamondon/go/src/github.com/peterpla/gowebapp/public/favicon.ico")
	})

	port := os.Getenv("PORT") // Google App Engine complains if "PORT" env var isn't checked
	if port == "" {
		port = strconv.Itoa(Cfg.port)
	}
	log.Printf("listening on port %s\n", port)
	err := http.ListenAndServe(":"+port, nil)

	log.Printf("Error return from http.ListenAndServe: %v", err)
}

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

<<<<<<< HEAD
	// log.Printf("Exiting, after BindEnv() and Unmarshal(), Cfg: %+v\n", Cfg)
>>>>>>> a888b49ed2ba4a5f84d77e5fbaed994151e29dc4
	return nil
}

// Home - controller
type Home struct {
}

// NewHome contructor for home
func NewHome() *Home {
	return &Home{}
}

func (h Home) registerRoutes() {
	http.HandleFunc("/", h.handleHome)
	http.HandleFunc("/home", h.handleHome)
}

func (h Home) handleHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.Header().Add("Content-Security-Policy", "script-src https://stackpath.bootstrapcdn.com https://ajax.googleapis.com https://cdnjs.cloudflare.com; object-src 'none'")

	fmt.Fprintf(w, `<!DOCTYPE html>
	<html lang="en">
	<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
	<title>Bootstrap 4 Responsive Layout Example</title>
	<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css">
	<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/font-awesome/4.7.0/css/font-awesome.min.css">
	<script src="https://ajax.googleapis.com/ajax/libs/jquery/3.4.1/jquery.min.js"></script>
	<script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.14.7/umd/popper.min.js"></script>
	<script src="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/js/bootstrap.min.js"></script>
	</head>
	<body>
	<nav class="navbar navbar-expand-md navbar-dark bg-dark mb-3">
		<div class="container-fluid">
			<a href="#" class="navbar-brand mr-3">Tutorial Republic</a>
			<button type="button" class="navbar-toggler" data-toggle="collapse" data-target="#navbarCollapse">
				<span class="navbar-toggler-icon"></span>
			</button>
			<div class="collapse navbar-collapse" id="navbarCollapse">
				<div class="navbar-nav">
					<a href="#" class="nav-item nav-link active">Home</a>
					<a href="#" class="nav-item nav-link">Services</a>
					<a href="#" class="nav-item nav-link">About</a>
					<a href="#" class="nav-item nav-link">Contact</a>
				</div>
				<div class="navbar-nav ml-auto">
					<a href="#" class="nav-item nav-link">Register</a>
					<a href="#" class="nav-item nav-link">Login</a>
				</div>
			</div>
		</div>    
	</nav>
	<div class="container">
		<div class="jumbotron">
			<h1>Learn to Create Websites</h1>
			<p class="lead">In today's world internet is the most popular way of connecting with the people. At <a href="https://www.tutorialrepublic.com" target="_blank">tutorialrepublic.com</a> you will learn the essential web development technologies along with real life practice examples, so that you can create your own website to connect with the people around the world.</p>
			<p><a href="https://www.tutorialrepublic.com" target="_blank" class="btn btn-success btn-lg">Get started today</a></p>
		</div>
		<div class="row">
			<div class="col-md-6 col-lg-4 col-xl-3">
				<h2>HTML</h2>
				<p>HTML is the standard markup language for describing the structure of the web pages. Our HTML tutorials will help you to understand the basics of latest HTML5 language, so that you can create your own website.</p>
				<p><a href="https://www.tutorialrepublic.com/html-tutorial/" target="_blank" class="btn btn-success">Learn More &raquo;</a></p>
			</div>
			<div class="col-md-6 col-lg-4 col-xl-3">
				<h2>CSS</h2>
				<p>CSS is used for describing the presentation of web pages. CSS can save a lot of time and effort. Our CSS tutorials will help you to learn the essentials of latest CSS3, so that you can control the style and layout of your website.</p>
				<p><a href="https://www.tutorialrepublic.com/css-tutorial/" target="_blank" class="btn btn-success">Learn More &raquo;</a></p>
			</div>
			<div class="col-md-6 col-lg-4 col-xl-3">
				<h2>JavaScript</h2>
				<p>JavaScript is the most popular and widely used client-side scripting language. Our JavaScript tutorials will provide in-depth knowledge of the JavaScript including ES6 features, so that you can create interactive websites.</p>
				<p><a href="https://www.tutorialrepublic.com/javascript-tutorial/" target="_blank" class="btn btn-success">Learn More &raquo;</a></p>
			</div>
			<div class="col-md-6 col-lg-4 col-xl-3">
				<h2>Bootstrap</h2>
				<p>Bootstrap is a powerful front-end framework for faster and easier web development. Our Bootstrap tutorials will help you to learn all the features of latest Bootstrap 4 framework so that you can easily create responsive websites.</p>
				<p><a href="https://www.tutorialrepublic.com/twitter-bootstrap-tutorial/" target="_blank" class="btn btn-success">Learn More &raquo;</a></p>
			</div>
			<div class="col-md-6 col-lg-4 col-xl-3">
				<h2>References</h2>
				<p>Our references section outlines all the standard HTML5 tags and CSS3 properties along with other useful references such as color names and values, character entities, web safe fonts, language codes, HTTP messages, and more.</p>
				<p><a href="https://www.tutorialrepublic.com/twitter-bootstrap-tutorial/" target="_blank" class="btn btn-success">Learn More &raquo;</a></p>
			</div>
			<div class="col-md-6 col-lg-4 col-xl-3">
				<h2>FAQ</h2>
				<p>Our Frequently Asked Questions (FAQ) section is an extensive collection of FAQs that provides quick and working solution of common questions and queries related to web design and development with live demo.</p>
				<p><a href="https://www.tutorialrepublic.com/twitter-bootstrap-tutorial/" target="_blank" class="btn btn-success">Learn More &raquo;</a></p>
			</div>
		</div>
		<hr>
		<footer>
			<div class="row">
				<div class="col-md-6">
					<p>Copyright &copy; 2019 Tutorial Republic</p>
				</div>
				<div class="col-md-6 text-md-right">
					<a href="#" class="text-dark">Terms of Use</a> 
					<span class="text-muted mx-2">|</span> 
					<a href="#" class="text-dark">Privacy Policy</a>
				</div>
			</div>
		</footer>
	</div>
	</body>
	</html>                            `)
}
