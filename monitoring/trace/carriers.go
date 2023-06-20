// Package trace contains functions for tracing HTTP requests.
package trace

// FUNCTIONS COPIED FROM HERE: https://github.com/ricardo-ch/go-tracing/blob/master/span.go
import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// InjectIntoCarrier returns a textMapCarrier, basically a map[string]string,
// which can be used to transmit a span context to another service with ExtractFromCarrier.
func InjectIntoCarrier(ctx context.Context) opentracing.TextMapCarrier {
	carrier := opentracing.TextMapCarrier{}

	// Retrieve the Span from context
	if span := opentracing.SpanFromContext(ctx); span != nil {
		// We are going to use this span in a client request, so mark as such.
		ext.SpanKindProducer.Set(span)
		// Retrieve tracer
		tracer := opentracing.GlobalTracer()
		// Inject the Span context into the outgoing HTTP Request
		tracer.Inject(
			span.Context(),
			opentracing.TextMap,
			carrier,
		)
	}
	return carrier
}

// ExtractFromCarrier returns a span with context passed  by the carrier.
// ctx should not already have span in it.
func ExtractFromCarrier(ctx context.Context, carrier opentracing.TextMapCarrier, spanName string) (opentracing.Span, context.Context) {
	tracer := opentracing.GlobalTracer()

	wireContext, _ := tracer.Extract(
		opentracing.TextMap,
		carrier,
	)
	span := tracer.StartSpan(spanName, opentracing.FollowsFrom(wireContext))

	// store span in context
	if ctx == nil {
		ctx = context.Background()
	}
	childCtx := opentracing.ContextWithSpan(ctx, span)

	return span, childCtx
}
