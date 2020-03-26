package http

import (
	"context"
	"fmt"
	//"io"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"k8s.io/api/admission/v1beta1"
	//admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	log "github.com/uthng/golog"
	utils "github.com/uthng/goutils"

	"github.com/uthng/container-injector/sidecar"
)

var deserializer = func() runtime.Decoder {
	codecs := serializer.NewCodecFactory(runtime.NewScheme())
	return codecs.UniversalDeserializer()
}

var ignoredNamespaces = []string{
	metav1.NamespaceSystem,
	metav1.NamespacePublic,
}

// Server describes a list of functions for this interface
type Server interface {
	Serve() error
}

// Server describes server's elements
type server struct {
	logger *log.Logger

	addr     string
	certFile string
	keyFile  string
}

// NewServer returns a new interface
func NewServer(ctx context.Context, addr, certFile, keyFile string, logger *log.Logger) Server {
	s := &server{
		logger:   logger,
		addr:     addr,
		certFile: certFile,
		keyFile:  keyFile,
	}

	return s
}

// Serv launches http server
func (s *server) Serve() error {
	r := mux.NewRouter()

	r.HandleFunc("/mutate", s.handleMutate).Methods("POST")

	http.Handle("/", accessControl(r))

	return http.ListenAndServeTLS(s.addr, s.certFile, s.keyFile, nil)
}

///////////// INTERNAL FUNCTIONS /////////////////

func (s *server) handleMutate(w http.ResponseWriter, r *http.Request) {
	var body []byte
	var err error
	var admReq v1beta1.AdmissionReview
	var admResp v1beta1.AdmissionReview

	// Check content-type which must be application/json
	if ct := r.Header.Get("Content-Type"); ct != "application/json" {
		s.logger.Errorw("Invalid content-type", "content-type", ct)

		msg := fmt.Sprintf("invalid content-type: %q", ct)
		http.Error(w, msg, http.StatusBadRequest)

		return
	}

	if r.Body != nil {
		if body, err = ioutil.ReadAll(r.Body); err != nil {
			s.logger.Errorw("Error to read request", "err", err)

			msg := fmt.Sprintf("Error reading request body: %s", err)
			http.Error(w, msg, http.StatusBadRequest)

			return
		}
	}

	if len(body) == 0 {
		msg := "Empty request body"
		s.logger.Errorw(msg)
		http.Error(w, msg, http.StatusBadRequest)

		return
	}

	if _, _, err := deserializer().Decode(body, nil, &admReq); err != nil {
		s.logger.Errorw("Error to decode adminssion request", "err", err)

		msg := fmt.Sprintf("Error decoding admission request: %s", err)
		http.Error(w, msg, http.StatusInternalServerError)

		return
	}

	admResp.Response = s.mutate(admReq.Request)

	resp, err := json.Marshal(&admResp)
	if err != nil {
		s.logger.Errorw("Error to marshal admission response", "err", err)

		msg := fmt.Sprintf("error marshalling admission response: %s", err)
		http.Error(w, msg, http.StatusInternalServerError)

		return
	}

	if _, err := w.Write(resp); err != nil {
		s.logger.Errorw("Error while writing response", "err", err)
	}
}

// mutate takes an admission request and performs mutation if necessary,
// returning the final API response.
func (s *server) mutate(req *v1beta1.AdmissionRequest) *v1beta1.AdmissionResponse {
	// Decode the pod from the request
	var pod corev1.Pod
	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		s.logger.Errorw("Could not unmarshal request to pod", "err", err)
		s.logger.Debugf("Request Object Raw: %s", req.Object.Raw)

		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	// Build the basic response
	resp := &v1beta1.AdmissionResponse{
		Allowed: true,
		UID:     req.UID,
	}

	s.logger.Debugw("Checking if a container should be inject...")

	inject, err := needInject(&pod)
	if err != nil {
		return admissionError(fmt.Errorf("error checking if a container should be injected: %s", err))
	} else if !inject {
		return resp
	}

	s.logger.Debugw("Checking namespaces..")

	if pos := utils.SliceFindElemStr(ignoredNamespaces, req.Namespace); pos < 0 {
		err := fmt.Errorf("error with request namespace: cannot inject into system namespaces: %s", req.Namespace)
		return admissionError(err)
	}

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

func admissionError(err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}
