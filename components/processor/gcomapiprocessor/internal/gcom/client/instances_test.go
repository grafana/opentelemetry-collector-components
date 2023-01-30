package client

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func compareInstances(got, wanted Instance) error {
	if got.ID != wanted.ID {
		return fmt.Errorf("incorrect instance ID, got(%v) != expected(%v)", got.ID, wanted.ID)
	}
	if got.OrgID != wanted.OrgID {
		return fmt.Errorf("incorrect org ID, got(%v) != expected(%v)", got.OrgID, wanted.OrgID)
	}
	if got.Type != wanted.Type {
		return fmt.Errorf("incorrect instance type, got(%v) != expected(%v)", got.Type, wanted.Type)
	}
	if got.ClusterID != wanted.ClusterID {
		return fmt.Errorf("incorrect cluster ID, got(%v) != expected(%v)", got.ClusterID, wanted.ClusterID)
	}
	if got.Name != wanted.Name {
		return fmt.Errorf("incorrect instance name, got(%v) != expected(%v)", got.Name, wanted.Name)
	}
	return nil
}

func Test_client_ListInstances(t *testing.T) {
	c, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/hosted-logs", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, sampleLogResponse)
	})

	mux.HandleFunc("/hosted-metrics", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		typ := r.Form.Get("type")
		sep := ""

		w.WriteHeader(http.StatusOK)
		// header
		fmt.Fprint(w, `{"items": [`)

		if typ == Prometheus.String() || typ == "" {
			fmt.Fprint(w, hostedPrometheusInstance)
			sep = ","
		}
		if typ == Graphite.String() || typ == "" {
			fmt.Fprint(w, sep)
			fmt.Fprint(w, hostedGraphiteInstance)
		}

		// footer.
		fmt.Fprint(w, `], "orderBy": "name", "direction": "asc", "links": [{ "rel": "self", "href": "/hosted-metrics" }]}`)
	})

	mux.HandleFunc("/hosted-alerts", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, sampleAlertsResponse)
	})

	mux.HandleFunc("/instances", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, sampleGrafanaResponse)
	})

	tests := []struct {
		name    string
		options InstanceRequestOptions
		want    []Instance
		wantErr bool
	}{
		{
			name: "hosted metrics (prometheus) instances",
			options: InstanceRequestOptions{
				Type: Prometheus,
			},
			want: []Instance{samplePrometheusInstanceParsed},
		},
		{
			name: "hosted metrics (graphite) instances",
			options: InstanceRequestOptions{
				Type: Graphite,
			},
			want: []Instance{sampleGraphiteInstanceParsed},
		},
		{
			name: "hosted metrics (all) instances",
			options: InstanceRequestOptions{
				Type: Metrics,
			},
			want: []Instance{samplePrometheusInstanceParsed, sampleGraphiteInstanceParsed},
		},
		{
			name: "hosted logs instances",
			options: InstanceRequestOptions{
				Type: Logs,
			},
			want: sampleLogsParsed,
		},
		{
			name: "hosted alerts instances",
			options: InstanceRequestOptions{
				Type: Alerts,
			},
			want: sampleAlertsParsed,
		},
		{
			name: "grafana instances",
			options: InstanceRequestOptions{
				Type: Grafana,
			},
			want: sampleGrafanaParsed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.ListInstances(context.Background(), tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("client.ListInstances() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("client.ListInstances() returned the wrong number of instances, %d != %d",
					len(got),
					len(tt.want))
				return
			}

			for i := range got {
				err := compareInstances(got[i], tt.want[i])
				if err != nil {
					t.Errorf("client.ListInstances() returned the incorrect instance, %v", err)
					return
				}
			}
		})
	}
}

func Test_client_ListInstancesWithPagination(t *testing.T) {
	c, mux, teardown := setup()
	defer teardown()

	// Test using non-metrics instances since that is the default, and we
	// want to ensure the default behavior doesn't assist in making the test pass.
	mux.HandleFunc("/hosted-alerts", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)

		cursor := r.Form.Get("cursor")
		if cursor == "1" {
			fmt.Fprint(w, `{"items": [`)
			fmt.Fprint(w, `{ "id": 1 }`)
			fmt.Fprint(w, `]}`)
		} else if cursor == "2" {
			fmt.Fprint(w, `{"items": [`)
			fmt.Fprint(w, `{ "id": 2 }`)
			fmt.Fprint(w, `]}`)
		} else {
			fmt.Fprint(w, `{"items": []}`)
		}
	})

	opts := InstanceRequestOptions{Type: Alerts, PageSize: 1}
	instances, err := c.ListInstancesWithPagination(context.Background(), opts)
	require.NoError(t, err)
	require.Equal(t, 2, len(instances))
}
