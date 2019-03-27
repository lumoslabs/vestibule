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
	debug     = app.Flag("debug", "Debug output").Bool()
	format    = app.Flag("format", fmt.Sprintf("Format of the output file. Available formats: %v", environ.Marshallers())).Short('F').Default("json").HintOptions(environ.Marshallers()...).Enum(environ.Marshallers()...)
	providers = app.Flag("provider", fmt.Sprintf("Secret provider. Can be used multiple times. Available providers: %v", secretProviders)).Short('p').Default("vault").Strings()
	filename  = app.Arg("file", "Path of output file").Required().String()
)

func main() {
	app.Author(author)
	app.Version(appVersion())
	app.HelpFlag.Short('h')
	kingpin.MustParse(app.Parse(os.Args[1:]))

	logLevel := ""
	if *debug {
		logLevel = "debug"
	}

	log := newLogger(logLevel, os.Stderr)
	logger.SetLogger(log)

	secrets := environ.New()
	secrets.Populate(*providers)

	log.Debugf("Writing secrets to file. file=%s fmt=%s", *filename, *format)
	if file, er := os.Create(*filename); er == nil {
		secrets.SetMarshaller(*format)
		secrets.Write(file)
		file.Close()
	} else {
		log.Infof("Failed to write secrets to file. file=%s msg=%s", *filename, er.Error())
	}
}
