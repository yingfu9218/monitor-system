package reporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/monitor-system/internal/server/model"
)

type Reporter struct {
	endpoint string
	agentKey string
	client   *http.Client
}

func New(endpoint, agentKey string) *Reporter {
	return &Reporter{
		endpoint: endpoint,
		agentKey: agentKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (r *Reporter) Report(report *model.AgentReport) error {
	data, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	req, err := http.NewRequest("POST", r.endpoint+"/api/v1/agent/report", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Agent-Key", r.agentKey)

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}
