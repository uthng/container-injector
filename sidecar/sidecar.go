package sidecar

import (
	//"errors"
	"fmt"
	"strings"
	//jsonpatch "github.com/evanphx/json-patch"
	"github.com/spf13/cast"

	corev1 "k8s.io/api/core/v1"
)

const (
	AnnotationPrefixEnv = "container-injector.uthng.me/env"
	AnnotationPrefixVol = "container-injector.uthng.me/vol"
)

// Container defines the container to be injected in the pod
type Container struct {
	// Pod is the original Kubernetes pod spec.
	Pod *corev1.Pod

	// Annotations are the current pod annotations used to
	// configure the Vault Agent container.
	Annotations map[string]string

	// Inject is the flag used to determine if a container should be requested
	// in a pod request.
	Inject bool

	// Status is the current injection status. The only status considered is "injected",
	// which prevents further mutations. A user can patch this annotation to force a new
	// mutation.
	Status string

	// ServiceAccountName is the Kubernetes service account name for the pod.
	// This is used when we mount the service account to the  Vault Agent container(s).
	ServiceAccountName string

	// ServiceAccountPath is the path on disk where the service account JWT
	// can be located.  This is used when we mount the service account to the
	// Vault Agent container(s).
	ServiceAccountPath string

	// Name is the name of the container to inject
	Name string

	// ImageName is the name of the image to use for the
	// sidecar container.
	ImageName string

	// Command is the command to launch when the image is started
	Command string

	// Args is the arguments of the command to be launched.
	Args string

	// InitContainer tells whether the connainer is injected as init container
	InitContainer bool

	// InitFirst tells whether the connainer is started before the others
	InitFirst bool

	// ImagePullPolicy is the pull policy
	ImagePullPolicy string

	// LimitsCPU is the upper CPU limit the sidecar container is allowed to consume.
	LimitsCPU string

	// LimitsMem is the upper memory limit the sidecar container is allowed to consume.
	LimitsMem string

	// RequestsCPU is the requested minimum CPU amount required  when being scheduled to deploy.
	RequestsCPU string

	// RequestsMem is the requested minimum memory amount required when being scheduled to deploy.
	RequestsMem string

	// ConfigMapName is the name of the configmap containing
	// container configuration
	ConfigMapName string

	// RunAsUser is the user ID to run the Vault agent container(s) as.
	RunAsUser int64

	// RunAsGroup is the group ID to run the Vault agent container(s) as.
	RunAsGroup int64

	// TLSSecret is the name of the Kubernetes secret containing
	// client TLS certificates and keys
	TLSSecret string

	// Patches are all the mutations we will make to the pod request.
	Patches []patchOperation
}

// NewContainer creates a new container by parsing all Kubernetes annotations
func NewContainer(pod *corev1.Pod) (*Container, error) {
	c := &Container{}

	c.Pod = pod

	if val, ok := pod.Annotations[AnnotationContainerInject]; ok {
		c.Inject = cast.ToBool(val)
	} else {
		return nil, newAnnotationError(AnnotationContainerInject)
	}

	if val, ok := pod.Annotations[AnnotationContainerName]; ok {
		c.Name = cast.ToString(val)
	} else {
		return nil, newAnnotationError(AnnotationContainerName)
	}

	if val, ok := pod.Annotations[AnnotationContainerImage]; ok {
		c.ImageName = cast.ToString(val)
	} else {
		return nil, newAnnotationError(AnnotationContainerImage)
	}

	if val, ok := pod.Annotations[AnnotationContainerCommand]; ok {
		c.Command = cast.ToString(val)
	}

	if val, ok := pod.Annotations[AnnotationContainerArgs]; ok {
		c.Args = cast.ToString(val)
	}

	if val, ok := pod.Annotations[AnnotationContainerInitContainer]; ok {
		c.InitContainer = cast.ToBool(val)
	}

	if val, ok := pod.Annotations[AnnotationContainerInitFirst]; ok {
		c.InitFirst = cast.ToBool(val)
	}

	if val, ok := pod.Annotations[AnnotationContainerPullPolicy]; ok {
		c.ImagePullPolicy = cast.ToString(val)
	}

	if val, ok := pod.Annotations[AnnotationContainerConfigMap]; ok {
		c.ConfigMapName = cast.ToString(val)
	}

	if val, ok := pod.Annotations[AnnotationContainerLimitsCPU]; ok {
		c.LimitsCPU = cast.ToString(val)
	}

	if val, ok := pod.Annotations[AnnotationContainerLimitsMem]; ok {
		c.LimitsMem = cast.ToString(val)
	}

	if val, ok := pod.Annotations[AnnotationContainerRequestsCPU]; ok {
		c.RequestsCPU = cast.ToString(val)
	}

	if val, ok := pod.Annotations[AnnotationContainerRequestsMem]; ok {
		c.RequestsMem = cast.ToString(val)
	}

	if val, ok := pod.Annotations[AnnotationContainerRunAsUser]; ok {
		c.RunAsUser = cast.ToInt64(val)
	}

	if val, ok := pod.Annotations[AnnotationContainerRunAsGroup]; ok {
		c.RunAsGroup = cast.ToInt64(val)
	}

	if val, ok := pod.Annotations[AnnotationContainerTLSSecret]; ok {
		c.TLSSecret = cast.ToString(val)
	}

	return c, nil
}

// Validate verifies the coherence of all parameters specified by annotations
// is correct.
//func (c *Container) Validate() {

//}

// Patch creates the necessary pod patches to inject the container.
func (c *Container) Patch() {
	container, err := c.createContainer()
	if err != nil {
		return
	}

	c.Patches = append(c.Patches, addContainers(
		c.Pod.Spec.Containers,
		[]corev1.Container{container},
		"/spec/containers")...)

	fmt.Printf("%+v\n", c.Patches)
}

//
// INTERNAL FUNCTIONS
//

func (c *Container) createContainer() (corev1.Container, error) {
	var command []string
	var args []string

	if c.Command != "" {
		command = []string{c.Command}
	}

	if c.Args != "" {
		args = []string{c.Args}
	}

	envs, err := c.parseAnnotationsEnvVars()
	if err != nil {
		return corev1.Container{}, err
	}

	return corev1.Container{
		Name:            c.Name,
		Image:           c.ImageName,
		ImagePullPolicy: corev1.PullPolicy(c.ImagePullPolicy),
		Env:             envs,
		//Resources:       resources,
		//SecurityContext: a.securityContext(),
		//VolumeMounts:    volumeMounts,
		//Lifecycle:       &lifecycle,
		Command: command,
		Args:    args,
	}, nil
}

func (c *Container) parseAnnotationsEnvVars() ([]corev1.EnvVar, error) {
	var envs []corev1.EnvVar

	for k, v := range c.Pod.Annotations {
		if strings.HasPrefix(k, AnnotationPrefixEnv+"-") {
			var envName string

			_, err := fmt.Sscanf(k, AnnotationPrefixEnv+"-%s", &envName)
			if err != nil {
				return nil, err
			}

			envs = append(envs, corev1.EnvVar{
				Name:  envName,
				Value: v,
			})
		}
	}

	return envs, nil
}

//func (c *Container) securityContext() *corev1.SecurityContext {
//runAsNonRoot := true
//if c.RunAsUser == 0 || c.RunAsGroup == 0 {
//runAsNonRoot = false
//}

//return &corev1.SecurityContext{
//RunAsUser:    &c.RunAsUser,
//RunAsGroup:   &c.RunAsGroup,
//RunAsNonRoot: &runAsNonRoot,
//}
//}

//func getServiceAccount(pod *corev1.Pod) (string, string) {
//for _, container := range pod.Spec.Containers {
//for _, volumes := range container.VolumeMounts {
//if strings.Contains(volumes.MountPath, "serviceaccount") {
//return volumes.Name, volumes.MountPath
//}
//}
//}

//return "", ""
//}

func newAnnotationError(annotation string) error {
	return fmt.Errorf("Annotation '%s' not found", annotation)
}
