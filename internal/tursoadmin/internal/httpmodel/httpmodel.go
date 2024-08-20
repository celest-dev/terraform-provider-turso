package httpmodel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing/iotest"
)

type JSONClient struct {
	*http.Client
	baseURL string
	headers map[string]string
}

func NewJSONClient(client *http.Client, baseURL string, headers map[string]string) *JSONClient {
	return &JSONClient{Client: client, baseURL: baseURL, headers: headers}
}

type JSON[R any] struct {
	method   string
	url      string
	request  io.Reader
	response R
}

func NewJSON[R any](method string, url string, request interface{}) *JSON[R] {
	model := JSON[R]{
		method: method,
		url:    url,
	}
	body, err := json.Marshal(request)
	if err != nil {
		model.request = iotest.ErrReader(
			fmt.Errorf("error marshalling body (%T): %v", request, err))
	} else {
		model.request = bytes.NewReader(body)
	}
	return &model
}

func (m *JSON[R]) Send(ctx context.Context, client *JSONClient) (R, error) {
	req, err := http.NewRequestWithContext(ctx, m.method, client.baseURL+m.url, m.request)
	if err != nil {
		return m.response, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for key, value := range client.headers {
		req.Header.Set(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		return m.response, fmt.Errorf("error sending request: %v", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return m.response, fmt.Errorf("unexpected status code: %v", resp.Status)
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&m.response)
	if err != nil {
		if err == io.EOF {
			v := reflect.ValueOf(m.response)
			if v.Kind() == reflect.Struct && v.NumField() == 0 {
				return m.response, nil
			}
			return m.response, errors.New("empty response")
		}
		return m.response, fmt.Errorf("error reading response: %v", err)
	}
	return m.response, nil
}
