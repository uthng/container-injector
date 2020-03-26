// +build unit

package sidecar_test

import (
	"encoding/json"
	//"fmt"
	"strings"
	"testing"

	//jsonpatch "github.com/evanphx/json-patch"
	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/uthng/container-injector/sidecar"
)

func TestNewContainer(t *testing.T) {
	testCases := []struct {
		name        string
		annotations map[string]string
		result      interface{}
	}{
		{
			"ErrNewContainerAnnoInject",
			map[string]string{
				"container-injector.uthng.me/status": "injected",
			},
			"Annotation 'container-injector.uthng.me/inject' not found",
		},
		{
			"ErrNewContainerAnnoName",
			map[string]string{
				"container-injector.uthng.me/inject": "true",
			},
			"Annotation 'container-injector.uthng.me/name' not found",
		},
		{
			"ErrNewContainerAnnoImage",
			map[string]string{
				"container-injector.uthng.me/inject": "true",
				"container-injector.uthng.me/name":   "sleep",
			},
			"Annotation 'container-injector.uthng.me/image' not found",
		},
		{
			"OKContainerSimple",
			map[string]string{
				"container-injector.uthng.me/inject": "true",
				"container-injector.uthng.me/name":   "sleep",
				"container-injector.uthng.me/image":  "governmentpaas/curl-ssl",
			},
			`
{
	"op": "add",
	"path": "/spec/containers",
	"value": [
		{
			"name": "sleep",
			"image": "governmentpaas/curl-ssl",
			"resources": {}
		}
	]
}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tc.annotations,
				},
			}

			container, err := sidecar.NewContainer(pod)
			if strings.HasPrefix(tc.name, "Err") {
				assert.Equal(t, err.Error(), tc.result)
				return
			}

			container.Patch()

			assert.Nil(t, err)

			//jsonContainers, err := json.Marshal(pod.Spec.Containers)
			//assert.Nil(t, err)

			//fmt.Printf("%+v\n", string(jsonContainers))

			jsonPatch, err := json.Marshal(container.Patches[0])
			assert.Nil(t, err)
			//fmt.Printf("%+v\n", string(jsonPatch))

			//patch, err := jsonpatch.CreateMergePatch(jsonPatch, jsonContainers)
			//fmt.Printf("%+v\n", string(patch))
			assert.JSONEq(t, tc.result.(string), string(jsonPatch))
		})
	}
}

func TestCreateContainer(t *testing.T) {
	testCases := []struct {
		name        string
		annotations map[string]string
		result      interface{}
	}{
		{
			"OKContainerEnvVars",
			map[string]string{
				"container-injector.uthng.me/inject":         "true",
				"container-injector.uthng.me/name":           "sleep",
				"container-injector.uthng.me/image":          "governmentpaas/curl-ssl",
				"container-injector.uthng.me/env-ENVNAME":    "envname",
				"container-injector.uthng.me/env-ENV_NAME":   "env_name",
				"container-injector.uthng.me/env-ENV-NAME":   "env-name",
				"container-injector.uthng.me/env-ENV_NAME_1": "env_name_1",
			},
			`
{
	"op": "add",
	"path": "/spec/containers",
	"value": [
		{
			"name": "sleep",
			"image": "governmentpaas/curl-ssl",
			"env": [
				{
					"name": "ENVNAME",
					"value": "envname"
				},
				{
					"name": "ENV_NAME",
					"value": "env_name"
				},
				{
					"name": "ENV-NAME",
					"value": "env-name"
				},
				{
					"name": "ENV_NAME_1",
					"value": "env_name_1"
				}
			],
			"resources": {}
		}
	]
}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tc.annotations,
				},
			}

			container, err := sidecar.NewContainer(pod)
			if strings.HasPrefix(tc.name, "Err") {
				assert.Equal(t, err.Error(), tc.result)
				return
			}

			container.Patch()

			assert.Nil(t, err)

			//jsonContainers, err := json.Marshal(pod.Spec.Containers)
			//assert.Nil(t, err)

			//fmt.Printf("%+v\n", string(jsonContainers))

			jsonPatch, err := json.Marshal(container.Patches[0])
			assert.Nil(t, err)
			//fmt.Printf("%+v\n", string(jsonPatch))

			//patch, err := jsonpatch.CreateMergePatch(jsonPatch, jsonContainers)
			//fmt.Printf("%+v\n", string(patch))
			assert.JSONEq(t, tc.result.(string), string(jsonPatch))
		})
	}
}
