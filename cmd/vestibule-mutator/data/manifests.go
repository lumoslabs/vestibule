// +build build

package data

import (
	"net/http"

	"github.com/spf13/afero"
)

var (
	fs = afero.NewOsFs()
)

func newAssetsFs(path string) http.FileSystem {
	return afero.NewHttpFs(fs).Dir(path)
}
