package forecast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"time"
)

// API provides access to the forecastapp.com API
type API struct {
	URL       string
	AccountID string
	token     string       `ignored:"true"`
	client    *http.Client `ignored:"true"`
	timeout   time.Duration
}

// New returns a API that is authenticated with Forecast
func New(url string, accountID string, accessToken string) *API {
	return NewWithTimeout(url, accountID, accessToken,0)
}

// New returns a API that is authenticated with Forecast
func NewWithTimeout(url string, accountID string, accessToken string, timeout time.Duration) *API {
	return &API{
		URL:       url,
		AccountID: accountID,
		token:     accessToken,
		timeout: 	 timeout,
	}
}

func (api *API) fullPath(path string) string {
	return fmt.Sprintf("%s/%s", api.URL, path)
}

func (api *API) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.token))
	req.Header.Set("Forecast-Account-ID", api.AccountID)
	req.Header.Set("content-type", "application/json; charset=utf-8")
}

func (api *API) initializeClient() error {
	if api.client == nil {
		jar, e := cookiejar.New(nil)
		if e != nil {
			return e
		}
		api.client = &http.Client{Jar: jar, Timeout: api.timeout}
	}
	return nil
}

func (api *API) newRequest(method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, api.fullPath(path), body)
	if err != nil {
		return nil, err
	}
	err = api.initializeClient()
	if err != nil {
		return nil, err
	}
	api.setHeaders(req)
	return req, nil
}

func (api *API) decodeResponse(r *http.Response, result interface{}) error {
	if r.StatusCode >= http.StatusBadRequest {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("%s: %s", r.Status, string(body))
	}
	return json.NewDecoder(r.Body).Decode(result)
}

func (api *API) do(path string, result interface{}) error {
	req, err := api.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	resp, err := api.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return api.decodeResponse(resp, result)
}

func (api *API) put(path string, data []byte, result interface{}) error {
	req, err := api.newRequest(http.MethodPut, path, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	resp, err := api.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return api.decodeResponse(resp, result)
}

