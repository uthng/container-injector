package sidecar

const (
	// AnnotationContainerStatus is the key of the annotation that is added to
	// a pod after an injection is done.
	// There's only one valid status we care about: "injected".
	AnnotationContainerStatus = "container-injector.uthng.me/status"

	// AnnotationContainerInject is the key of the annotation that controls whether
	// injection is explicitly enabled or disabled for a pod. This should
	// be set to a true or false value, as parseable by strconv.ParseBool
	AnnotationContainerInject = "container-injector.uthng.me/inject"

	// AnnotationContainerName is the name of the  docker image to use.
	AnnotationContainerName = "container-injector.uthng.me/name"

	// AnnotationContainerImage is the name of the  docker image to use.
	AnnotationContainerImage = "container-injector.uthng.me/image"

	// AnnotationContainerCommand is the command to be executed when the container starts
	AnnotationContainerCommand = "container-injector.uthng.me/command"

	// AnnotationContainerArgs is the list of arguments for the command
	// to be executed when the container starts
	AnnotationContainerArgs = "container-injector.uthng.me/args"

	// AnnotationContainerInitFirst makes the initialization container the first container
	// to run when a pod starts. Default is last.
	AnnotationContainerInitContainer = "container-injector.uthng.me/init-container"

	// AnnotationContainerInitFirst makes the initialization container the first container
	// to run when a pod starts. Default is last.
	AnnotationContainerInitFirst = "container-injector.uthng.me/init-first"

	// AnnotationContainerPullPolicy specifies the pull policy for container image.
	AnnotationContainerPullPolicy = "container-injector.uthng.me/pull-policy"

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
