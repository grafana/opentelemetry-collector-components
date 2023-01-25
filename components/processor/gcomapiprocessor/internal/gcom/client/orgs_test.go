package client

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func compareOrgs(got, wanted Org) error {
	if got.ID != wanted.ID {
		return fmt.Errorf("incorrect ID, got(%v) != expected(%v)", got.ID, wanted.ID)
	}
	if got.Slug != wanted.Slug {
		return fmt.Errorf("incorrect slug, got(%v) != expected(%v)", got.Slug, wanted.Slug)
	}
	if got.Name != wanted.Name {
		return fmt.Errorf("incorrect name, got(%v) != expected(%v)", got.Name, wanted.Name)
	}
	if got.GrafanaCloudType != wanted.GrafanaCloudType {
		return fmt.Errorf("incorrect grafana cloud type, got(%v) != expected(%v)",
			got.GrafanaCloudType,
			wanted.GrafanaCloudType)
	}
	if got.ContractType != wanted.ContractType {
		return fmt.Errorf("incorrect contract type, got(%v) != expected(%v)", got.ContractType, wanted.ContractType)
	}
	if got.MetricsUsage != wanted.MetricsUsage {
		return fmt.Errorf("incorrect metrics usage, got(%v) != expected(%v)", got.MetricsUsage, wanted.MetricsUsage)
	}
	if got.MetricsOverageAmount != wanted.MetricsOverageAmount {
		return fmt.Errorf("incorrect metrics overage amount, got(%v) != expected(%v)",
			got.MetricsOverageAmount,
			wanted.MetricsOverageAmount)
	}
	if got.MetricsIncludedSeries != wanted.MetricsIncludedSeries {
		return fmt.Errorf("incorrect metrics included series, got(%v) != expected(%v)",
			got.MetricsIncludedSeries,
			wanted.MetricsIncludedSeries)
	}
	if got.LogsUsage != wanted.LogsUsage {
		return fmt.Errorf("incorrect logs usage, got(%v) != expected(%v)", got.LogsUsage, wanted.LogsUsage)
	}
	if got.LogsOverageAmount != wanted.LogsOverageAmount {
		return fmt.Errorf("incorrect logs overage amount, got(%v) != expected(%v)",
			got.LogsOverageAmount,
			wanted.LogsOverageAmount)
	}
	if got.LogsIncludedUsage != wanted.LogsIncludedUsage {
		return fmt.Errorf("incorrect logs included usage, got(%v) != expected(%v)",
			got.LogsIncludedUsage,
			wanted.LogsIncludedUsage)
	}

	return nil
}

func Test_client_ListOrgs(t *testing.T) {
	c, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/orgs", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, sampleOrgsResponse)
	})

	tests := []struct {
		name    string
		options *OrgRequestOptions
		want    []Org
		wantErr bool
	}{
		{
			name:    "orgs",
			options: &OrgRequestOptions{},
			want:    sampleOrgsParsed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.ListOrgs(context.Background(), tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("client.ListOrgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("client.ListOrgs() returned the wrong number of orgs, %d != %d", len(got), len(tt.want))
				return
			}

			for i := range got {
				err := compareOrgs(got[i], tt.want[i])
				if err != nil {
					t.Errorf("client.ListOrgs() returned the incorrect org, %v", err)
					return
				}
			}
		})
	}
}
