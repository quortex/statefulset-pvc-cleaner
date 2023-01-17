# statefulset-pvc-cleaner

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.1.0](https://img.shields.io/badge/AppVersion-0.1.0-informational?style=flat-square)

A Kubernetes controller to handle garbage collection of PersistentVolumeClaims created by StatefulSets

**Homepage:** <https://github.com/quortex/statefulset-pvc-cleaner>

## Overview
This project is a Kubernetes controller to handle garbage collection of `PersistentVolumeClaims` created by `StatefulSets`.

This controller is simpler and is not intended to replace the [Kubernetes StatefulSet PVC Auto-Deletion feature](https://kubernetes.io/blog/2021/12/16/kubernetes-1-23-statefulset-pvc-auto-deletion/) but rather to provide a `PersistentVolumeClaim` garbage collection mechanism on clusters that do not yet have this feature.

## Installation

1. Add `statefulset-pvc-cleaner` helm repository.

```sh
helm repo add statefulset-pvc-cleaner https://quortex.github.io/statefulset-pvc-cleaner
```

2. Create a namespace for `statefulset-pvc-cleaner`.

```sh
kubectl create ns statefulset-pvc-cleaner-system
```

3. Deploy the appropriate release.

```sh
helm install statefulset-pvc-cleaner statefulset-pvc-cleaner/statefulset-pvc-cleaner -n statefulset-pvc-cleaner-system
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| nameOverride | string | `""` | Helm's name computing override. |
| fullnameOverride | string | `""` | Helm's fullname computing override. |
| replicaCount | int | `1` | Number of desired pods. |
| manager.image.repository | string | `"ghcr.io/quortex/statefulset-pvc-cleaner"` | statefulset-pvc-cleaner manager image repository. |
| manager.image.tag | string | `""` | statefulset-pvc-cleaner manager image tag (default is the chart appVersion). |
| manager.image.pullPolicy | string | `"IfNotPresent"` | statefulset-pvc-cleaner manager image pull policy. |
| manager.resources | object | `{}` | statefulset-pvc-cleaner manager container required resources. |
| manager.securityContext | object | `{}` | statefulset-pvc-cleaner manager container securityContext. |
| manager.livenessProbe.httpGet.path | string | `"/healthz"` | Path of the manager liveness probe. |
| manager.livenessProbe.httpGet.port | int | `8081` | Name or number of the manager liveness probe port. |
| manager.livenessProbe.initialDelaySeconds | int | `15` | Number of seconds before the manager liveness probe is initiated. |
| manager.livenessProbe.periodSeconds | int | `20` | How often (in seconds) to perform the manager liveness probe. |
| manager.readinessProbe.httpGet.path | string | `"/readyz"` | Path of the manager readiness probe. |
| manager.readinessProbe.httpGet.port | int | `8081` | Name or number of the manager readiness probe port. |
| manager.readinessProbe.initialDelaySeconds | int | `5` | Number of seconds before the manager readiness probe is initiated. |
| manager.readinessProbe.periodSeconds | int | `10` | How often (in seconds) to perform the manager readiness probe. |
| manager.logs.development | bool | `false` | Whether to enable development logs. |
| manager.logs.encoder | string | `""` | Logs encoding (one of `json` or `console`). Defaults to `console` when `development` is true and `json` otherwise. |
| manager.logs.logLevel | string | `""` | Level to configure the verbosity of logging. Can be one of `debug`, `info`, `error`, or any integer value > 0 which corresponds to custom debug levels of increasing verbosity). Defaults to `debug` when `development` is true and `info` otherwise. |
| manager.logs.stacktraceLevel | string | `""` | Level at and above which stacktraces are captured (one of `info`, `error` or `panic`). Defaults to `warn` when `development` is true and `error` otherwise. |
| manager.logs.timeEncoding | string | `"epoch"` | Logs time encoding (one of `epoch`, `millis`, `nano`, `iso8601`, `rfc3339` or `rfc3339nano`). |
| kubeRBACProxy.enabled | bool | `true` | Specifies whether kube-rbac-proxy should be created. |
| kubeRBACProxy.image.repository | string | `"gcr.io/kubebuilder/kube-rbac-proxy"` | kube-rbac-proxy image repository. |
| kubeRBACProxy.image.tag | string | `"v0.13.1"` | kube-rbac-proxy image tag. |
| kubeRBACProxy.image.pullPolicy | string | `"IfNotPresent"` | kube-rbac-proxy image pull policy. |
| kubeRBACProxy.resources | object | `{}` | kube-rbac-proxy container required resources. |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created. |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account. |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template. |
| imagePullSecrets | list | `[]` | A list of secrets used to pull containers images. |
| deploymentAnnotations | object | `{}` | Annotations to be added to deployment. |
| podAnnotations | object | `{}` | Annotations to be added to pods. |
| terminationGracePeriod | int | `30` | How long to wait for pods to stop gracefully. |
| podSecurityContext | object | `{}` | Pods securityContext. |
| nodeSelector | object | `{}` | Node labels for statefulset-pvc-cleaner pod assignment. |
| tolerations | list | `[]` | Node tolerations for statefulset-pvc-cleaner scheduling to nodes with taints. |
| affinity | object | `{}` | Affinity for statefulset-pvc-cleaner pod assignment. |
| serviceMonitor.enabled | bool | `false` | Create a prometheus operator ServiceMonitor. |
| serviceMonitor.additionalLabels | object | `{}` | Labels added to the ServiceMonitor. |
| serviceMonitor.annotations | object | `{}` | Annotations added to the ServiceMonitor. |
| serviceMonitor.interval | string | `""` | Override prometheus operator scrapping interval. |
| serviceMonitor.scrapeTimeout | string | `""` | Override prometheus operator scrapping timeout. |
| serviceMonitor.relabelings | list | `[]` | Relabellings to apply to samples before scraping. |

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| vincentmrg |  | <https://github.com/vincentmrg> |
