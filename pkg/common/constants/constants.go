/*
Copyright 2020 The Knative Authors

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

package constants

const (

	// NatssChannel Spec Defaults
	DefaultNumPartitions     = 1
	DefaultReplicationFactor = 1

	// The name of the configmap used to hold eventing-Natss settings
	SettingsConfigMapName = "config-natss"
	SettingsSecretName    = "natss-cluster"

	// Mount path of the configmap used to hold eventing-Natss settings
	SettingsConfigMapMountPath = "/etc/" + SettingsConfigMapName

	// Config key of the config in the configmap used to hold eventing-Natss settings
	EventingNatssSettingsConfigKey = "eventing-natss"

	// The name of the keys in the Data section of the eventing-Natss configmap that holds Sarama and Eventing-Natss configuration YAML
	NatssSettingsConfigKey = "natss"
)
