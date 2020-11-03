package main

import (
	"context"
	"encoding/json"
	"github.com/k8s-autoops/autoops"
	admissionv1 "k8s.io/api/admission/v1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"net/http"
	"os"
	"strconv"
)

const (
	AnnotationKeyEnabled = "autoops.enforce-qcloud-fixed-ip"
)

func exit(err *error) {
	if *err != nil {
		log.Println("exited with error:", (*err).Error())
		os.Exit(1)
	} else {
		log.Println("exited")
	}
}

func main() {
	var err error
	defer exit(&err)

	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	var client *kubernetes.Clientset
	if client, err = autoops.InClusterClient(); err != nil {
		return
	}

	s := &http.Server{
		Addr: ":443",
		Handler: autoops.NewMutatingAdmissionHTTPHandler(
			func(ctx context.Context, request *admissionv1.AdmissionRequest, patches *[]map[string]interface{}) (err error) {
				var ns *corev1.Namespace
				if ns, err = client.CoreV1().Namespaces().Get(ctx, request.Namespace, metav1.GetOptions{}); err != nil {
					return
				}
				if ns.Annotations == nil {
					return
				}
				if ok, _ := strconv.ParseBool(ns.Annotations[AnnotationKeyEnabled]); !ok {
					return
				}

				var buf []byte
				if buf, err = request.Object.MarshalJSON(); err != nil {
					return
				}
				var sts v1.StatefulSet
				if err = json.Unmarshal(buf, &sts); err != nil {
					return
				}

				if sts.Spec.Template.Annotations == nil {
					*patches = append(*patches, map[string]interface{}{
						"op":    "replace",
						"path":  "/spec/template/metadata/annotations",
						"value": map[string]interface{}{},
					})
				}
				*patches = append(*patches, map[string]interface{}{
					"op":    "replace",
					"path":  "/spec/template/metadata/annotations/tke.cloud.tencent.com~1vpc-ip-claim-delete-policy",
					"value": "Never",
				})
				return
			}),
	}

	if err = autoops.RunAdmissionServer(s); err != nil {
		return
	}
}
