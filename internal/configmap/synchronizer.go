package configmap

import (
	"context"
	"sync"
	"time"

	"github.com/mxcd/configmap-controller/internal/controller"
	"github.com/mxcd/configmap-controller/internal/redis"
	"github.com/mxcd/configmap-controller/internal/repository"
	"github.com/mxcd/configmap-controller/internal/util"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
)

type ConfigMapSynchronizer struct {
	options *ConfigMapSynchronizerOptions
	jobs    map[string]*ConfigMapSynchronizationJob
	lock    *sync.Mutex
}

type ConfigMapSynchronizerOptions struct {
	Redis      *redis.RedisConnection
	Reconciler *controller.ConfigMapReconciler
}

type ConfigMapSynchronizationJob struct {
	ConfigMap       *corev1.ConfigMap
	DataHash        string
	RedisConnection *redis.RedisConnection
	Reconciler      *controller.ConfigMapReconciler
	Running         bool
	Lock            *sync.Mutex
}

func (s *ConfigMapSynchronizer) Handle(event *repository.RepositoryEvent[corev1.ConfigMap]) {
	switch event.Type {
	case repository.RepositoryEventUpdated:
		s.handleConfigMapUpdated(event)
	case repository.RepositoryEventDeleted:
		s.handleConfigMapDeleted(event)
	default:
		log.Warn().Str("type", string(event.Type)).Msg("unsupported event type")
	}
}

func NewConfigMapSynchronizer(options *ConfigMapSynchronizerOptions) *ConfigMapSynchronizer {
	return &ConfigMapSynchronizer{
		options: options,
		jobs:    make(map[string]*ConfigMapSynchronizationJob),
		lock:    &sync.Mutex{},
	}
}

func (s *ConfigMapSynchronizer) Stop() {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, job := range s.jobs {
		job.Stop()
	}
}

func (s *ConfigMapSynchronizer) handleConfigMapUpdated(event *repository.RepositoryEvent[corev1.ConfigMap]) {
	namespacedNameString := util.GetNamespacedNameString(event.Name)

	if job, ok := s.jobs[namespacedNameString]; ok {
		job.Lock.Lock()
		defer job.Lock.Unlock()
		job.ConfigMap = event.Element
		job.WriteRedisConfigMap(context.Background())
	} else {
		job := &ConfigMapSynchronizationJob{
			ConfigMap:       event.Element,
			DataHash:        "",
			Reconciler:      s.options.Reconciler,
			RedisConnection: s.options.Redis,
			Running:         false,
			Lock:            &sync.Mutex{},
		}
		s.lock.Lock()
		s.jobs[namespacedNameString] = job
		s.lock.Unlock()
		job.WriteRedisConfigMap(context.Background())
		job.Start()
	}
}

func (s *ConfigMapSynchronizer) handleConfigMapDeleted(event *repository.RepositoryEvent[corev1.ConfigMap]) {
	namespacedNameString := util.GetNamespacedNameString(event.Name)
	if job, ok := s.jobs[namespacedNameString]; ok {
		job.Stop()
		s.lock.Lock()
		delete(s.jobs, namespacedNameString)
		s.lock.Unlock()
	} else {
		log.Warn().Str("name", namespacedNameString).Msg("job not found")
	}
}

func (j *ConfigMapSynchronizationJob) Start() {
	j.Lock.Lock()
	defer j.Lock.Unlock()
	if j.Running {
		return
	}
	j.Running = true
	go j.run()
}

func (j *ConfigMapSynchronizationJob) Stop() {
	j.Lock.Lock()
	defer j.Lock.Unlock()
	j.Running = false
}

func (j *ConfigMapSynchronizationJob) run() {
	namespacedNameString := util.GetConfigMapNamespacedNameString(j.ConfigMap)

	for {
		j.Lock.Lock()
		running := j.Running
		j.Lock.Unlock()
		if !running {
			return
		}

		ctx := context.Background()

		err := j.pullRedisConfigMap(ctx)
		if err != nil {
			log.Error().Err(err).Str("name", namespacedNameString).Msg("unable to pull configmap from redis")
		}

		time.Sleep(1 * time.Second)
	}
}
