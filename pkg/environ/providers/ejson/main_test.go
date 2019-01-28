package ejson

import (
	"os"
	"strings"
	"testing"

	"github.com/lumoslabs/vestibule/pkg/environ"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

var (
	fs   = afero.NewOsFs()
	keys = []string{
		"a04086f26d0a6b01a9ca7954b60c4de7517070da7940e698b9250e124042eb29;06f0b309fb3023422bd50f9ed3bb8bbafafb1254052b202cd139a01484ba95d2",
		"3fc11ee2b4d1228d7648765fcfa95c5476a758d42328db2f52c020033ad8342d;3f25827973d60dc3ebdd1f6e6cb5560b5d2e29f7b4de7abac648b69a334d9e48",
	}
)

func TestIsProvider(t *testing.T) {
	assert.Implements(t, (*environ.Provider)(nil), new(EjsonProvider))
}

func TestAddToEnviron(t *testing.T) {

	tests := []struct {
		name       string
		errorFunc  func(assert.TestingT, error, string, ...interface{}) bool
		keyvals    map[string]string
		keys, docs []string
	}{
		{
			"good-document",
			assert.NoErrorf,
			map[string]string{"TEST_KEY": "first test"},
			keys,
			[]string{`{"_public_key": "a04086f26d0a6b01a9ca7954b60c4de7517070da7940e698b9250e124042eb29","TEST_KEY": "EJ[1:CWMhGji3q8i0vGCGnLI4jHScp2lXA/VjETOtNBEsXB4=:CFpXDOdnEhsVXvd5tabbUcDlilzpSgc8:IBq+xoe33AnbCljM1cdY1y44ISW5VIIdE6s=]"}`},
		},
		{
			"bad-document",
			assert.Errorf,
			map[string]string{},
			keys,
			[]string{`{"_public_key": "a04086f26d0a6b01a9ca7954b60c4de7517070da7940e698b9250e124042eb29","TEST_KEY": "EJ[1:IoG1Nh938QR1RZaSYSuxwyLxNhmIyGqP/HK5N]"}`},
		},
		{
			"good-documents",
			assert.NoErrorf,
			map[string]string{"TEST_KEY_1": "val1", "TEST_KEY_2": "val2", "TEST_KEY_3": "val3"},
			keys,
			[]string{
				`{"_public_key": "a04086f26d0a6b01a9ca7954b60c4de7517070da7940e698b9250e124042eb29","TEST_KEY_1": "EJ[1:elxnnyi+twYPdZLA0mR0cNWDPeIOCZ4M3EBXLW6LhyY=:SKDM4mPALFaqnBgPfzhxZV0pw955aAOL:CI2A0NpC6LBCnYCLEdOFJklMEmA=]"}`,
				`{"_public_key": "3fc11ee2b4d1228d7648765fcfa95c5476a758d42328db2f52c020033ad8342d","TEST_KEY_2": "EJ[1:5u+kui5EVe0FBdHu4OJvYMC2+xYHXI1N3CvyKSEFjFk=:f2JWU6kB/Re9UAYCpkPxDRnsFSIGGx5G:hjzFnGG0Q7Zan3Zn4qLxKCt3Dak=]","_TEST_KEY_3": "val3"}`,
			},
		},
	}

	for _, tt := range tests {
		e := environ.NewEnviron()

		files := make([]string, 0, len(tt.docs))
		for _, data := range tt.docs {
			f, _ := afero.TempFile(fs, "", "")
			f.WriteString(data)
			files = append(files, f.Name())
			f.Close()
		}

		os.Setenv("EJSON_KEYS", strings.Join(keys, KeyPairEnvSeparator))
		os.Setenv("EJSON_FILES", strings.Join(files, ":"))

		ej, er := NewEjsonProvider()
		assert.NoErrorf(t, er, tt.name)
		tt.errorFunc(t, ej.AddToEnviron(e), tt.name)

		ea := environ.NewEnviron()
		ea.Append(tt.keyvals)
		assert.Equalf(t, e.String(), ea.String(), tt.name)

		for _, f := range files {
			fs.Remove(f)
		}
	}
}
