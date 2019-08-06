package kubernetes

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/appscode/jsonpatch"
	"github.com/imdario/mergo"
	"github.com/lumoslabs/vestibule/pkg/log"
	"k8s.io/api/admission/v1beta1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

const (
	WebhookPath       = "/mutate"
	HeaderContentType = "Content-Type"
	JsonContentType   = "application/json"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()

	ErrWrongKind = errors.New("invalid request object kind")
)

func init() {
	_ = corev1.AddToScheme(runtimeScheme)
	_ = admissionregistrationv1beta1.AddToScheme(runtimeScheme)
}

func (wh *WebhookHandler) err(w http.ResponseWriter, status int, msgf string, m ...interface{}) {
	msg := fmt.Sprintf(msgf, m...)
	log.Info(msg)
	w.Header().Set(HeaderContentType, JsonContentType)
	http.Error(
		w,
		fmt.Sprintf(`{"status":"error","msg":"%s"}`, msg),
		status,
	)
}

func (wh *WebhookHandler) review(data []byte) (v1beta1.AdmissionReview, error) {
	review := v1beta1.AdmissionReview{}

	obj, _, er := deserializer.Decode(data, nil, &v1beta1.AdmissionReview{})
	if er != nil || obj == nil {
		review.Response = &v1beta1.AdmissionResponse{
			Result: &metav1.Status{Message: er.Error()},
		}
		return review, er
	}

	ar, ok := obj.(*v1beta1.AdmissionReview)
	if !ok {
		review.Response = &v1beta1.AdmissionResponse{
			Result: &metav1.Status{Message: ErrWrongKind.Error()},
		}
		return review, ErrWrongKind
	}

	request := ar.Request
	var pod corev1.Pod
	if er := json.Unmarshal(request.Object.Raw, &pod); er != nil {
		review.Response = &v1beta1.AdmissionResponse{
			Result: &metav1.Status{Message: er.Error()},
		}
		return review, er
	}

	log.Infof("reviewing admission request. kind=%v namespace=%v name=%v pod=%v", request.Kind, request.Namespace, request.Name, pod.Name)
	if skip(wh.Config.AnnotationNamespace, &pod.ObjectMeta) {
		log.Infof("skipping object.  namespace=%v pod=%v", pod.Namespace, pod.Name)
		review.Response = &v1beta1.AdmissionResponse{Allowed: true}
		return review, nil
	}

	var newPod corev1.Pod
	p, _ := wh.Assets.Open("pod.yaml")
	pData, _ := ioutil.ReadAll(p)
	_, _, er = deserializer.Decode(pData, nil, &newPod)
	if er != nil {
		return review, er
	}

	dstPod := pod.DeepCopy()
	ann := dstPod.GetAnnotations()
	delete(ann, wh.Config.AnnotationNamespace+"/inject")
	ann[wh.Config.AnnotationNamespace+"/status"] = "completed"
	dstPod.SetAnnotations(ann)

	mergo.Merge(&dstPod.Spec, newPod.Spec, mergo.WithOverride)
	dstJson, _ := json.Marshal(dstPod)
	podJson, _ := json.Marshal(pod)
	patch, _ := jsonpatch.CreatePatch(podJson, dstJson)
	patchJson, _ := json.Marshal(patch)

	review.Response = &v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchJson,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
	return review, nil
}

func (wh *WebhookHandler) HandleMutate(w http.ResponseWriter, r *http.Request) {
	if ct := r.Header.Get(HeaderContentType); ct != JsonContentType {
		wh.err(w, http.StatusUnsupportedMediaType, "invalid content-type '%s'", ct)
		return
	}

	data, er := ioutil.ReadAll(r.Body)
	if er != nil {
		wh.err(w, http.StatusBadRequest, "failed to read request body: %v", er)
		return
	}

	response, er := wh.review(data)
	if er != nil {
		log.Infof("error: %v", er)
	}
	resp, er := json.Marshal(response)
	if er != nil {
		wh.err(w, http.StatusInternalServerError, "failed to encode response: %v", er)
		return
	}

	if _, er := w.Write(resp); er != nil {
		wh.err(w, http.StatusInternalServerError, "failed to write response: %v", er)
	}
}

func skip(ns string, metadata *metav1.ObjectMeta) bool {
	switch metadata.Namespace {
	case metav1.NamespaceSystem, metav1.NamespacePublic:
		return true
	}

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	if status, ok := annotations[ns+"/status"]; ok && status == "completed" {
		return true
	}

	ok, _ := strconv.ParseBool(annotations[ns+"/inject"])
	return !ok
}
