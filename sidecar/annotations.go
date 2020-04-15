package sidecar

const (
	// AnnotationContainerStatus is the annotation that is added to
	// a pod after an injection is done.
	// The value must be "injected".
	AnnotationContainerStatus = "container-injector.uthng.me/status"

	// AnnotationContainerInject controls whether injection is explicitly
	// enabled or disabled for a pod. This should be set to a true or false value,
	// as parseable by strconv.ParseBool
	AnnotationContainerInject = "container-injector.uthng.me/inject"

	// AnnotationContainerName is the name of the injected container.
	AnnotationContainerName = "container-injector.uthng.me/name"

	// AnnotationContainerImage is the name of the  docker image to use.
	AnnotationContainerImage = "container-injector.uthng.me/image"

	// AnnotationContainerCommand specifies the command to be executed
	// when the container starts.
	AnnotationContainerCommand = "container-injector.uthng.me/command"

	// AnnotationContainerArgs is the list of arguments for the command
	// to be executed when the container starts.
	AnnotationContainerArgs = "container-injector.uthng.me/args"

	// AnnotationContainerInitContainer makes the initialization container the first container
	// to run when a pod starts. Default is last.
	AnnotationContainerInitContainer = "container-injector.uthng.me/init-container"

	// AnnotationContainerInitFirst makes the initialization container the first container
	// to run when a pod starts. Default is last.
	AnnotationContainerInitFirst = "container-injector.uthng.me/init-first"

	// AnnotationContainerPullPolicy specifies the pull policy for container image.
	AnnotationContainerPullPolicy = "container-injector.uthng.me/pull-policy"

	// AnnotationContainerEnv specifies the environment variables and their values
	// for the container. The name of the environment variables is the part after
	// "container-injector.uthng.me/env-" such as "container-injector.uthng.me/env-TLS_SECRETS".
	AnnotationContainerEnv = "container-injector.uthng.me/env"

	// AnnotationContainerVolumeMount specifies the volume mount paths
	// in the container. The name of the volumes is the part after
	// "container-injector.uthng.me/volume-mount-" such as
	//"container-injector.uthng.me/volume-mount-config".
	AnnotationContainerVolumeMount = "container-injector.uthng.me/volume-mount"

	// AnnotationContainerVolumeSource specifies the source of volumes to mount
	// in the pod. The name of the volume source is the part after
	// "container-injector.uthng.me/volume-source" such as
	//"container-injector.uthng.me/volume-source-tlscert".
	AnnotationContainerVolumeSource = "container-injector.uthng.me/volume-source"

	// AnnotationContainerConfigMap is the name of the configuration map where  Container
	// configuration file and templates can be found.
	AnnotationContainerConfigMap = "container-injector.uthng.me/configmap"

	// AnnotationContainerLimitsCPU sets the CPU limit on the  Container containers.
	AnnotationContainerLimitsCPU = "container-injector.uthng.me/limits-cpu"

	// AnnotationContainerLimitsMem sets the memory limit on the  Container containers.
	AnnotationContainerLimitsMem = "container-injector.uthng.me/limits-mem"

	// AnnotationContainerRequestsCPU sets the requested CPU amount on the  Container containers.
	AnnotationContainerRequestsCPU = "container-injector.uthng.me/requests-cpu"

	// AnnotationContainerRequestsMem sets the requested memory amount on the  Container containers.
	AnnotationContainerRequestsMem = "container-injector.uthng.me/requests-mem"

	// AnnotationContainerRunAsUser sets the User ID to run the Container Container containers as.
	AnnotationContainerRunAsUser = "container-injector.uthng.me/run-as-user"

	// AnnotationContainerRunAsGroup sets the Group ID to run the Container Container containers as.
	AnnotationContainerRunAsGroup = "container-injector.uthng.me/run-as-group"

	// AnnotationContainerTLSSecret is the name of the Kubernetes secret containing
	// client TLS certificates and keys.
	AnnotationContainerTLSSecret = "container-injector.uthng.me/tls-secret"
)
