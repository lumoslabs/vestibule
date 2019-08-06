package kubernetes

import (
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v2alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var ignoredNamespaces = []string{
	metav1.NamespaceSystem,
	metav1.NamespacePublic,
}

var (
	kinds = []struct {
		groupVersion schema.GroupVersion
		obj          runtime.Object
		resource     string
		apiPath      string
	}{
		{v1.SchemeGroupVersion, &v1.ReplicationController{}, "replicationcontrollers", "/api"},
		{v1.SchemeGroupVersion, &v1.Pod{}, "pods", "/api"},

		{appsv1.SchemeGroupVersion, &appsv1.Deployment{}, "deployments", "/apis"},
		{appsv1.SchemeGroupVersion, &appsv1.DaemonSet{}, "daemonsets", "/apis"},
		{appsv1.SchemeGroupVersion, &appsv1.ReplicaSet{}, "replicasets", "/apis"},

		{batchv1.SchemeGroupVersion, &batchv1.Job{}, "jobs", "/apis"},
		{v2alpha1.SchemeGroupVersion, &v2alpha1.CronJob{}, "cronjobs", "/apis"},

		{appsv1.SchemeGroupVersion, &appsv1.StatefulSet{}, "statefulsets", "/apis"},

		{v1.SchemeGroupVersion, &v1.List{}, "lists", "/apis"},
	}
	injectScheme = runtime.NewScheme()
)

func init() {
	for _, kind := range kinds {
		injectScheme.AddKnownTypes(kind.groupVersion, kind.obj)
		injectScheme.AddUnversionedTypes(kind.groupVersion, kind.obj)
	}
}
