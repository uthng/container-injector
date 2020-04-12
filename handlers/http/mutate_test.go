package http_test

import (
	"bytes"
	"encoding/base64"
	//"encoding/json"
	//"fmt"
	"net/http"
	"net/http/httptest"
	//"strings"
	"io/ioutil"
	"testing"

	"github.com/json-iterator/go"
	"github.com/stretchr/testify/require"

	"k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	log "github.com/uthng/golog"

	httphandler "github.com/uthng/container-injector/handlers/http"
	"github.com/uthng/container-injector/sidecar"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// encodeRaw is a helper to encode some data into a RawExtension.
func encodeRaw(t *testing.T, input interface{}) runtime.RawExtension {
	data, err := json.Marshal(input)
	require.NoError(t, err)

	return runtime.RawExtension{Raw: data}
}

func TestHandlerMutateChecks(t *testing.T) {
	basicSpec := corev1.PodSpec{
		InitContainers: []corev1.Container{
			{
				Name: "web-init",
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "foobar",
						MountPath: "serviceaccount/somewhere",
					},
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name: "web",
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "foobar",
						MountPath: "serviceaccount/somewhere",
					},
				},
			},
		},
	}

	basicTypeMeta := metav1.TypeMeta{
		Kind:       "AdmissionReview",
		APIVersion: "v1",
	}

	testCases := []struct {
		name   string
		header map[string]string
		body   interface{}
		result interface{}
	}{
		{
			"ErrContentType",
			map[string]string{
				"Content-Type": "text/plain",
			},
			nil,
			httptest.ResponseRecorder{
				Code: http.StatusBadRequest,
				Body: bytes.NewBuffer([]byte("invalid content-type: text/plain\n")),
			},
		},
		{
			"ErrBodyNil",
			map[string]string{
				"Content-Type": "application/json",
			},
			nil,
			httptest.ResponseRecorder{
				Code: http.StatusBadRequest,
				Body: bytes.NewBuffer([]byte("Empty request body\n")),
			},
		},
		{
			"ErrDecodeAdmReviewReq",
			map[string]string{
				"Content-Type": "application/json",
			},
			v1.AdmissionRequest{
				Namespace: metav1.NamespaceSystem,
				Object: encodeRaw(t, &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							sidecar.AnnotationContainerInject: "true",
						},
					},
					Spec: basicSpec,
				}),
			},
			httptest.ResponseRecorder{
				Code: http.StatusInternalServerError,
				Body: bytes.NewBuffer([]byte("Error decoding admission request: couldn't get version/kind; json parse error: json: cannot unmarshal object into Go struct field .kind of type string\n")),
			},
		},
		{
			"ErrDecodePod",
			map[string]string{
				"Content-Type": "application/json",
			},
			v1.AdmissionReview{
				TypeMeta: basicTypeMeta,
				Request: &v1.AdmissionRequest{
					Namespace: metav1.NamespaceSystem,
					Object:    encodeRaw(t, nil),
				},
			},
			httptest.ResponseRecorder{
				Code: http.StatusOK,
				Body: bytes.NewBuffer([]byte(`{"response":{"uid":"","allowed":false,"status":{"metadata":{},"message":"unexpected end of JSON input"}}}`)),
			},
		},
		{
			"ErrNamespace",
			map[string]string{
				"Content-Type": "application/json",
			},
			v1.AdmissionReview{
				TypeMeta: basicTypeMeta,
				Request: &v1.AdmissionRequest{
					Namespace: metav1.NamespaceSystem,
					Object: encodeRaw(t, &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								sidecar.AnnotationContainerInject: "true",
							},
						},
						Spec: basicSpec,
					}),
				},
			},
			httptest.ResponseRecorder{
				Code: http.StatusOK,
				Body: bytes.NewBuffer([]byte(`{"response":{"uid":"","allowed":false,"status":{"metadata":{},"message":"error with request namespace: cannot inject into system namespaces: kube-system"}}}`)),
			},
		},
		{
			"ErrAnnoInjectValue",
			map[string]string{
				"Content-Type": "application/json",
			},
			v1.AdmissionReview{
				TypeMeta: basicTypeMeta,
				Request: &v1.AdmissionRequest{
					Namespace: metav1.NamespaceSystem,
					Object: encodeRaw(t, &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								sidecar.AnnotationContainerInject: "hello",
							},
						},
						Spec: basicSpec,
					}),
				},
			},
			httptest.ResponseRecorder{
				Code: http.StatusOK,
				Body: bytes.NewBuffer([]byte(`{"response":{"uid":"","allowed":false,"status":{"metadata":{},"message":"error checking if a container should be injected: strconv.ParseBool: parsing \"hello\": invalid syntax"}}}`)),
			},
		},
	}

	// Set logger
	httpLogger := log.NewLogger()

	//myserver := myhttp.NewServer("", "", "", httpLogger)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var body []byte
			var err error

			if tc.body != nil {
				body, err = json.Marshal(tc.body)
				require.Nil(t, err)
			}

			req, err := http.NewRequest("POST", "/", bytes.NewBuffer(body))
			require.Nil(t, err)

			req.Header.Set("Content-Type", tc.header["Content-Type"])

			rec := httptest.NewRecorder()

			handlerMutate := httphandler.NewMutate(httpLogger)
			handlerMutate.ServeHTTP(rec, req)

			result := httptest.ResponseRecorder{
				Code: rec.Code,
				Body: rec.Body,
			}

			require.Equal(t, tc.result.(httptest.ResponseRecorder), result)
		})
	}
}

func TestHandlerMutateInjection(t *testing.T) {
	basicSpec := corev1.PodSpec{
		InitContainers: []corev1.Container{
			{
				Name: "web-init",
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "foobar",
						MountPath: "serviceaccount/somewhere",
					},
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name: "web",
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "foobar",
						MountPath: "serviceaccount/somewhere",
					},
				},
			},
		},
	}

	basicTypeMeta := metav1.TypeMeta{
		Kind:       "AdmissionReview",
		APIVersion: "v1",
	}

	testCases := []struct {
		name   string
		header map[string]string
		body   interface{}
		result interface{}
	}{
		{
			"OKContainerSimple",
			map[string]string{
				"Content-Type": "application/json",
			},
			v1.AdmissionReview{
				TypeMeta: basicTypeMeta,
				Request: &v1.AdmissionRequest{
					Namespace: "container-injector",
					Object: encodeRaw(t, &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								sidecar.AnnotationContainerInject: "true",
								sidecar.AnnotationContainerName:   "curl-ssl",
								sidecar.AnnotationContainerImage:  "govermentpaas/curl-ssl",
							},
						},
						Spec: basicSpec,
					}),
				},
			},
			`[{"op":"add","path":"/spec/containers/-","value":{"name":"curl-ssl","image":"govermentpaas/curl-ssl","resources":{}}}]`,
		},
		{
			"OKContainerFullOpts",
			map[string]string{
				"Content-Type": "application/json",
			},
			v1.AdmissionReview{
				TypeMeta: basicTypeMeta,
				Request: &v1.AdmissionRequest{
					Namespace: "container-injector",
					Object: encodeRaw(t, &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								sidecar.AnnotationContainerInject:                     "true",
								sidecar.AnnotationContainerName:                       "curl-ssl",
								sidecar.AnnotationContainerImage:                      "govermentpaas/curl-ssl",
								sidecar.AnnotationContainerCommand:                    "/bin/sh -ec echo 'hello world'",
								sidecar.AnnotationContainerEnv + "-ENVNAME":           "envname",
								sidecar.AnnotationContainerVolumeMount + "-gitconfig": "/opt/gitconfig",
								sidecar.AnnotationContainerVolumeSource + "-volsecret": `
{
	"secret": {
		"secretName": "volsecret"
	}
}`,
							},
						},
						Spec: basicSpec,
					}),
				},
			},
			`[{"op":"add","path":"/spec/volumes","value":[{"name":"volsecret","secret":{"secretName":"volsecret"}}]},{"op":"add","path":"/spec/containers/-","value":{"name":"curl-ssl","image":"govermentpaas/curl-ssl","command":["/bin/sh -ec echo 'hello world'"],"env":[{"name":"ENVNAME","value":"envname"}],"resources":{},"volumeMounts":[{"name":"gitconfig","mountPath":"/opt/gitconfig"}]}}]`,
		},
	}

	// Set logger
	httpLogger := log.NewLogger()

	//myserver := myhttp.NewServer("", "", "", httpLogger)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var body []byte
			var err error

			if tc.body != nil {
				body, err = json.Marshal(tc.body)
				require.Nil(t, err)
			}

			req, err := http.NewRequest("POST", "/", bytes.NewBuffer(body))
			require.Nil(t, err)

			req.Header.Set("Content-Type", tc.header["Content-Type"])

			rec := httptest.NewRecorder()

			handlerMutate := httphandler.NewMutate(httpLogger)
			handlerMutate.ServeHTTP(rec, req)

			bodyData, err := ioutil.ReadAll(rec.Body)
			require.Nil(t, err)

			patch, err := base64.StdEncoding.DecodeString(json.Get(bodyData, "response", "patch").ToString())
			require.Nil(t, err)

			require.Equal(t, tc.result.(string), string(patch))
		})
	}
}
