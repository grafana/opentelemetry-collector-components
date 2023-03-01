package faroreceiver

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-logfmt/logfmt"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
)

type Handler struct {
	logConsumer   consumer.Logs
	traceConsumer consumer.Traces
}

var _ http.Handler = (*Handler)(nil)

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var p Payload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.consumePayload(ctx, p); err != nil {
		http.Error(w, "collector error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h Handler) consumePayload(ctx context.Context, p Payload) error {
	logs, err := translate(ctx, p)
	if err != nil {
		return err
	}
	if logs != nil {
		err = h.logConsumer.ConsumeLogs(ctx, *logs)
	}
	if p.Traces != nil {
		err = h.traceConsumer.ConsumeTraces(ctx, p.Traces.Traces)
	}

	return err
}

// kvTime this would be unneeded once KeyVal maps are replaced
// with pcommon.Map types
type kvTime struct {
	kv   *KeyVal
	ts   time.Time
	kind Kind
}

func translate(_ context.Context, p Payload) (*plog.Logs, error) {
	var kvList []*kvTime
	for _, logItem := range p.Logs {
		kvList = append(kvList, &kvTime{
			kv:   logItem.KeyVal(),
			ts:   logItem.Timestamp,
			kind: KindLog,
		})
	}
	for _, exception := range p.Exceptions {

		kvList = append(kvList, &kvTime{
			kv:   exception.KeyVal(),
			ts:   exception.Timestamp,
			kind: KindException,
		})
	}
	for _, measurement := range p.Measurements {
		kvList = append(kvList, &kvTime{
			kv:   measurement.KeyVal(),
			ts:   measurement.Timestamp,
			kind: KindMeasurement,
		})
	}
	for _, event := range p.Events {
		kvList = append(kvList, &kvTime{
			kv:   event.KeyVal(),
			ts:   event.Timestamp,
			kind: KindEvent,
		})
	}

	if len(kvList) == 0 {
		return nil, nil
	}

	logs := plog.NewLogs()
	meta := p.Meta.KeyVal()
	sl := logs.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty()
	attrs := commonAttributes()

	for _, i := range kvList {
		MergeKeyVal(i.kv, meta)
		line, err := logfmt.MarshalKeyvals(KeyValToInterfaceSlice(i.kv)...)
		if err != nil {
			return nil, err
		}
		logRecord := sl.LogRecords().AppendEmpty()
		logRecord.Body().SetStr(string(line))
		attrs.CopyTo(logRecord.Attributes())
		logRecord.Attributes().PutStr("kind", string(i.kind))
	}

	return &logs, nil
}

func commonAttributes() pcommon.Map {
	attrs := pcommon.NewMap()
	labels := []string{"kind"}
	attrs.PutStr("loki.attribute.labels", strings.Join(labels, ","))
	attrs.PutStr("loki.format", "logfmt")
	return attrs
}
