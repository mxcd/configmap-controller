package repository

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"

	"github.com/mxcd/configmap-controller/internal/util"
	cache "github.com/mxcd/go-cache"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/types"
)

type ConfigMapRepository struct {
	// holds all jobs by NamespacedName
	Cache     *cache.LocalCache[types.NamespacedName, corev1.ConfigMap]
	listeners []RepositoryEventListener[corev1.ConfigMap]
}

var configMapRepository *ConfigMapRepository

func GetConfigMapRepository() *ConfigMapRepository {
	if configMapRepository == nil {
		configMapRepository = &ConfigMapRepository{
			Cache: cache.NewLocalCache[types.NamespacedName, corev1.ConfigMap](&cache.LocalCacheOptions[types.NamespacedName]{
				TTL:      0,
				Size:     0,
				CacheKey: &NamespacedNameCacheKey{},
			}),
			listeners: []RepositoryEventListener[corev1.ConfigMap]{},
		}
	}

	return configMapRepository
}

func (r *ConfigMapRepository) AddListener(listener RepositoryEventListener[corev1.ConfigMap]) {
	r.listeners = append(r.listeners, listener)
}

func (r *ConfigMapRepository) Notify(event *RepositoryEvent[corev1.ConfigMap]) {
	for _, listener := range r.listeners {
		listener.Handle(event)
	}
}

func (r *ConfigMapRepository) GetJob(ctx context.Context, name types.NamespacedName) (*corev1.ConfigMap, error) {
	log.Trace().Msgf("Getting ConfigMap: %s", util.GetNamespacedNameString(name))
	job, ok := r.Cache.Get(name)
	if !ok {
		log.Warn().Msgf("ConfigMap not found: %s", util.GetNamespacedNameString(name))
		return nil, errors.New("ConfigMap not found")
	}
	return job, nil
}

func (r *ConfigMapRepository) SetConfigMap(ctx context.Context, name types.NamespacedName, configMap *corev1.ConfigMap) error {
	log.Trace().Msgf("Setting ConfigMap: %s", util.GetNamespacedNameString(name))
	r.Cache.Set(name, *configMap)
	r.Notify(&RepositoryEvent[corev1.ConfigMap]{
		Type:    RepositoryEventUpdated,
		Name:    name,
		Element: configMap,
	})
	return nil
}

func (r *ConfigMapRepository) RemoveConfigMap(ctx context.Context, name types.NamespacedName) error {
	log.Trace().Msgf("Deleting ConfigMap: %s", util.GetNamespacedNameString(name))
	_, ok := r.Cache.Get(name)
	if !ok {
		log.Debug().Msgf("ConfigMap not found: %s", util.GetNamespacedNameString(name))
		return errors.New("configMap not found")
	}

	r.Cache.Remove(name)
	r.Notify(&RepositoryEvent[corev1.ConfigMap]{
		Type: RepositoryEventDeleted,
		Name: name,
	})

	return nil
}

func (r *ConfigMapRepository) GetAllConfigMaps(ctx context.Context) ([]*corev1.ConfigMap, error) {
	log.Trace().Msg("Getting all ConfigMaps")
	entries, err := r.Cache.Load()
	if err != nil {
		log.Error().Err(err).Msg("Error loading ConfigMaps")
		return nil, err
	}

	configMaps := make([]*corev1.ConfigMap, len(entries))
	for i, entry := range entries {
		configMaps[i] = entry.Value
	}

	return configMaps, nil
}
