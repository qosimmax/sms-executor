package trace

import (
	"fmt"
	"io"

	"github.com/qosimmax/sms-executor/config"

	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
)

// InitGlobalTracer creates the global tracer object.
func InitGlobalTracer(config *config.Config) (io.Closer, error) {
	cfg := jaegercfg.Configuration{
		Reporter: &jaegercfg.ReporterConfig{
			LocalAgentHostPort: fmt.Sprintf("%s:%s", config.JaegerAgentHost, config.JaegerAgentPort),
		},
		Sampler: &jaegercfg.SamplerConfig{
			Type:  config.JaegerSamplerType,
			Param: config.JaegerSamplerParam,
		},
	}

	jLogger := jaegerlog.StdLogger

	closer, err := cfg.InitGlobalTracer(
		fmt.Sprintf("sms-executor.%s", config.NatsTopic),
		jaegercfg.Logger(jLogger),
	)

	if err != nil {
		return nil, err
	}

	return closer, nil
}
