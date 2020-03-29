// +build unit

package sidecar_test

import (
	"encoding/json"
	//"fmt"
	"strings"
	"testing"

	//jsonpatch "github.com/evanphx/json-patch"
	//"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
				require.Equal(t, err.Error(), tc.result)
				return
			}

			container.Patch()

			jsonPatch, err := json.Marshal(container.Patches[0])
			require.Nil(t, err)
			require.JSONEq(t, tc.result.(string), string(jsonPatch))
		})
	}
}

func TestCreateContainerEnvVar(t *testing.T) {
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
			[]corev1.EnvVar{
				{
					Name:  "ENVNAME",
					Value: "envname",
				},
				{
					Name:  "ENV_NAME",
					Value: "env_name",
				},
				{
					Name:  "ENV-NAME",
					Value: "env-name",
				},
				{
					Name:  "ENV_NAME_1",
					Value: "env_name_1",
				},
			},
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
				require.Equal(t, err.Error(), tc.result)
				return
			}

			container.Patch()

			jsonContainer, err := json.Marshal(container.Patches[0].Value.([]corev1.Container)[0])
			require.Nil(t, err)

			result := corev1.Container{}
			err = json.Unmarshal(jsonContainer, &result)
			require.Nil(t, err)

			require.ElementsMatch(t, tc.result, result.Env)
		})
	}
}

func TestCreateContainerVolumeMounts(t *testing.T) {
	testCases := []struct {
		name        string
		annotations map[string]string
		result      interface{}
	}{
		{
			"OKContainerVolumeMounts",
			map[string]string{
				"container-injector.uthng.me/inject":                  "true",
				"container-injector.uthng.me/name":                    "sleep",
				"container-injector.uthng.me/image":                   "governmentpaas/curl-ssl",
				"container-injector.uthng.me/volume-mount-gitconfig":  "/opt/gitconfig",
				"container-injector.uthng.me/volume-mount-git-config": "/opt/git-config",
				"container-injector.uthng.me/volume-mount-gitconfigjson": `
{
	"mountPath": "/opt/gitConfig",
	"readOnly": true
}`,
			},
			[]corev1.VolumeMount{
				{
					Name:      "gitconfig",
					MountPath: "/opt/gitconfig",
				},
				{
					Name:      "git-config",
					MountPath: "/opt/git-config",
				},
				{
					Name:      "gitconfigjson",
					MountPath: "/opt/gitConfig",
					ReadOnly:  true,
				},
			},
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
				require.Equal(t, err.Error(), tc.result)
				return
			}

			container.Patch()

			jsonContainer, err := json.Marshal(container.Patches[0].Value.([]corev1.Container)[0])
			require.Nil(t, err)

			result := corev1.Container{}
			err = json.Unmarshal(jsonContainer, &result)
			require.Nil(t, err)

			require.ElementsMatch(t, tc.result, result.VolumeMounts)
		})
	}
}
