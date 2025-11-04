package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"smoldungeon-cli/types"
)

// Client handles API communication with the game server
type Client struct {
	BaseURL string
	client  *http.Client
}

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		client:  &http.Client{},
	}
}

// CreateSession creates a new game session
func (c *Client) CreateSession(scenarioName string) (*types.CreateSessionResponse, error) {
	payload := map[string]string{}
	if scenarioName != "" {
		payload["scenarioName"] = scenarioName
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Post(c.BaseURL+"/sessions", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result types.CreateSessionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// ApplyAction sends an action to the server and returns the updated state
func (c *Client) ApplyAction(sessionID string, action types.Action) (*types.ActionResponse, error) {
	payload := map[string]interface{}{
		"sessionId": sessionID,
		"action":    action,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Post(c.BaseURL+"/tools/apply_action", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result types.ActionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w\nBody: %s", err, string(body))
	}

	return &result, nil
}

// GetScenarios fetches available scenarios
func (c *Client) GetScenarios() ([]string, error) {
	resp, err := c.client.Get(c.BaseURL + "/scenarios")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var scenarios []string
	if err := json.Unmarshal(body, &scenarios); err != nil {
		return nil, fmt.Errorf("failed to parse scenarios: %w", err)
	}

	return scenarios, nil
}
