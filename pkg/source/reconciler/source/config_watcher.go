/*
Copyright 2019 The Knative Authors

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

package source

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/eventing-natss/pkg/common/constants"
	"knative.dev/eventing/pkg/reconciler/source"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/logging"
)

const (
	EnvNatssCfg           = "K_NATSS_CONFIG"
	natssConfigMapNameEnv = "CONFIG_NATSS_NAME"
)

type NatssConfig struct {
	YamlString string
}

type NatssSourceConfigAccessor interface {
	source.ConfigAccessor
	NatssConfig() *NatssConfig
}

// ToEnvVars serializes the contents of the ConfigWatcher to individual
// environment variables.
func (cw *NatssSourceConfigWatcher) ToEnvVars() []corev1.EnvVar {
	envs := cw.ConfigWatcher.ToEnvVars()

	if cw.NatssConfig() != nil {
		envs = append(envs, cw.natssConfigEnvVar())
	}

	return envs
}

// NatssConfig returns the logging configuration from the ConfigWatcher.
func (cw *NatssSourceConfigWatcher) NatssConfig() *NatssConfig {
	if cw == nil {
		return nil
	}
	return cw.natssCfg
}

var _ NatssSourceConfigAccessor = (*NatssSourceConfigWatcher)(nil)

type NatssSourceConfigWatcher struct {
	*source.ConfigWatcher
	logger *zap.SugaredLogger

	natssCfg *NatssConfig
}

// WatchConfigurations returns a ConfigWatcher initialized with the given
// options. If no option is passed, the ConfigWatcher observes ConfigMaps for
// logging, metrics, tracing and natss.
func WatchConfigurations(loggingCtx context.Context, component string,
	cmw configmap.Watcher) *NatssSourceConfigWatcher {

	configWatcher := source.WatchConfigurations(loggingCtx, component, cmw)

	cw := &NatssSourceConfigWatcher{
		ConfigWatcher: configWatcher,
		logger:        logging.FromContext(loggingCtx),
		natssCfg:      nil,
	}

	WatchConfigMapWithNatss(cw, cmw)

	return cw
}

// WatchConfigMapWithNatss observes a Natss ConfigMap.
func WatchConfigMapWithNatss(cw *NatssSourceConfigWatcher, cmw configmap.Watcher) {
	cw.natssCfg = &NatssConfig{}
	watchConfigMap(cmw, NatssConfigMapName(), cw.updateFromNatssConfigMap)
}

func watchConfigMap(cmw configmap.Watcher, cmName string, obs configmap.Observer) {
	if dcmw, ok := cmw.(configmap.DefaultingWatcher); ok {
		dcmw.WatchWithDefault(corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: cmName},
			Data:       map[string]string{},
		}, obs)

	} else {
		cmw.Watch(cmName, obs)
	}
}

func (cw *NatssSourceConfigWatcher) updateFromNatssConfigMap(cfg *corev1.ConfigMap) {
	natssCfg, err := NewNatssConfigFromConfigMap(cfg)
	if err != nil {
		cw.logger.Warnw("ignoring configuration in Natss configmap ", zap.String("cfg.Name", cfg.Name), zap.Error(err))
		return
	}

	cw.natssCfg = natssCfg

	cw.logger.Debugw("Updated Natss config from ConfigMap", zap.Any("ConfigMap", cfg))
	cw.logger.Debugf("Updated Natss config from ConfigMap 2 %+v", cw.natssCfg)
}

func NewNatssConfigFromConfigMap(cfg *corev1.ConfigMap) (*NatssConfig, error) {
	if cfg == nil {
		return nil, fmt.Errorf("natss configmap does not exist")
	}
	// todo(yangyunfeng)
	if _, ok := cfg.Data[constants.NatssSettingsConfigKey]; !ok {
		return nil, fmt.Errorf("'%s' key does not exist in natss configmap", constants.NatssSettingsConfigKey)
	}
	delete(cfg.Data, "_example")
	return &NatssConfig{
		YamlString: cfg.Data[constants.NatssSettingsConfigKey],
	}, nil
}

// natssConfigEnvVar returns an EnvVar containing the serialized Natss
// configuration from the ConfigWatcher.
func (cw *NatssSourceConfigWatcher) natssConfigEnvVar() corev1.EnvVar {
	cfg, err := natssConfigToJSON(cw.NatssConfig())
	if err != nil {
		cw.logger.Warnw("Error while serializing Natss config", zap.Error(err))
	}

	return corev1.EnvVar{
		Name:  EnvNatssCfg,
		Value: cfg,
	}
}

func natssConfigToJSON(cfg *NatssConfig) (string, error) { //nolint // for backcompat.
	if cfg == nil {
		return "", nil
	}

	jsonOpts, err := json.Marshal(cfg)
	return string(jsonOpts), err
}

// ConfigMapName gets the name of the Natss ConfigMap
func NatssConfigMapName() string {
	if cm := os.Getenv(natssConfigMapNameEnv); cm != "" {
		return cm
	}
	return constants.SettingsConfigMapName
}
