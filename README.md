# statefulset-pvc-cleaner

## Overview

This project is a Kubernetes controller to handle garbage collection of `PersistentVolumeClaims` created by `StatefulSets`.

This controller is simpler and is not intended to replace the [Kubernetes StatefulSet PVC Auto-Deletion feature](https://kubernetes.io/blog/2021/12/16/kubernetes-1-23-statefulset-pvc-auto-deletion/) but rather to provide a PersistentVolumeClaim garbage collection mechanism on clusters that do not yet have this feature.

## Installation

### Helm

Helm chart coming soon!

## Usage

In order to handle the removal of `PersisPersistentVolumeClaims` from a `StatefulSet` when it is deleted, this controller simply updates them to set an owner reference to the `StatefulSet` (to learn more about garbage collection of kubernetes resources, refer to [this documentation](https://kubernetes.io/docs/concepts/architecture/garbage-collection/#owners-dependents)).

To do this, it relies on two annotations to put on the `PersisPersistentVolumeClaims`.

- **`statefulset-pvc-cleaner.quortex.io/retention`**: Set this annotation value to `delete` to enable garbage collection on this `PersisPersistentVolumeClaim`.
- **`statefulset-pvc-cleaner.quortex.io/statefulset`**: Specify the name of the `PersisPersistentVolumeClaim`'s `StatefulSet` as the value of this annotation. This is required by the controller to avoid heavy processing in order to deduce it.

Once the owner reference is defined, we have no way of knowing if the owner reference of the `StatefulSet` is ours alone (another controller could be responsible for similar behavior...), so we don't proceed to remove the reference, even if the annotations changes.

An example of `StatefulSet` configuration, in this case, only `xxx` `PersisPersistentVolumeClaim` will be garbage collected on `StatefulSet`'s deletion.

```yml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
spec:
  serviceName: "nginx"
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
        - name: nginx
          image: nginx:latest
          ports:
            - containerPort: 80
              name: web
          volumeMounts:
            - name: www
              mountPath: /usr/share/nginx/html
  volumeClaimTemplates:
    - metadata:
        name: www
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 1Gi
    - metadata:
        name: xxx
        annotations:
          statefulset-pvc-cleaner.quortex.io/retention: delete
          statefulset-pvc-cleaner.quortex.io/statefulset: web
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 1Gi
```

## Configuration

### <a id="Configuration_Optional_args"></a>Optional args

The `statefulset-pvc-cleaner` container takes the following flags as argument.

| Key                         | Description                                                                                                                                                                      | Default                                                              |
| --------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------- |
| `metrics-bind-address`      | The address the metric endpoint binds to.                                                                                                                                        | `:8080`                                                              |
| `health-probe-bind-address` | The address the probe endpoint binds to.                                                                                                                                         | `:8081`                                                              |
| `leader-elect`              | Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.                                                            | `false`                                                              |
| `zap-devel`                 | Whether to enable development logs.                                                                                                                                              | `false`                                                              |
| `zap-encoder`               | Logs encoding (one of `json` or `console`)                                                                                                                                       | Defaults to `console` when `zap-devel` is true and `json` otherwise. |
| `zap-log-level`             | Level to configure the verbosity of logging. Can be one of `debug`, `info`, `error`, or any integer value > 0 which corresponds to custom debug levels of increasing verbosity). | Defaults to `debug` when `zap-devel` is true and `info` otherwise.   |
| `zap-stacktrace-level`      | Level at and above which stacktraces are captured (one of `info`, `error` or `panic`).                                                                                           | Defaults to `warn` when `zap-devel` is true and `error` otherwise.   |
| `zap-time-encoding`         | Logs time encoding (one of `epoch`, `millis`, `nano`, `iso8601`, `rfc3339` or `rfc3339nano`).                                                                                    | `epoch`                                                              |

## License

Distributed under the Apache 2.0 License. See `LICENSE` for more information.

## Contributing

To contribute to this project, please first consult the [contribution rules guide](CONTRIBUTING.md).

Got a question?
File a GitHub [issue](https://github.com/quortex/statefulset-pvc-cleaner/issues).
