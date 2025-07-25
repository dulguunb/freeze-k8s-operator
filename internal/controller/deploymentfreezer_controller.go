/*
Copyright 2025.

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

package controller

import (
	"context"
	"fmt"
	"strconv"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	dulguuntestiov1alpha1 "github.com/example/my-operator/api/v1alpha1"
)

const (
	frozenByAnnotation         = "frozenby"
	frozenByReplicasAnnotation = "frozenby-replicas"
	frozenByTimeAnnotation     = "frozenby-time"
)

// DeploymentFreezerReconciler reconciles a DeploymentFreezer object
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;update;patch
type DeploymentFreezerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=dulguun-test.io.dulguun-test.io,resources=deploymentfreezers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=dulguun-test.io.dulguun-test.io,resources=deploymentfreezers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=dulguun-test.io.dulguun-test.io,resources=deploymentfreezers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DeploymentFreezer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *DeploymentFreezerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	var freezer dulguuntestiov1alpha1.DeploymentFreezer
	if err := r.Get(ctx, req.NamespacedName, &freezer); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var deploy appsv1.Deployment
	deployKey := types.NamespacedName{
		Name:      freezer.Spec.DeploymentName,
		Namespace: freezer.Spec.DeploymentNamespace,
	}
	if err := r.Get(ctx, deployKey, &deploy); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	annotations := deploy.Annotations
	if annotations == nil {
		annotations = map[string]string{}
	}

	now := time.Now().UTC()
	shouldFreeze := true
	shouldUnfreeze := false
	var frozenSince *time.Time

	// Check if already frozen by another DeploymentFreezer
	if frozenBy, exists := annotations[frozenByAnnotation]; exists && frozenBy != freezer.Name {
		logger.Info("Deployment is already frozen by another DeploymentFreezer", "frozenby", frozenBy)
		freezer.Status.IsFrozen = false
		freezer.Status.Reason = fmt.Sprintf("Deployment is already frozen by %s", frozenBy)
		_ = r.Status().Update(ctx, &freezer)
		return ctrl.Result{}, nil
	}

	// Check if we are already frozen by this resource
	if annotations[frozenByAnnotation] == freezer.Name {
		// Parse frozen time
		if tStr, ok := annotations[frozenByTimeAnnotation]; ok {
			t, err := time.Parse(time.RFC3339, tStr)
			if err == nil {
				frozenSince = &t
			}
		}
		shouldFreeze = false
	}

	// If not frozen, freeze now
	if shouldFreeze {
		// Store original replicas in annotation
		originalReplicas := int32(1)
		if deploy.Spec.Replicas != nil {
			originalReplicas = *deploy.Spec.Replicas
		}
		annotations[frozenByAnnotation] = freezer.Name
		annotations[frozenByReplicasAnnotation] = fmt.Sprintf("%d", originalReplicas)
		annotations[frozenByTimeAnnotation] = now.Format(time.RFC3339)
		var zero int32 = 0
		deploy.Spec.Replicas = &zero
		deploy.Annotations = annotations
		if err := r.Update(ctx, &deploy); err != nil {
			logger.Error(err, "Failed to freeze deployment")
			return ctrl.Result{}, err
		}
		frozenSince = &now
		logger.Info("Deployment frozen", "deployment", deployKey.String(), "originalReplicas", originalReplicas)
	}

	// If frozen, check if duration has elapsed
	if frozenSince != nil {
		duration := time.Duration(freezer.Spec.DurationSeconds) * time.Second
		frozenFor := now.Sub(*frozenSince)
		freezer.Status.FrozenSince = &metav1.Time{Time: *frozenSince}
		freezer.Status.FrozenDuration = frozenFor.String()
		freezer.Status.IsFrozen = true
		freezer.Status.Reason = ""
		_ = r.Status().Update(ctx, &freezer)
		if duration > 0 && frozenFor >= duration {
			shouldUnfreeze = true
		}
	}

	if shouldUnfreeze {
		// Only unfreeze if frozen by this resource
		origReplicasStr, ok := annotations[frozenByReplicasAnnotation]
		if ok {
			if origReplicas, err := strconv.Atoi(origReplicasStr); err == nil {
				orig := int32(origReplicas)
				deploy.Spec.Replicas = &orig
				logger.Info("Restoring original replicas", "replicas", orig)
			} else {
				logger.Error(err, "Malformed frozenby-replicas annotation, storing to 1 replica count")
				// Maybe make the replica to 1 instead of erroring out?
				orig := int32(1)
				deploy.Spec.Replicas = &orig
			}
		} else {
			logger.Info("No frozenby-replicas annotation found, cannot restore replicas")
		}
		// Remove annotations
		delete(annotations, frozenByAnnotation)
		delete(annotations, frozenByReplicasAnnotation)
		delete(annotations, frozenByTimeAnnotation)
		deploy.Annotations = annotations
		if err := r.Update(ctx, &deploy); err != nil {
			logger.Error(err, "Failed to unfreeze deployment")
			return ctrl.Result{}, err
		}
		freezer.Status.IsFrozen = false
		freezer.Status.Reason = "Unfrozen after duration elapsed"
		_ = r.Status().Update(ctx, &freezer)
		logger.Info("Deployment unfrozen after duration", "deployment", deployKey.String())
	}

	// Requeue if still frozen and duration not yet elapsed
	if freezer.Status.IsFrozen && !shouldUnfreeze && freezer.Spec.DurationSeconds > 0 {
		timeLeft := time.Duration(freezer.Spec.DurationSeconds)*time.Second - now.Sub(*frozenSince)
		if timeLeft > 0 {
			return ctrl.Result{RequeueAfter: timeLeft}, nil
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeploymentFreezerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dulguuntestiov1alpha1.DeploymentFreezer{}).
		Named("deploymentfreezer").
		Complete(r)
}
