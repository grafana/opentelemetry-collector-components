package gcpricingestimatorconnector

import (
	"context"
	"sync"
	"time"
	"unsafe"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)


type sum struct {
	count      uint64
	attributes pcommon.Map
}

func (s *sum) add(v uint64) {
	s.count += v
}

type sumMetrics struct {
	metrics map[string]*sum
}

func newSumMetrics() sumMetrics {
	return sumMetrics{
		metrics: make(map[string]*sum),
	}
}

func (m *sumMetrics) getOrCreate(k string, attributes pcommon.Map) *sum {
	s, ok := m.metrics[k]
	if !ok {
		s = &sum{
			attributes: attributes,
		}
		m.metrics[k] = s
	}
	return s
}


type estimator struct {
  lock sync.Mutex
	logger *zap.Logger

  started bool
	done   chan struct{}
  ticker *time.Ticker

	metrics sumMetrics
	mc      consumer.Metrics
}

var (
	_ connector.Metrics = (*estimator)(nil)
	_ connector.Traces  = (*estimator)(nil)
	_ connector.Logs    = (*estimator)(nil)
)

func newEstimator(logger *zap.Logger, _ *Config) (*estimator, error) {
	logger.Info("Building estimator")
	return &estimator{
		logger:  logger,
		metrics: newSumMetrics(),
	}, nil
}

func (e *estimator) exportMetrics(ctx context.Context) {
  e.lock.Lock()
  defer e.lock.Unlock()
  var pMetrics pmetric.Metrics

  // tada?

  if err := e.mc.ConsumeMetrics(ctx, pMetrics); err != nil {
    e.logger.Error("Failed ConsumeMetrics", zap.Error(err))
  }
}

func (e *estimator) Start(ctx context.Context, host component.Host) error {
  e.logger.Info("Started gcpricingestimator connector")
  e.started = true

  go func() {
    for {
      select {
      case <-e.done:
        return
      case <-ctx.Done():
        return
      case <-e.ticker.C:
        e.exportMetrics(ctx)
      }
    }
  }()

  return nil
}

func (e *estimator) Shutdown(ctx context.Context) error {
  e.logger.Info("Stopping gcpricingestimator connector")
  if e.started {
    e.ticker.Stop() 
    e.done <- struct{}{}
    e.started = false
  }
  return nil
}


func (*estimator) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func (e *estimator) ConsumeMetrics(ctx context.Context, metrics pmetric.Metrics) error {
  e.lock.Lock()
  defer e.lock.Unlock()

	// count active series
  m := e.metrics.getOrCreate("estimator_active_series", pcommon.NewMap())
  m.add(uint64(metrics.MetricCount()))

  return nil
}

func (e *estimator) ConsumeLogs(ctx context.Context, logs plog.Logs) error {
  e.lock.Lock()
  defer e.lock.Unlock()
  m := e.metrics.getOrCreate("estimator_log_bytes", pcommon.NewMap())

  // get byte count of logs (probably VERY inaccurately for now)
  // This iterates over the logs and adds the sum of bytes making up
  // the body (only string types for now).
  for i := 0; i < logs.ResourceLogs().Len(); i++ {
    rl := logs.ResourceLogs().At(i)
    for j := 0; j < rl.ScopeLogs().Len(); j++ {
      sl := rl.ScopeLogs().At(j)
      for k := 0; k < sl.LogRecords().Len(); k++ {
        lr := sl.LogRecords().At(k)
        switch t := lr.Body().Type(); t {
        case pcommon.ValueTypeStr:
          m.add(uint64(len(lr.Body().Str())))
        }
      }
    }
  }
  m.add(uint64(unsafe.Sizeof(logs)))
  
  return nil
}

func (e *estimator) ConsumeTraces(ctx context.Context, traces ptrace.Traces) error {
  e.lock.Lock()
  defer e.lock.Unlock()
  m := e.metrics.getOrCreate("estimator_trace_bytes", pcommon.NewMap())

  // get byte count of traces (probably VERY inaccurately for now)
  // This iterates over the spans and adds the sum of bytes that make
  // up the span attributes and span events (only string types for now).
  for i := 0; i < traces.ResourceSpans().Len(); i++ {
    rs := traces.ResourceSpans().At(i)
    for j := 0; j < rs.ScopeSpans().Len(); j++ {
      ss := rs.ScopeSpans().At(j)
      for k := 0; k < ss.Spans().Len(); k++ {
        span := ss.Spans().At(k)
        sa := span.Attributes()
        for key, val := range sa.AsRaw() {
          m.add(uint64(len(key)))
          vt := val.(pcommon.Value).Type()
          switch vt {
          case pcommon.ValueTypeStr:
            m.add(uint64(len(val.(pcommon.Value).Str())))
          }
        }

        for h := 0; h < span.Events().Len(); h++ {
          event := span.Events().At(h)
          m.add(uint64(len(event.Name())))
          for key, val := range event.Attributes().AsRaw() {
            m.add(uint64(len(key)))
            vt := val.(pcommon.Value).Type()
            switch vt {
            case pcommon.ValueTypeStr:
              m.add(uint64(len(val.(pcommon.Value).Str())))
            }
          }
        }
      }
    }
  }

  return nil
}
