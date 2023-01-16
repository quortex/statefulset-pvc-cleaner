/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	annDomain       = "statefulset-pvc-cleaner.quortex.io"
	annRetention    = annDomain + "/" + "retention"
	annStatefulSet  = annDomain + "/" + "statefulset"
	retentionDelete = "delete"
)

// PersistentVolumeClaimReconciler reconciles a PersistentVolumeClaim object
type PersistentVolumeClaimReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.1/pkg/reconcile
func (r *PersistentVolumeClaimReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.V(1).Info("PersistentVolumeClaim reconcile function started")

	pvc := &corev1.PersistentVolumeClaim{}
	err := r.Get(ctx, req.NamespacedName, pvc)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile
			// request.
			log.Info("PersistentVolumeClaim resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get PersistentVolumeClaim")
		return ctrl.Result{}, err
	}

	// If the PersistentVolumeClaim contains the
	// statefulset-pvc-cleaner.quortex.io/retention: "delete" annotation, we add
	// the StatefulSet referenced in the
	// statefulset-pvc-cleaner.quortex.io/statefulset annotation as an owner
	// reference of the PersistentVolumeClaim.
	//
	// Thus, we let the kubernetes garbage collection mechanism manage the
	// possible deletion of the PersistentVolumeClaim in case of deletion of it's
	// StatefulSet owner. To learn more about garbage collection of kubernetes
	// resources, refer to this documentation
	// https://kubernetes.io/docs/concepts/architecture/garbage-collection/#owners-dependents
	//
	// We have no way of knowing if the owner reference of the StatefulSet is ours
	// alone (another controller could be responsible for similar behavior...), so
	// we don't proceed to remove the reference.
	if pvc.Annotations[annRetention] == retentionDelete {
		// Query the StatefulSet.
		stsNsName := types.NamespacedName{
			Namespace: req.Namespace,
			Name:      pvc.Annotations[annStatefulSet],
		}
		log = log.WithValues("StatefulSet", struct {
			Namespace, Name string
		}{stsNsName.Namespace, stsNsName.Name})
		sts := &appsv1.StatefulSet{}
		if err := r.Get(ctx, stsNsName, sts); err != nil {
			if errors.IsNotFound(err) {
				log.Info("StatefulSet resource not found. Ignoring since object must be deleted")
				return ctrl.Result{}, nil
			}
			// Error reading the object - requeue the request.
			log.Error(err, "Failed to get StatefulSet")
			return ctrl.Result{}, err
		}

		// Check the StatefulSet consistency with the PersistentVolumeClaim.
		//
		// Kubernetes creates PVCs for each replica specified VolumeClaimTemplate,
		// which is part of the STS specification. The naming scheme for the PVC is
		// $PVC_TEMPLATE_NAME-$STS_NAME-$REPLICA_INDEX.
		matched := false
		for _, vct := range sts.Spec.VolumeClaimTemplates {
			suffix := strings.TrimPrefix(pvc.Name, strings.Join([]string{vct.Name, sts.Name, ""}, "-"))
			if _, err := strconv.Atoi(suffix); err == nil {
				matched = true
				break
			}
		}
		if !matched {
			err := fmt.Errorf("%s annotation consistency error", annStatefulSet)
			log.Error(err, "Invalid StatefulSet")
			return ctrl.Result{}, err
		}

		// Set OwnerReference then update the PersistentVolumeClaim.
		if err := controllerutil.SetOwnerReference(sts, pvc, r.Scheme); err != nil {
			log.Error(err, "Failed to set owner reference")
			return ctrl.Result{}, err
		}
		if err := r.Update(ctx, pvc); err != nil {
			if errors.IsConflict(err) {
				log.V(1).Info("PersistentVolumeClaim update failed due to conflict, reconciliation requeued")
				return ctrl.Result{RequeueAfter: time.Second}, nil
			}
			log.Error(err, "Failed to update PersistentVolumeClaim")
			return ctrl.Result{}, err
		}
	}

	log.Info("PersistentVolumeClaim successfully reconciled")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PersistentVolumeClaimReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.PersistentVolumeClaim{}, r.reconciliationPredicates()).
		Complete(r)
}

// reconciliationPredicates returns predicates for the
// PersistentVolumeClaimReconciler.
func (r *PersistentVolumeClaimReconciler) reconciliationPredicates() builder.Predicates {
	return builder.WithPredicates(predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return r.shouldReconcile(e.Object.(*corev1.PersistentVolumeClaim))
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return r.shouldReconcile(e.ObjectNew.(*corev1.PersistentVolumeClaim)) || r.shouldReconcile(e.ObjectOld.(*corev1.PersistentVolumeClaim))
		},
	})
}

// shouldReconcile returns if given PersistentVolumeClaim should be reconciled
// by the controller.
func (r *PersistentVolumeClaimReconciler) shouldReconcile(obj *corev1.PersistentVolumeClaim) bool {
	// We should consider reconciliation for PersistentVolumeClaimReconciler with
	// statefulset-pvc-cleaner.quortex.io/retention: "delete" annotation and
	// statefulset-pvc-cleaner.quortex.io/statefulset annotation.
	return obj.ObjectMeta.Annotations[annRetention] == retentionDelete && len(obj.ObjectMeta.Annotations[annStatefulSet]) > 0
}
