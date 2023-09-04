package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type CommandsAPI struct {
	client *Client
}

type CommandQuery struct {
	DeviceID    string                 `json:"device_id,omitempty"`
	HardwareID  string                 `json:"hardware_id,omitempty"`
	CommandName string                 `json:"command_name"`
	Arguments   map[string]interface{} `json:"arguments,omitempty"`
}

type CommandResponse struct {
	State   CommandState           `json:"state"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

type CommandState string

const (
	CommandSucceeded     CommandState = "succeeded"
	CommandError         CommandState = "error"
	CommandPlatformError CommandState = "platform_error"
	CommandStarted       CommandState = "started"
	CommandInProgress    CommandState = "device_in_progress"
)

func (c *CommandsAPI) Execute(
	ctx context.Context, query CommandQuery,
) (CommandResponse, error) {
	showProgress := false
	resp, err := c.execute(ctx, query, showProgress)
	if err != nil {
		return CommandResponse{}, err
	}
	defer resp.Body.Close()

	var cmdResp CommandResponse
	if err := json.NewDecoder(resp.Body).Decode(&cmdResp); err != nil {
		return CommandResponse{}, fmt.Errorf("unmarshal response: %w", err)
	}

	return cmdResp, nil
}

type CommandProgress struct {
	CommandResponse
	Error error
}

func (c *CommandsAPI) ExecuteWithProgress(
	ctx context.Context, query CommandQuery,
) (<-chan CommandProgress, error) {
	showProgress := true
	//nolint:bodyclose // closed in the reading goroutine
	resp, err := c.execute(ctx, query, showProgress)
	if err != nil {
		return nil, err
	}

	progressCh := make(chan CommandProgress)
	go func() {
		defer resp.Body.Close()
		defer close(progressCh)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			var p CommandProgress
			p.Error = json.Unmarshal(scanner.Bytes(), &p.CommandResponse)

			select {
			case <-ctx.Done():
				return
			case progressCh <- p:
			}
		}
	}()

	return progressCh, nil
}

func (c *CommandsAPI) execute(
	ctx context.Context, query CommandQuery, showProgress bool,
) (*http.Response, error) {
	queryBody := new(bytes.Buffer)
	if err := json.NewEncoder(queryBody).Encode(query); err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	const path = "/commands/v1/execute"
	req, err := c.client.NewRequestWithContext(ctx, http.MethodPost, path, queryBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if showProgress {
		values := req.URL.Query()
		values.Set("show_progress", "true")
		req.URL.RawQuery = values.Encode()
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
