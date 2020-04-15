GENERIC KUBERNETES CONTAINER INJECTOR
-----

This is a generic Kubernetes container injector. It means that it can inject any container into a Pod using Mutating Admission Webhook. The code is largely inspired by the Hashicorp Vault Agent Injector.

Any container can be injected into a Pod using:
- Pod annotations: User makes the choice to inject or not. It is currently supported.
- Configuration (later): Containers are arbitrarily injected to the Pod according to several criteria such as namespaces, service account etc. defined in the configuration file. Not available at this moment but later.

### Deployment

The `container-injector` can be deployed to any Kubernetes cluster using `Kustomize`.

#### Generate certificates for Mutating Admission Webhook

```bash
$ kustomize build deploy/certs-init | kubectl apply -f -
```

#### Deploy container-injector

Deploy `container-injector` without waiting for `certs-init` job completed.

```bash
$ kustomize build deploy/container-injector | kubectl apply -f -
```

Check if everything goes well:

```bash
$ kubectl get pods -n container-injector
NAME                                  READY   STATUS      RESTARTS   AGE
certs-init-5czxg                      0/1     Completed   0          25h
container-injector-6d6c67b54d-cskf7   1/1     Running     0          24h
```

### Examples

#### Inject a simple container

The following annotations allow to inject a simple container:

```yaml
  template:
    metadata:
      annotations:
        container-injector.uthng.me/inject: "true"
        container-injector.uthng.me/name: "sleep"
        container-injector.uthng.me/image: "pstauffer/curl"
        container-injector.uthng.me/command: "/bin/sleep 3650d"
```

#### Inject a complexe container

Injection of a container with envrionment variables, volume mounts and volume sources can be done as below:

```yaml
  template:
    metadata:
      annotations:
        container-injector.uthng.me/inject: "true"
        container-injector.uthng.me/name: "git-sync"
        container-injector.uthng.me/image: "k8s.gcr.io/git-sync:v3.1.3"
        container-injector.uthng.me/env-GIT_SYNC_REPO: "https://github.com/kubernetes/git-sync.git"
        container-injector.uthng.me/env-GIT_SYNC_DEST: "git-sync"
        container-injector.uthng.me/env-GIT_SYNC_WAIT: "10"
        container-injector.uthng.me/volume-mount-markdown: "/tmp/git"
        container-injector.uthng.me/volume-source-markdown: '{"emptyDir": {}}'
```

### Annotations

- **container-injector.uthng.me/status:** is added to a pod after an injection is done. It must be `injected`.
- **container-injector.uthng.me/inject:** controls whether injection is explicitly enabled or disabled for a pod. This should be set to a "true" or "false" value.
- **container-injector.uthng.me/name:** is the name of the injected container.
- **container-injector.uthng.me/image:** is the name of the  docker image to use.
- **container-injector.uthng.me/command:** specifies the command to be executed when the container starts.
- **container-injector.uthng.me/args:** specifies the list of arguments for the command to be passed to the command when the container starts.
- **container-injector.uthng.me/env:** specifies the environment variables and their values for the container. The name of the environment variables is the part after `container-injector.uthng.me/env-` such as `container-injector.uthng.me/env-TLS_SECRETS`. The value must be a simple string.
- **container-injector.uthng.me/volume-mount:** specifies the volume mount paths in the container. The name of the volumes is the part after `container-injector.uthng.me/volume-mount-` such as`container-injector.uthng.me/volume-mount-config`. Value can be a simple string or json string. For example: `/opt/gitConfig` or `{"mountPath": "/opt/gitConfig", "readOnly": true}`.
- **container-injector.uthng.me/volume-source:** specifies the source of volumes to mount in the pod. The name of the volume source is the part after `container-injector.uthng.me/volume-source-` such as `container-injector.uthng.me/volume-source-tlscert`. Value must be a json string. For example: `{"emptyDir": {}}`.
