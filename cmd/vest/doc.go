package main

import (
	"bytes"
	"fmt"
	"go/doc"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

var (
	version      = "dev"
	commit, date string
)

func appVersion() string {
	return fmt.Sprintf(
		`%s (%s on %s/%s; %s)`,
		strings.Join([]string{version, commit, date}, "/"),
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
		runtime.Compiler,
	)
}

func usage() string {
	funcs := template.FuncMap{
		"Wrap": func(indent int, s string) string {
			buf := bytes.NewBuffer(nil)
			indentText := strings.Repeat(" ", indent)
			doc.ToText(buf, s, indentText, "  "+indentText, 80-indent)
			return buf.String()
		},
	}
	t := template.Must(template.New("usage").Funcs(funcs).Parse(`
Usage: {{ .Self }} user-spec command [args]
   eg: {{ .Self }} myuser bash
       {{ .Self }} nobody:root bash -c 'whoami && id'
       {{ .Self }} 1000:1 id
{{- if .EnvVars }}

  Environment Variables:
  {{ range .EnvVars }}
  {{- range $envVar, $description := . }}
    {{ $envVar }}
{{ $description | Wrap 6 }}
  {{- end }}
  {{- end }}
{{- end }}
{{ .Self }} version: {{ .Version }}
{{ .Self }} license: GPL-3 (full text at https://github.com/lumoslabs/vestibule)
`))
	var b bytes.Buffer
	template.Must(t, t.Execute(&b, struct {
		Self    string
		Version string
		EnvVars []map[string]string
	}{
		Self:    filepath.Base(os.Args[0]),
		Version: appVersion(),
		EnvVars: secretProviderEnvVars,
	}))
	return strings.TrimSpace(b.String()) + "\n"
}
