package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type AssetsAPI struct {
	client *Client
}

type DevicesQuery struct {
	PageToken          string
	PageSize           int
	ExpandManifest     bool
	ExpandProperties   bool
	ExpandConnectivity bool
	FilterTypeIn       []DeviceType
	FilterDeviceIDIn   []string
}

type DeviceType string

const (
	DeviceTypeEndpoint DeviceType = "endpoint"
	DeviceTypeUCM      DeviceType = "ucm"
	DeviceTypeGateway  DeviceType = "gateway"
)

type DevicesResponse struct {
	Devices       []Device `json:"devices"`
	Errors        []Error  `json:"errors"`
	NextPageToken string   `json:"next_page_token"`
}

type Device struct {
	DeviceID     string                 `json:"device_id"`
	Type         DeviceType             `json:"type"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Properties   map[string]interface{} `json:"properties,omitempty"`
	Connectivity *DeviceConnectivity    `json:"connectivity,omitempty"`
	Manifest     json.RawMessage        `json:"manifest,omitempty"`
}

type DeviceConnectivity struct {
	Online bool `json:"online"`
}

func (a *AssetsAPI) Devices(
	ctx context.Context, query DevicesQuery,
) (DevicesResponse, error) {
	const path = "/api/v1/devices"
	req, err := a.client.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return DevicesResponse{}, fmt.Errorf("create request: %w", err)
	}

	values := req.URL.Query()
	if query.PageToken != "" {
		values.Set("page_token", query.PageToken)
	}
	if query.PageSize != 0 {
		values.Set("page_size", strconv.Itoa(query.PageSize))
	}
	if expand := a.encodeExpand(query); len(expand) != 0 {
		values.Set("expand", a.encodeExpand(query))
	}
	if len(query.FilterTypeIn) != 0 {
		values.Set("filter[type_in]", a.encodeTypes(query.FilterTypeIn))
	}
	if len(query.FilterDeviceIDIn) != 0 {
		values.Set("filter[device_id_in]",
			strings.Join(query.FilterDeviceIDIn, ","))
	}
	req.URL.RawQuery = values.Encode()

	resp, err := a.client.Do(req)
	if err != nil {
		return DevicesResponse{}, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	var response DevicesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return DevicesResponse{}, fmt.Errorf("unmarshal response: %w", err)
	}

	return response, nil
}

func (a *AssetsAPI) encodeExpand(query DevicesQuery) string {
	var expand []string
	if query.ExpandManifest {
		expand = append(expand, "manifest")
	}
	if query.ExpandProperties {
		expand = append(expand, "properties")
	}
	if query.ExpandConnectivity {
		expand = append(expand, "connectivity")
	}
	return strings.Join(expand, ",")
}

func (a *AssetsAPI) encodeTypes(types []DeviceType) string {
	s := make([]string, len(types))
	for i, t := range types {
		s[i] = string(t)
	}
	return strings.Join(s, ",")
}
