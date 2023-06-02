package mock

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/grafana/opentelemetry-collector-components/processor/gcomapiprocessor/internal/gcom/client"
)

type Config struct {
	Keys             map[string]client.APIKey `json:"keys"`
	MetricsInstances []client.Instance        `json:"hosted_metrics"`
	LogsInstances    []client.Instance        `json:"hosted_logs"`
	AlertsInstances  []client.Instance        `json:"hosted_alerts"`
	GrafanaInstances []client.Instance        `json:"instances"`
}

var (
	// This is the error returned if a token is not included in the payload
	errNoToken = `{
	"code": "InvalidArgument",
	"message": "Field is required: token"
}`
	errTokenInvalid = `{
	"code": "InvalidArgument",
	"message": "invalid token"
}`
)

// Server is used mock the grafana.com API
type Server struct {
	cfg Config
}

func NewServer(cfg Config) *Server {
	return &Server{
		cfg: cfg,
	}
}

func (s *Server) RegisterRoutes(r *mux.Router) {
	r.Path("/api/api-keys/check").Methods("POST").HandlerFunc(s.apiKeysCheck)
	r.Path("/api/hosted-metrics").Methods("GET").HandlerFunc(s.hostedMetricsInstances)
	r.Path("/api/hosted-logs").Methods("GET").HandlerFunc(s.hostedLogsInstances)
	r.Path("/api/hosted-alerts").Methods("GET").HandlerFunc(s.hostedAlertsInstances)
	r.Path("/api/instances").Methods("GET").HandlerFunc(s.grafanaInstances)
}

func (s *Server) apiKeysCheck(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token := r.PostForm.Get("token")

	if token == "" {
		w.WriteHeader(http.StatusConflict)
		_, err := w.Write([]byte(errNoToken))
		if err != nil {
			log.Println(err)
		}
		return
	}

	k, exists := s.cfg.Keys[token]

	if !exists {
		w.WriteHeader(http.StatusConflict)
		_, err := w.Write([]byte(errTokenInvalid))
		if err != nil {
			log.Println(err)
			return
		}
		return
	}

	payload, err := json.Marshal(k)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) hostedMetricsInstances(w http.ResponseWriter, r *http.Request) {
	matcher := parseOptions(r.URL.Query())

	instances := client.InstanceResponse{
		Items: []client.Instance{},
	}

	for _, i := range s.cfg.MetricsInstances {
		if matcher.matchInstance(i) {
			instances.Items = append(instances.Items, i)
		}
	}

	payload, err := json.Marshal(instances)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) hostedLogsInstances(w http.ResponseWriter, r *http.Request) {
	matcher := parseOptions(r.URL.Query())

	instances := client.InstanceResponse{
		Items: []client.Instance{},
	}

	for _, i := range s.cfg.LogsInstances {
		if matcher.matchInstance(i) {
			instances.Items = append(instances.Items, i)
		}
	}

	payload, err := json.Marshal(instances)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) hostedAlertsInstances(w http.ResponseWriter, r *http.Request) {
	matcher := parseOptions(r.URL.Query())

	instances := client.InstanceResponse{
		Items: []client.Instance{},
	}

	for _, i := range s.cfg.AlertsInstances {
		if matcher.matchInstance(i) {
			instances.Items = append(instances.Items, i)
		}
	}

	payload, err := json.Marshal(instances)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) grafanaInstances(w http.ResponseWriter, r *http.Request) {
	matcher := parseOptions(r.URL.Query())

	instances := client.InstanceResponse{
		Items: []client.Instance{},
	}

	for _, i := range s.cfg.GrafanaInstances {
		if matcher.matchInstance(i) {
			instances.Items = append(instances.Items, i)
		}
	}

	payload, err := json.Marshal(instances)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type instanceMatcher struct {
	clusterSlug   string
	orgIDs        []string
	orgSlugs      []string
	instanceType  string
	instanceNames []string
	instanceIDs   []string
}

func (m *instanceMatcher) matchInstance(i client.Instance) bool {
	if len(m.instanceIDs) != 0 && !searchStringSlice(m.instanceIDs, strconv.Itoa(i.ID)) {
		return false
	}

	if len(m.instanceNames) != 0 && !searchStringSlice(m.instanceNames, i.Name) {
		return false
	}

	if len(m.orgSlugs) != 0 && !searchStringSlice(m.orgSlugs, i.OrgSlug) {
		return false
	}

	if len(m.orgIDs) != 0 && !searchStringSlice(m.orgIDs, strconv.Itoa(i.OrgID)) {
		return false
	}

	if m.instanceType != "" && i.Type.String() != m.instanceType {
		return false
	}

	if m.clusterSlug != "" && i.ClusterSlug != m.clusterSlug {
		return false
	}

	return true
}

func parseOptions(vals url.Values) instanceMatcher {
	matcher := instanceMatcher{}
	ids, exists := vals["id"]
	if exists {
		matcher.instanceIDs = ids
	}

	orgIDs, exists := vals["orgId"]
	if exists {
		matcher.orgIDs = orgIDs
	}

	orgSlugs, exists := vals["orgSlug"]
	if exists {
		matcher.orgSlugs = orgSlugs
	}

	instanceNames, exists := vals["name"]
	if exists {
		matcher.instanceNames = instanceNames
	}

	instanceType := vals.Get("type")
	if exists {
		matcher.instanceType = instanceType
	}

	clusterSlug := vals.Get("clusterSlug")
	if exists {
		matcher.clusterSlug = clusterSlug
	}

	return matcher
}

func searchStringSlice(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
