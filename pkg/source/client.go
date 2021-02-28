package source

import (
	"context"

	"github.com/kelseyhightower/envconfig"
	"github.com/nats-io/stan.go"
)

type AdapterSASL struct {
	Enable   bool   `envconfig:"NATSS_NET_SASL_ENABLE" required:"false"`
	User     string `envconfig:"NATSS_NET_SASL_USER" required:"false"`
	Password string `envconfig:"NATSS_NET_SASL_PASSWORD" required:"false"`
	Type     string `envconfig:"NATSS_NET_SASL_TYPE" required:"false"`
}

type AdapterTLS struct {
	Enable bool   `envconfig:"NATSS_NET_TLS_ENABLE" required:"false"`
	Cert   string `envconfig:"NATSS_NET_TLS_CERT" required:"false"`
	Key    string `envconfig:"NATSS_NET_TLS_KEY" required:"false"`
	CACert string `envconfig:"NATSS_NET_TLS_CA_CERT" required:"false"`
}

type AdapterNet struct {
	SASL AdapterSASL
	TLS  AdapterTLS
}

type NatssEnvConfig struct {
	NatssUrl  string `envconfig:"NATSS_URL" required:"true"`
	ClusterID string `envconfig:"CLUSTER_ID" required:"true"`
	ClientID  string `envconfig:"CLIENT_ID" required:"false"`
	Net       AdapterNet
}

// NewConfig extracts the NATSS configuration from the environment.
func NewConfig(ctx context.Context) (*NatssEnvConfig, *stan.Options, error) {
	var env NatssEnvConfig
	if err := envconfig.Process("", &env); err != nil {
		return nil, nil, err
	}
	opts := stan.GetDefaultOptions()

	return &env, &opts, nil
}
