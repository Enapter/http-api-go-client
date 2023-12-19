package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type BlueprintsAPI struct {
	client *Client
}

func (b *BlueprintsAPI) Download(
	ctx context.Context, blueprintID string,
) ([]byte, error) {
	const path = "/blueprints/v1/download"
	req, err := b.client.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	query := req.URL.Query()
	query.Set("blueprint_id", blueprintID)
	req.URL.RawQuery = query.Encode()

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	blueprintBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	return blueprintBytes, nil
}
