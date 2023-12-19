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

const devicesBasePath = "/api/v1/devices"

type AssetsAPI struct {
	client *Client
}

type ExpandDeviceParams struct {
	Manifest     bool
	Properties   bool
	Connectivity bool
}

type DevicesQuery struct {
	PageToken    string
	PageSize     int
	FilterTypeIn []DeviceType
	Expand       ExpandDeviceParams
}

type DeviceByIDQuery struct {
	ID     string
	Expand ExpandDeviceParams
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

type DeviceByIDResponse struct {
	Device Device  `json:"device"`
	Errors []Error `json:"errors"`
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
	const path = devicesBasePath
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
	if expand := a.encodeExpand(query.Expand); len(expand) != 0 {
		values.Set("expand", expand)
	}
	if len(query.FilterTypeIn) != 0 {
		values.Set("filter[type_in]", a.encodeTypes(query.FilterTypeIn))
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

func (a *AssetsAPI) DeviceByID(
	ctx context.Context, query DeviceByIDQuery,
) (DeviceByIDResponse, error) {
	path := devicesBasePath + "/" + query.ID
	req, err := a.client.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return DeviceByIDResponse{}, fmt.Errorf("create request: %w", err)
	}

	values := req.URL.Query()
	if expand := a.encodeExpand(query.Expand); len(expand) != 0 {
		values.Set("expand", expand)
	}
	req.URL.RawQuery = values.Encode()

	resp, err := a.client.Do(req)
	if err != nil {
		return DeviceByIDResponse{}, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	var response DeviceByIDResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return DeviceByIDResponse{}, fmt.Errorf("unmarshal response: %w", err)
	}

	return response, nil
}

func (a *AssetsAPI) encodeExpand(expand ExpandDeviceParams) string {
	var list []string
	if expand.Manifest {
		list = append(list, "manifest")
	}
	if expand.Properties {
		list = append(list, "properties")
	}
	if expand.Connectivity {
		list = append(list, "connectivity")
	}
	return strings.Join(list, ",")
}

func (a *AssetsAPI) encodeTypes(types []DeviceType) string {
	s := make([]string, len(types))
	for i, t := range types {
		s[i] = string(t)
	}
	return strings.Join(s, ",")
}
