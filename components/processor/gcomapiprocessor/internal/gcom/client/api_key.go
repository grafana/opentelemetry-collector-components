package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-kit/log/level"
)

var (
	// ErrAccess represents errors connecting to Grafana Cloud
	ErrAccess = errors.New("unable to access grafana.com")

	// ErrInvalidKey respresents errors were the key being checked is not valid
	ErrInvalidKey = errors.New("invalid api key")
)

// APIKey represents a Grafana Cloud API Key
type APIKey struct {
	ID        int         `json:"id"`
	OrgID     int         `json:"orgId"`
	OrgSlug   string      `json:"orgSlug"`
	OrgName   string      `json:"orgName"`
	Name      string      `json:"name"`
	Role      RoleType    `json:"role"`
	CreatedAt time.Time   `json:"createdAt"`
	UpdatedAt interface{} `json:"updatedAt"`
	FirstUsed time.Time   `json:"firstUsed"`
	Links     []struct {
		Rel  string `json:"rel"`
		Href string `json:"href"`
	} `json:"links"`
}

// IsAdminOrg is used to determine if the APIKey is part of the admin org
func (a *APIKey) IsAdminOrg() bool {
	return (a.OrgID == 1)
}

// CheckAPIKey verifies a GrafanaCloud api key and returns a API Key struct
func (c *client) CheckAPIKey(ctx context.Context, key string) (*APIKey, error) {
	payload := url.Values{}
	payload.Add("token", key)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		strings.TrimRight(c.endpoint.String(), "/")+"/api-keys/check",
		strings.NewReader(payload.Encode()),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("err=%w, msg=%v", ErrAccess, err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		apiKey, err := apiKeyFromJSON(res.Body)
		if err != nil {
			return nil, fmt.Errorf("err=%w, msg=%v", ErrAccess, err)
		}

		return apiKey, nil
	}

	// For unsuccessful requests log the message body
	// Note that the message can be "Invalid token" for valid tokens; this is done for security.
	msg, err := io.ReadAll(res.Body)
	if err != nil {
		level.Error(c.logger).Log("msg", "invalid response when checking key", "status", res.StatusCode, "err", err)
	} else {
		level.Error(c.logger).Log("msg",
			"invalid response when checking key",
			"status",
			res.StatusCode,
			"errMsg",
			string(msg),
			"err",
			err)
	}

	switch res.StatusCode {
	case http.StatusConflict:
		// Invalid keys return a 409 error code
		return nil, ErrInvalidKey
	default:
		level.Warn(c.logger).Log("msg", "error for api check", "status", res.StatusCode)
		return nil, ErrAccess
	}
}

func apiKeyFromJSON(body io.Reader) (*APIKey, error) {
	var apiKey APIKey
	if err := json.NewDecoder(body).Decode(&apiKey); err != nil {
		return nil, err
	}
	return &apiKey, nil
}
