package httpclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"

	"github.com/rs/zerolog"
)

type HttpClientConfig struct {
	CACert             string
	InsecureSkipVerify bool
}

type Transport struct {
	http.Transport
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if Logger.GetLevel() <= zerolog.TraceLevel {
		cookies := zerolog.Dict()
		for _, cookie := range req.Cookies() {
			cookies = cookies.Str(cookie.Name, cookie.Value)
		}
		headers := zerolog.Dict()
		for name, values := range req.Header {
			headers = headers.Strs(name, values)
		}
		Logger.Trace().
			Str("method", req.Method).
			Str("url", req.URL.String()).
			Dict("headers", headers).
			Dict("cookies", cookies).
			Msg("request")
	}

	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	if Logger.GetLevel() <= zerolog.TraceLevel {
		cookies := zerolog.Dict()
		for _, cookie := range resp.Cookies() {
			cookies = cookies.Str(cookie.Name, cookie.Value)
		}
		Logger.Trace().Dict("cookies", cookies).Msg("response")
	}

	return resp, err
}

func NewHttpClient(config HttpClientConfig) (*http.Client, error) {
	tr := &Transport{
		http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: config.InsecureSkipVerify,
			},
		},
	}

	if config.CACert != "" {
		Logger.Info().Str("CACert", config.CACert).Msg("setting explicit CA certificate")
		certPool := x509.NewCertPool()
		pem, err := os.ReadFile(config.CACert)
		if err != nil {
			return nil, fmt.Errorf("read root ca: %w", err)
		}
		certPool.AppendCertsFromPEM(pem)
		tr.TLSClientConfig.RootCAs = certPool
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("cookiejar: %w", err)
	}

	return &http.Client{
		Jar:       jar,
		Transport: tr,
	}, nil
}
