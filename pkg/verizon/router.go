package verizon

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/lucastheisen/verizon-router-dyndns/pkg/httpclient"
)

const (
	DefaultRouterHost = "192.168.1.1"
)

type API struct {
	DoSetupWizard            bool   `json:"doSetupWizard"`
	RequirePassword          bool   `json:"requirePassword"`
	PasswordSalt             string `json:"passwordSalt"`
	IsWireless               bool   `json:"isWireless"`
	Error                    int    `json:"error"`
	MaxUsers                 int    `json:"maxUsers"`
	DenyState                int    `json:"denyState"`
	MeshNetworkEnabledStatus bool   `json:"meshNetworkEnabledStatus"`
	MeshUserEnabledConfig    bool   `json:"meshUserEnabledConfig"`
}

type Network struct {
	ConnectionID   int    `json:"connectionId"`
	ConnectionType string `json:"connectionType"`
	Name           string `json:"name"`
	IPAddress      string `json:"ipAddress"`
	IPv6Address    string `json:"ipv6Address"`
	Type           int    `json:"type"`
}

type Router struct {
	api    API
	client *http.Client
	Host   string
	httpclient.HttpClientConfig
	Password string
}

func (r Router) ApiUrl(path string) string {
	host := r.Host
	if host == "" {
		host = DefaultRouterHost
	}
	return fmt.Sprintf("https://%s/api%s", host, path)
}

func (r *Router) Close() {
	r.client.CloseIdleConnections()
	r.client = nil
}

// Connect will create an authenticated session with the router.
// Based upon [this python library].
//
// [this python library]: https://github.com/matray/quantum_gateway_reverse_engineering/blob/master/q_gateway/router.py
func (r *Router) Connect() error {
	var err error
	r.client, err = httpclient.NewHttpClient(r.HttpClientConfig)
	if err != nil {
		return err
	}

	resp, err := r.client.Get(r.ApiUrl(""))
	if err != nil {
		return fmt.Errorf("get api: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// response body comes back even for 401, it should contain the api metadata
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read api: %w", err)
	}

	err = json.Unmarshal(data, &r.api)
	if err != nil {
		return fmt.Errorf("unmarshal api: %w", err)
	}

	if r.api.PasswordSalt == "" {
		return fmt.Errorf("unable to detect password salt: %s", data)
	}

	hashedPwd := sha512.New()
	_, err = hashedPwd.Write([]byte(r.Password + r.api.PasswordSalt))
	if err != nil {
		return fmt.Errorf("hash pwd: %w", err)
	}
	hashedPwdSum := hashedPwd.Sum(nil)
	encodedPwd := make([]byte, hex.EncodedLen(len(hashedPwdSum)))
	hex.Encode(encodedPwd, hashedPwd.Sum(nil))
	loginReq, err := json.Marshal(map[string]string{"password": string(encodedPwd)})
	if err != nil {
		return fmt.Errorf("unmarshal api: %w", err)
	}

	Logger.Trace().Bytes("body", loginReq).Msg("sending login")
	resp, err = r.client.Post(
		r.ApiUrl("/login"),
		"applicaion/json",
		bytes.NewBuffer(loginReq))
	if err != nil {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login (%s): %w", body, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed: %d", resp.StatusCode)
	}
	Logger.Debug().Msg("authentication successful")

	return nil
}

func (r *Router) Network() ([]Network, error) {
	var networks []Network

	var xsrfToken string
	u, _ := url.Parse(r.ApiUrl(""))
	Logger.Trace().Interface("cookies", r.client.Jar.Cookies).Msg("cookies")
	for _, c := range r.client.Jar.Cookies(u) {
		if c.Name == "XSRF-TOKEN" {
			Logger.Debug().Str("XSRF-TOKEN", c.Value).Msg("setting xsrf token")
			xsrfToken = c.Value
		}
	}
	if xsrfToken == "" {
		return networks, errors.New("did not find XSRF-TOKEN")
	}

	req, err := http.NewRequest(http.MethodGet, r.ApiUrl("/network"), nil)
	if err != nil {
		return networks, fmt.Errorf("new request: %w", err)
	}
	req.Header.Add("X-XSRF-TOKEN", xsrfToken)

	resp, err := r.client.Do(req)
	if err != nil {
		return networks, fmt.Errorf("get network: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return networks, fmt.Errorf("network not success: %d", resp.StatusCode)
	}

	// response body comes back even for 401, it should contain the api metadata
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return networks, fmt.Errorf("read network: %w", err)
	}

	err = json.Unmarshal(data, &networks)
	if err != nil {
		return networks, fmt.Errorf("unmarshal network: %w", err)
	}

	return networks, nil
}
