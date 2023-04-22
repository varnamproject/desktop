package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/getlantern/systray"
	flag "github.com/spf13/pflag"
	"github.com/varnamproject/desktop/icon"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/stuffbin"
)

type userConfig map[string]string

const (
	downloadPageSize = 100
)

var (
	kf = koanf.New(".")

	appAddress            string
	varnamdConfig         *config // config instance used across the application
	syncDispatcherRunning bool
	startedAt             time.Time
	buildVersion          string
	buildDate             string
	maxHandleCount        int
	authEnabled           bool

	// User accounts are stored here.
	users map[string]userConfig
)

type appConfig struct {
	Address string

	EnableInternalApis bool   `koanf:"enable-internal-api"` // internal APIs are not exposed to public
	EnableSSL          bool   `koanf:"enable-ssl"`
	CertFilePath       string `koanf:"cert-path"`
	KeyFilePath        string `koanf:"key-file-path"`

	DownloadEnabledSchemes string        `koanf:"download-enabled-schemes"`
	SyncInterval           time.Duration `koanf:"sync-interval"`
	UpstreamURL            string        `koanf:"upstream-url"`
	// Web                    bool          `koanf:"web"`

	VstDir string `koanf:"vst-dir"`
}

// App is a singleton to share across handlers.
type App struct {
	cache Cache
	log   *log.Logger
	fs    stuffbin.FileSystem
}

// varnamd configurations
// this is populated from various command line flags
type config struct {
	upstream          string
	schemesToDownload map[string]bool
	syncInterval      time.Duration
}

func init() {
	// Set max processors to number of CPUs to maximize performance.
	runtime.GOMAXPROCS(runtime.NumCPU())

	//  Setup flags to read from user to start the application.
	// Initialize 'config' flagset.
	flagSet := flag.NewFlagSet("config", flag.ContinueOnError)
	flagSet.Usage = func() {
		log.Fatal(flagSet.FlagUsages())
	}

	// Create  config flag to read 'config.toml' from user.
	flagSet.String("config", "config.toml", "Path to the TOML configuration file")

	flagSet.Int("p", 8080, "Run daemon in specified port")
	flagSet.String("host", "", "Host for the varnam daemon server")

	// Create flag for version check.
	flagSet.Bool("version", false, "Current version of the build")

	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("error parsing flags: %v", err)
	}

	// Load commandline params user given.
	if err = kf.Load(posflag.Provider(flagSet, ".", kf), nil); err != nil {
		log.Fatal(err.Error())
	}

	// Handle --version flag. Print build version, build date and die.
	if kf.Bool("version") {
		fmt.Printf("Commit: %v\nBuild: %v\n", buildVersion, buildDate)
		os.Exit(0)
	}

	// Load the config file.
	log.Printf("reading config: %s", kf.String("config"))

	if err = kf.Load(file.Provider(kf.String("config")), toml.Parser()); err != nil {
		log.Printf("error reading config: %v", err)
	}
}

func syncRequired() bool {
	return len(varnamdConfig.schemesToDownload) > 0
}

// Starts the sync process only if it is not running
func startSyncDispatcher() {
	if syncRequired() && !syncDispatcherRunning {
		sync := newSyncDispatcher(varnamdConfig.syncInterval / time.Second)
		sync.start()
		sync.runNow() // run one round of sync immediatly rather than waiting for the next interval to occur

		syncDispatcherRunning = true
	}
}

func openInBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func onTrayReady() {
	// systray.SetIcon(icon.Data)
	systray.SetTitle("Varnam Desktop")
	systray.SetTemplateIcon(icon.Data, icon.Data)

	mOpen := systray.AddMenuItem("Open", "Open in browser")
	go func() {
		<-mOpen.ClickedCh
		openInBrowser(appAddress)
	}()

	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	go func() {
		<-mQuit.ClickedCh
		fmt.Println("Requesting quit")
		systray.Quit()
		fmt.Println("Finished quitting")
	}()
}

func main() {
	fmt.Printf("aaa")

	config, err := initAppConfig()
	if err != nil {
		log.Fatal(err.Error())
	}

	if config.VstDir != "" {
		os.Setenv("VARNAM_VST_DIR", config.VstDir)
	}

	appAddress = "http://" + config.Address

	maxHandleCount = kf.Int("app.max-handles")
	if maxHandleCount <= 0 {
		maxHandleCount = 10
	}

	authEnabled = kf.Bool("app.accounts-enabled")
	if authEnabled {
		if err = kf.Unmarshal("users", &users); err != nil {
			log.Fatal(err.Error())
		}
	}

	varnamdConfig = initConfig(config)
	startedAt = time.Now()

	log.Printf("varnamd %s-%s", buildVersion, buildDate)

	fs, err := initVFS()
	if err != nil {
		log.Fatal(err.Error())
	}

	app := &App{
		cache: NewCache(),
		log:   log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile),
		fs:    fs,
	}

	startSyncDispatcher()
	go startDaemon(app, config)

	// Wait a bit for daemon to start
	time.Sleep(2 * time.Second)
	openInBrowser(appAddress)

	systray.Run(onTrayReady, func() {})
}
