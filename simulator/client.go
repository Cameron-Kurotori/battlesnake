package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Cameron-Kurotori/battlesnake/sdk"
)

type BattlesnakeClient interface {
	Info() (info sdk.BattlesnakeInfoResponse, err error)
	Start(state sdk.GameState) error
	End(state sdk.GameState) error
	Move(state sdk.GameState) (sdk.BattlesnakeMoveResponse, error)
}

type client struct {
	host   string
	port   string
	client *http.Client
}

func NewClient(host, port string) BattlesnakeClient {
	return &client{
		host:   host,
		port:   port,
		client: http.DefaultClient,
	}
}

func (c *client) request(uri string, method string, body []byte) ([]byte, *http.Response, error) {
	r, err := http.NewRequest(method, fmt.Sprintf("http://%s:%s"+uri, c.host, c.port), bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}
	resp, err := c.client.Do(r)
	if err != nil {
		return nil, resp, err
	}

	var responseBody []byte
	if resp.Body != nil {
		responseBody, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, resp, err
		}

		resp.Body = ioutil.NopCloser(bytes.NewReader(responseBody))
	}

	return responseBody, resp, nil
}

func (c *client) Info() (info sdk.BattlesnakeInfoResponse, err error) {
	body, resp, err := c.request("/", http.MethodGet, nil)
	if err != nil {
		return info, err
	}
	if resp.StatusCode >= 300 {
		return info, fmt.Errorf("non successful code received status_code=%d response_body=%s", resp.StatusCode, string(body))
	}

	err = json.Unmarshal(body, &info)
	if err != nil {
		return info, err
	}

	return info, nil
}

func (c *client) Start(state sdk.GameState) error {
	reqBody, err := json.Marshal(state)
	if err != nil {
		return err
	}

	body, resp, err := c.request("/start", http.MethodGet, reqBody)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("non successful code received status_code=%d response_body=%s", resp.StatusCode, string(body))
	}

	return nil

}

func (c *client) End(state sdk.GameState) error {
	reqBody, err := json.Marshal(state)
	if err != nil {
		return err
	}

	body, resp, err := c.request("/end", http.MethodGet, reqBody)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("non successful code received status_code=%d response_body=%s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *client) Move(state sdk.GameState) (move sdk.BattlesnakeMoveResponse, err error) {
	reqBody, err := json.Marshal(state)
	if err != nil {
		return move, err
	}

	body, resp, err := c.request("/move", http.MethodGet, reqBody)
	if err != nil {
		return move, err
	}
	if resp.StatusCode >= 300 {
		return move, fmt.Errorf("non successful code received status_code=%d response_body=%s", resp.StatusCode, string(body))
	}

	err = json.Unmarshal(body, &move)
	if err != nil {
		return move, err
	}

	return move, nil
}
