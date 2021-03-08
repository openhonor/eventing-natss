package source

import (
	"context"

	"github.com/kelseyhightower/envconfig"
	"github.com/nats-io/stan.go"
)

type NatssEnvConfig struct {
	NatssUrl  string `envconfig:"NATSS_URL" required:"true"`
	ClusterID string `envconfig:"NATSS_CLUSTER_ID" required:"true"`
	ClientID  string `envconfig:"NATSS_CLIENT_ID" required:"false"`
}

// NewConfig extracts the Natss configuration from the environment.
func NewConfig(ctx context.Context) (*NatssEnvConfig, *stan.Options, error) {
	var env NatssEnvConfig
	if err := envconfig.Process("", &env); err != nil {
		return nil, nil, err
	}
	opts := stan.GetDefaultOptions()

	return &env, &opts, nil
}
