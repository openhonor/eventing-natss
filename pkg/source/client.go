package source

import (
	"context"
	"github.com/nats-io/stan.go"

	"github.com/kelseyhightower/envconfig"
)

type AdapterSASL struct {
	Enable   bool   `envconfig:"KAFKA_NET_SASL_ENABLE" required:"false"`
	User     string `envconfig:"KAFKA_NET_SASL_USER" required:"false"`
	Password string `envconfig:"KAFKA_NET_SASL_PASSWORD" required:"false"`
	Type     string `envconfig:"KAFKA_NET_SASL_TYPE" required:"false"`
}

type AdapterTLS struct {
	Enable bool   `envconfig:"KAFKA_NET_TLS_ENABLE" required:"false"`
	Cert   string `envconfig:"KAFKA_NET_TLS_CERT" required:"false"`
	Key    string `envconfig:"KAFKA_NET_TLS_KEY" required:"false"`
	CACert string `envconfig:"KAFKA_NET_TLS_CA_CERT" required:"false"`
}

type AdapterNet struct {
	SASL AdapterSASL
	TLS  AdapterTLS
}

type NatssEnvConfig struct {
	BootstrapServers []string `envconfig:"KAFKA_BOOTSTRAP_SERVERS" required:"true"`
	Net              AdapterNet
}

// NewConfig extracts the Kafka configuration from the environment.
func NewConfig(ctx context.Context) ([]string, *stan.Options, error) {
	var env NatssEnvConfig
	if err := envconfig.Process("", &env); err != nil {
		return nil, nil, err
	}

	return NewConfigWithEnv(ctx, &env)
}

// NewConfig extracts the natss option from the environment.
func NewConfigWithEnv(ctx context.Context, env *NatssEnvConfig) ([]string, *stan.Options, error) {
	opts := stan.GetDefaultOptions()

	//if env.Net.TLS.Enable {
	//	kafkaAuthConfig.TLS = &client.KafkaTlsConfig{
	//		Cacert:   env.Net.TLS.CACert,
	//		Usercert: env.Net.TLS.Cert,
	//		Userkey:  env.Net.TLS.Key,
	//	}
	//}
	//
	//if env.Net.SASL.Enable {
	//	kafkaAuthConfig.SASL = &client.KafkaSaslConfig{
	//		User:     env.Net.SASL.User,
	//		Password: env.Net.SASL.Password,
	//		SaslType: env.Net.SASL.Type,
	//	}
	//}
	//
	//cfg, err := client.NewConfigBuilder().
	//	WithDefaults().
	//	WithAuth(kafkaAuthConfig).
	//	WithVersion(&sarama.V2_0_0_0).
	//	Build(ctx)
	//if err != nil {
	//	return nil, nil, fmt.Errorf("error creating Sarama config: %w", err)
	//}

	return env.BootstrapServers, &opts, nil
}
