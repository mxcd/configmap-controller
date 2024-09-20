package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/rs/zerolog/log"

	"github.com/mxcd/configmap-controller/internal/repository"
	"github.com/mxcd/configmap-controller/internal/util"
)

type ConfigMapReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *ConfigMapReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log.Trace().Str("name", util.GetNamespacedNameString(req.NamespacedName)).Msg("reconciling configmap")

	configMap := &corev1.ConfigMap{}
	err := r.Get(ctx, req.NamespacedName, configMap)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Debug().Str("name", util.GetNamespacedNameString(req.NamespacedName)).Msg("configmap deleted")
			repository.GetConfigMapRepository().RemoveConfigMap(ctx, req.NamespacedName)
			return ctrl.Result{}, nil
		}
		log.Error().Err(err).Msg("unable to fetch configmap")
		return ctrl.Result{}, err
	}

	if !hasControlAnnotation(configMap) {
		log.Trace().Str("name", util.GetNamespacedNameString(req.NamespacedName)).Msg("configmap not managed")
		return ctrl.Result{}, nil
	}

	if configMap.GetDeletionTimestamp() != nil {
		log.Debug().Str("name", util.GetNamespacedNameString(req.NamespacedName)).Msg("configmap deleted")
		repository.GetConfigMapRepository().RemoveConfigMap(ctx, req.NamespacedName)
	} else {
		log.Debug().Str("name", util.GetNamespacedNameString(req.NamespacedName)).Msg("configmap updated")
		repository.GetConfigMapRepository().SetConfigMap(ctx, req.NamespacedName, configMap)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigMapReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		Complete(r)
}

func hasControlAnnotation(configMap *corev1.ConfigMap) bool {
	_, ok := configMap.Annotations["configmap-controller.mxcd.de/managed"]
	return ok
}
