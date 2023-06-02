package client

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_client_ListClusters(t *testing.T) {
	c, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/hm-clusters", func(w http.ResponseWriter, r *http.Request) {
		slugIn := r.URL.Query()["slugIn"]
		if len(slugIn) == 1 && slugIn[0] == "alertmanager-dev-us-central-0" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, sampleHMClusterResponse)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	})

	mux.HandleFunc("/hg-clusters", func(w http.ResponseWriter, r *http.Request) {
		slugIn := r.URL.Query()["slugIn"]
		if len(slugIn) == 1 && slugIn[0] == "dev-us-central-0" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, sampleHGClusterResponse)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	})

	t.Run("lookup clusters by single slug", func(t *testing.T) {
		got, err := c.ListClusters(context.Background(), ClusterRequestOptions{
			Slugs: []string{"alertmanager-dev-us-central-0"},
		})
		require.NoError(t, err)
		assert.Len(t, got, 1)
		assert.Equal(t, got[0].ID, sampleHMClusterParsed.ID)
	})

	t.Run("lookup grafana clusters by single slug", func(t *testing.T) {
		got, err := c.ListClusters(context.Background(), ClusterRequestOptions{
			Slugs: []string{"dev-us-central-0"},
			Type:  Grafana,
		})
		require.NoError(t, err)
		assert.Len(t, got, 1)
		assert.Equal(t, got[0].ID, sampleHGClusterParsed.ID)
	})

	t.Run("lookup unknown cluster", func(t *testing.T) {
		_, err := c.ListClusters(context.Background(), ClusterRequestOptions{
			Slugs: []string{"nope"},
		})
		require.Error(t, err)
	})
}
