package configmap

import (
	"context"
	"encoding/json"

	"github.com/mxcd/configmap-controller/internal/util"
	"github.com/rs/zerolog/log"
	"github.com/zeebo/blake3"
)

func (j *ConfigMapSynchronizationJob) pullRedisConfigMap(ctx context.Context) error {
	key := util.GetConfigMapNamespacedNameString(j.ConfigMap)

	configMapData, err := j.RedisConnection.Client.HGetAll(ctx, key).Result()
	if err != nil {
		log.Err(err).Str("name", key).Msg("unable to get configmap data from redis")
		return err
	}

	// config map not in redis
	if len(configMapData) == 0 {
		log.Debug().Str("name", key).Msg("configmap data not found in redis")
		return j.WriteRedisConfigMap(ctx)
	}

	// config map with data in redis => remove empty flag
	_, hasEmptyFlag := configMapData["_empty"]
	if len(configMapData) > 1 && hasEmptyFlag {
		_, err = j.RedisConnection.Client.HDel(ctx, key, "_empty").Result()
		if err != nil {
			log.Err(err).Str("name", key).Msg("unable to remove empty flag from redis")
			return err
		}
	}

	delete(configMapData, "_empty")

	hashString := generateConfigMapDataHash(configMapData)
	if hashString == j.DataHash {
		log.Trace().Str("name", key).Msg("configmap data unchanged")
		return nil
	}

	log.Info().Str("name", key).Msg("updating configmap data in k8s")

	j.ConfigMap.Data = configMapData
	j.DataHash = hashString

	err = j.Reconciler.Update(ctx, j.ConfigMap)
	if err != nil {
		log.Err(err).Str("name", key).Msg("unable to update configmap data in k8s")
		return err
	}

	log.Debug().Str("name", key).Msg("configmap data updated")
	return nil
}

func (j *ConfigMapSynchronizationJob) WriteRedisConfigMap(ctx context.Context) error {
	key := util.GetConfigMapNamespacedNameString(j.ConfigMap)

	log.Info().Str("name", key).Msg("writing configmap data to redis")

	configMapData := j.ConfigMap.Data

	if configMapData == nil {
		configMapData = make(map[string]string)
	}

	// place empty key if no data exists so HMSet will be written to redis
	if len(configMapData) == 0 {
		log.Debug().Str("name", key).Msg("configmap data is empty")
		configMapData["_empty"] = ""
	}

	redisConfigMapData, err := j.RedisConnection.Client.HGetAll(ctx, key).Result()
	if err != nil {
		log.Err(err).Str("name", key).Msg("unable to get configmap data from redis")
		return err
	}

	// remove keys that are not in the configMapData
	removeFields := []string{}
	for k := range redisConfigMapData {
		if _, ok := configMapData[k]; !ok {
			removeFields = append(removeFields, k)
		}
	}

	if len(removeFields) > 0 {
		_, err = j.RedisConnection.Client.HDel(ctx, key, removeFields...).Result()
		if err != nil {
			log.Err(err).Str("name", key).Msg("unable to remove fields from redis")
			return err
		}
	}

	_, err = j.RedisConnection.Client.HMSet(ctx, key, configMapData).Result()
	if err != nil {
		log.Err(err).Str("name", key).Msg("unable to write configmap data to redis")
		return err
	}

	j.DataHash = generateConfigMapDataHash(configMapData)

	log.Debug().Str("name", key).Msg("configmap data written")
	return nil
}

func generateConfigMapDataHash(configMapData map[string]string) string {
	delete(configMapData, "_empty")
	data, err := json.Marshal(configMapData)
	if err != nil {
		log.Err(err).Msg("unable to marshal configmap data")
		return ""
	}

	hash := blake3.Sum256(data)
	return string(hash[:])
}
