package main

import (
	"fmt"
	"os"

	"github.com/lumoslabs/vestibule/pkg/environ"
	"github.com/lumoslabs/vestibule/pkg/environ/providers/dotenv"
	"github.com/lumoslabs/vestibule/pkg/environ/providers/ejson"
	"github.com/lumoslabs/vestibule/pkg/environ/providers/sops"
	"github.com/lumoslabs/vestibule/pkg/environ/providers/vault"
	logger "github.com/lumoslabs/vestibule/pkg/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	secretProviders = []string{
		dotenv.Name,
		ejson.Name,
		vault.Name,
		sops.Name,
	}

	app       = kingpin.New("bule", "Write secrets to a file! What could go wrong?").DefaultEnvars()
	debug     = app.Flag("debug", "Debug output").Short('D').Bool()
	verbose   = app.Flag("verbose", "Verbose output").Short('v').Bool()
	format    = app.Flag("format", fmt.Sprintf("Format of the output file. Available formats: %v", environ.Marshallers())).Short('F').Default("json").HintOptions(environ.Marshallers()...).Enum(environ.Marshallers()...)
	providers = app.Flag("provider", fmt.Sprintf("Secret provider. Can be used multiple times. Available providers: %v", secretProviders)).Short('p').Default("vault").Strings()
	upcase    = app.Flag("upcase-var-names", "Upcase environment variable names gathered from secret providers.").Default("true").Bool()
	filename  = app.Arg("file", "Path of output file").Required().String()
)

func main() {
	app.Author(author)
	app.Version(appVersion())
	app.HelpFlag.Short('h')
	kingpin.MustParse(app.Parse(os.Args[1:]))

	logLevel := "disabled"
	if *debug {
		logLevel = "debug"
	} else if *verbose {
		logLevel = "info"
	}

	log := newLogger(logLevel, os.Stderr)
	logger.SetLogger(log)

	secrets := environ.New()
	secrets.UpcaseKeys = *upcase
	secrets.Populate(*providers)
	secrets.SetMarshaller(*format)

	log.Debugf("Writing secrets to file. file=%s fmt=%s", *filename, *format)
	file, er := os.Create(*filename)
	defer file.Close()

	if er != nil {
		log.Infof("Failed to write secrets to file. file=%s err=%v", *filename, er)
		os.Exit(1)
	}

	if er := secrets.Write(file); er != nil {
		log.Infof("Failed to write secrets to file. file=%s err=%v", *filename, er)
	}
}
