package http

import (
	"fmt"
	//"io"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"k8s.io/api/admission/v1"
	//admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	log "github.com/uthng/golog"
	utils "github.com/uthng/goutils"

	"github.com/uthng/container-injector/sidecar"
)

// Mutate represents a struct for http.Handler
type Mutate struct {
	logger *log.Logger
}

var deserializer = func() runtime.Decoder {
	codecs := serializer.NewCodecFactory(runtime.NewScheme())
	return codecs.UniversalDeserializer()
}

var ignoredNamespaces = []string{
	metav1.NamespaceSystem,
	metav1.NamespacePublic,
}

// NewMutate return new mutate instance implementing http.Handler
func NewMutate(l *log.Logger) *Mutate {
	return &Mutate{
		logger: l,
	}
}

// ServeHTTP implements http.Handler
func (m *Mutate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body []byte
	var err error
	var admReviewReq v1.AdmissionReview
	var admReviewResp v1.AdmissionReview

	// Check content-type which must be application/json
	if ct := r.Header.Get("Content-Type"); ct != "application/json" {
		m.logger.Errorw("Invalid content-type", "content-type", ct)

		msg := fmt.Sprintf("invalid content-type: %s", ct)
		http.Error(w, msg, http.StatusBadRequest)

		return
	}

	if r.Body != nil {
		if body, err = ioutil.ReadAll(r.Body); err != nil {
			m.logger.Errorw("Error to read request", "err", err)

			msg := fmt.Sprintf("Error reading request body: %s", err)
			http.Error(w, msg, http.StatusBadRequest)

			return
		}
	}

	if len(body) == 0 {
		msg := "Empty request body"
		m.logger.Errorw(msg)
		http.Error(w, msg, http.StatusBadRequest)

		return
	}

	if _, _, err := deserializer().Decode(body, nil, &admReviewReq); err != nil {
		m.logger.Errorw("Error to decode adminssion request", "err", err)

		msg := fmt.Sprintf("Error decoding admission request: %s", err)
		http.Error(w, msg, http.StatusInternalServerError)

		return
	}

	admReviewResp.Response = m.mutate(admReviewReq.Request)

	resp, err := json.Marshal(&admReviewResp)
	if err != nil {
		m.logger.Errorw("Error to marshal admission response", "err", err)

		msg := fmt.Sprintf("error marshalling admission response: %s", err)
		http.Error(w, msg, http.StatusInternalServerError)

		return
	}

	if _, err := w.Write(resp); err != nil {
		m.logger.Errorw("Error while writing response", "err", err)
	}
}

// mutate takes an admission request and performs mutation if necessary,
// returning the final API response.
func (m *Mutate) mutate(req *v1.AdmissionRequest) *v1.AdmissionResponse {
	// Decode the pod from the request
	var pod corev1.Pod
	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		m.logger.Errorw("Could not unmarshal request to pod", "err", err)
		m.logger.Debugf("Request Object Raw: %s", req.Object.Raw)

		return &v1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	// Build the basic response
	resp := &v1.AdmissionResponse{
		Allowed: true,
		UID:     req.UID,
	}

	m.logger.Debugw("Checking if a container should be inject...")

	inject, err := needInject(&pod)
	if err != nil {
		return admissionError(fmt.Errorf("error checking if a container should be injected: %s", err))
	} else if !inject {
		return resp
	}

	m.logger.Debugw("Checking namespaces...")

	if pos := utils.SliceFindElemStr(ignoredNamespaces, req.Namespace); pos >= 0 {
		err := fmt.Errorf("error with request namespace: cannot inject into system namespaces: %s", req.Namespace)
		m.logger.Errorw("Error request namespace", "namespace", req.Namespace)

		return admissionError(err)
	}

	m.logger.Infow("Init container to be injected...")

	container, err := sidecar.NewContainer(&pod)
	if err != nil {
		m.logger.Errorw("Error to initialize container to be injected", "err", err)
		return admissionError(err)
	}

	m.logger.Infow("Creating patches for Pod...")

	patch, err := container.Patch()
	if err != nil {
		m.logger.Errorw("Error to create patches for Pod", "err", err)
		return admissionError(err)
	}

	resp.Patch = patch
	patchType := v1.PatchTypeJSONPatch
	resp.PatchType = &patchType

	return resp
}

func needInject(pod *corev1.Pod) (bool, error) {
	raw, ok := pod.Annotations[sidecar.AnnotationContainerInject]
	if !ok {
		return false, nil
	}

	inject, err := strconv.ParseBool(raw)
	if err != nil {
		return false, err
	}

	if !inject {
		return false, nil
	}

	// This shouldn't happen so bail.
	raw, ok = pod.Annotations[sidecar.AnnotationContainerStatus]
	if !ok {
		return true, nil
	}

	// "injected" is the only status we care about. Don't do
	// anything if it's set.  The user can update the status
	// to force a new mutation.
	if raw == "injected" {
		return false, nil
	}

	return true, nil
}

func admissionError(err error) *v1.AdmissionResponse {
	return &v1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}
