package http_test

import (
	//"encoding/json"
	"bytes"
	//"fmt"
	"net/http"
	"net/http/httptest"
	//"strings"
	//"io"
	"testing"

	"github.com/stretchr/testify/require"

	//corev1 "k8s.io/api/core/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	log "github.com/uthng/golog"

	httphandler "github.com/uthng/container-injector/handlers/http"
)

func TestHandlerMutate(t *testing.T) {
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
	}

	// Set logger
	httpLogger := log.NewLogger()

	//myserver := myhttp.NewServer("", "", "", httpLogger)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/", nil)
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
