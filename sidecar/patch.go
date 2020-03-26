package sidecar

import (
	"strings"

	//"github.com/mattbaird/jsonpatch"
	corev1 "k8s.io/api/core/v1"
)

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func addVolumes(target, volumes []corev1.Volume, base string) []patchOperation {
	var result []patchOperation
	var value interface{}

	first := len(target) == 0

	for _, v := range volumes {
		path := base
		value = v

		if first {
			first = false
			value = []corev1.Volume{v}
		} else {
			path = path + "/-"
		}

		result = append(result, patchOperation{
			Op:    "add",
			Path:  path,
			Value: value,
		})
	}

	return result
}

func addVolumeMounts(target, mounts []corev1.VolumeMount, base string) []patchOperation {
	var result []patchOperation
	var value interface{}

	first := len(target) == 0

	for _, v := range mounts {
		value = v
		path := base

		if first {
			first = false
			value = []corev1.VolumeMount{v}
		} else {
			path = path + "/-"
		}

		result = append(result, patchOperation{
			Op:    "add",
			Path:  path,
			Value: value,
		})
	}

	return result
}

func removeContainers(path string) []patchOperation {
	var result []patchOperation

	return append(result, patchOperation{
		Op:   "remove",
		Path: path,
	})
}

func addContainers(target, containers []corev1.Container, base string) []patchOperation {
	var result []patchOperation
	var value interface{}

	first := len(target) == 0

	for _, container := range containers {
		value = container
		path := base

		if first {
			first = false
			value = []corev1.Container{container}
		} else {
			path = path + "/-"
		}

		result = append(result, patchOperation{
			Op:    "add",
			Path:  path,
			Value: value,
		})
	}

	return result
}

func updateAnnotations(target, annotations map[string]string) []patchOperation {
	var result []patchOperation

	if len(target) == 0 {
		result = append(result, patchOperation{
			Op:    "add",
			Path:  "/metadata/annotations",
			Value: annotations,
		})

		return result
	}

	for key, value := range annotations {
		result = append(result, patchOperation{
			Op:    "add",
			Path:  "/metadata/annotations/" + EscapeJSONPointer(key),
			Value: value,
		})
	}

	return result
}

// EscapeJSONPointer escapes a JSON string to be compliant with the
// JavaScript Object Notation (JSON) Pointer syntax RFC:
// https://tools.ietf.org/html/rfc6901.
func EscapeJSONPointer(s string) string {
	s = strings.Replace(s, "~", "~0", -1)
	s = strings.Replace(s, "/", "~1", -1)

	return s
}
