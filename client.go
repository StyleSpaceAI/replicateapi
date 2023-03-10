package replicateapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Client for the replicate.com api. Use NewClient for smooth initialization
type Client struct {
	AuthorizationToken string
	Owner              string
	Model              string
	Version            string

	HTTPClient *http.Client
}

var (
	// URI of the replicate API
	URI = "https://api.replicate.com"
	// Version of the replicate API
	Version = "v1"
)

func buildURL(path string) string {
	return fmt.Sprintf("%s/%s%s", URI, Version, path)
}

// NewClient creates a new API client
func NewClient(token, model, version string) (*Client, error) {
	splits := strings.Split(model, "/")
	if len(splits) != 2 {
		return nil, errors.New("format of the model name must be owner/modelname")
	}

	return &Client{
		AuthorizationToken: token,
		Owner:              splits[0],
		Model:              splits[1],
		Version:            version,

		HTTPClient: http.DefaultClient,
	}, nil
}

// PredictionResult is a represenation of a single prediction from the replicate API
type PredictionResult struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Urls    struct {
		Get    string `json:"get"`
		Cancel string `json:"cancel"`
	} `json:"urls"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   any                    `json:"started_at"`
	CompletedAt any                    `json:"completed_at"`
	Status      PredictionStatus       `json:"status"`
	Input       map[string]interface{} `json:"input"`
	Output      any                    `json:"output"`
	Error       any                    `json:"error"`
	Logs        any                    `json:"logs"`
	Metrics     map[string]interface{} `json:"metrics"`
}

// CreatePrediction will register an asynchronous prediction request with the replicate API
func (c *Client) CreatePrediction(ctx context.Context, input map[string]interface{}) (*PredictionResult, error) {
	const path = "/predictions"

	type request struct {
		Version string                 `json:"version"`
		Input   map[string]interface{} `json:"input"`
	}

	reqBody := request{
		Version: c.Version,
		Input:   input,
	}
	reqBodyb, err := json.Marshal(reqBody)
	if err != nil {
		return nil, errors.Wrap(err, "encode request")
	}

	req, err := c.newReq(ctx, http.MethodPost, buildURL(path), bytes.NewReader(reqBodyb))
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "create prediction request")
	}
	defer resp.Body.Close()

	if err = handleStatusCode(resp.StatusCode); err != nil {
		return nil, err
	}

	respBody := &PredictionResult{}
	err = json.NewDecoder(resp.Body).Decode(respBody)
	if err != nil {
		return nil, errors.Wrap(err, "decoding the response")
	}
	return respBody, nil
}

// PredictionStatus is returned from replicate API
type PredictionStatus = string

const (
	// PredictionStatusStarting - the prediction is starting up. If this status lasts longer than a few seconds, then it's typically because a new worker is being started to run the prediction.
	PredictionStatusStarting = "starting"
	// PredictionStatusProcessing - the predict() method of the model is currently running.
	PredictionStatusProcessing = "processing"
	// PredictionStatusSucceeded - the prediction completed successfully.
	PredictionStatusSucceeded = "succeeded"
	// PredictionStatusFailed - the prediction encountered an error during processing.
	PredictionStatusFailed = "failed"
	// PredictionStatusCanceled - the prediction was canceled by the user.
	PredictionStatusCanceled = "canceled"
)

// Refresh the status of the prediction inplace
func (p *PredictionResult) Refresh(ctx context.Context, c *Client) error {
	readCloser, err := c.getResult(ctx, p.ID)
	if err != nil {
		return err
	}
	defer readCloser.Close()

	err = json.NewDecoder(readCloser).Decode(p)
	if err != nil {
		return errors.Wrap(err, "decoding the response")
	}
	return nil
}

func (c *Client) getResult(ctx context.Context, id string) (io.ReadCloser, error) {
	const path = "/predictions/"

	req, err := c.newReq(ctx, http.MethodGet, buildURL(path+id), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "create prediction request")
	}

	if err = handleStatusCode(resp.StatusCode); err != nil {
		defer resp.Body.Close()
		return nil, err
	}

	return resp.Body, nil
}

// GetResult of a prediction by its ID
func (c *Client) GetResult(ctx context.Context, predictionID string) (*PredictionResult, error) {
	readCloser, err := c.getResult(ctx, predictionID)
	if err != nil {
		return nil, err
	}
	defer readCloser.Close()

	respBody := &PredictionResult{}
	err = json.NewDecoder(readCloser).Decode(respBody)
	if err != nil {
		return nil, errors.Wrap(err, "decoding the response")
	}
	return respBody, nil
}

// ModelVersion represents a single version of the model with the related schema
type ModelVersion struct {
	ID            string                 `json:"id"`
	CreatedAt     time.Time              `json:"created_at"`
	CogVersion    string                 `json:"cog_version"`
	OpenapiSchema map[string]interface{} `json:"openapi_schema"`
}

// GetModelVersions will return the list of versions available for the model set in the client
// All the versions are sorted by the creation time
func (c *Client) GetModelVersions(ctx context.Context) ([]*ModelVersion, error) {
	path := fmt.Sprintf("/models/%s/%s/versions", c.Owner, c.Model)

	req, err := c.newReq(ctx, http.MethodGet, buildURL(path), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "create prediction request")
	}
	defer resp.Body.Close()

	if err = handleStatusCode(resp.StatusCode); err != nil {
		return nil, err
	}

	type response struct {
		Previous any             `json:"previous"`
		Next     any             `json:"next"`
		Results  []*ModelVersion `json:"results"`
	}

	respBody := &response{}
	err = json.NewDecoder(resp.Body).Decode(respBody)
	if err != nil {
		return nil, errors.Wrap(err, "decode the response")
	}

	sort.Slice(respBody.Results, func(i, j int) bool {
		return respBody.Results[i].CreatedAt.After(respBody.Results[j].CreatedAt)
	})
	return respBody.Results, nil
}

func (c *Client) newReq(ctx context.Context, method string, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, errors.Wrap(err, "init new request")
	}
	req.Header.Add("Authorization", "Token "+c.AuthorizationToken)
	req.Header.Add("Content-Type", "application/json")

	return req, nil
}
