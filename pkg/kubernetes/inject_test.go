package kubernetes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/ghodss/yaml"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

var assetsPath = filepath.Join("fixtures", "assets")

func TestReview(t *testing.T) {
	cases := []struct {
		inputFile string
	}{
		{"simple.yaml"},
	}

	wh := NewWebhookHandler(http.Dir(assetsPath), &WebhookHandlerConfig{AnnotationNamespace: "test"})

	for i, c := range cases {
		path := filepath.Join("fixtures", c.inputFile)

		t.Run(fmt.Sprintf("[%d] %s", i, c.inputFile), func(t *testing.T) {
			t.Parallel()
			podYaml, er := ioutil.ReadFile(path)
			if er != nil {
				t.Fatal(er.Error())
			}
			podJson, er := yaml.YAMLToJSON(podYaml)
			if er != nil {
				t.Fatal(er.Error())
			}

			data, _ := json.Marshal(&v1beta1.AdmissionReview{
				Request: &v1beta1.AdmissionRequest{
					Object: runtime.RawExtension{
						Raw: podJson,
					},
				},
			})
			review, er := wh.review(data)
			if er != nil {
				t.Fatal(er.Error())
			}

			if resp := review.Response; resp != nil {
				var prettyPatch bytes.Buffer
				if er := json.Indent(&prettyPatch, resp.Patch, "", " "); er != nil {
					t.Fatal(er.Error())
				}
				t.Log(prettyPatch.String())
			}
		})
	}
}
