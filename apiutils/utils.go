package apiutils

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// ServiceVersion holds the version of a servie
type ServiceVersion struct {
	Libs    map[string]string
	Version string
	Sha     string
}

// GetServiceVersions returns the version of the services.
func GetServiceVersions(api string, tlsConfig *tls.Config) (map[string]ServiceVersion, error) {

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/_meta/version", api), nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Bad response status: %s", resp.Status)
	}

	out := map[string]ServiceVersion{}

	defer resp.Body.Close() // nolint: errcheck
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return out, nil
}

// GetPublicCA returns the public CA used by the api.
func GetPublicCA(api string, tlsConfig *tls.Config) ([]byte, error) {

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/_meta/ca", api), nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Bad response status: %s", resp.Status)
	}

	defer resp.Body.Close() // nolint: errcheck
	return ioutil.ReadAll(resp.Body)
}

// GetPublicCAPool returns the public CA used by the api as a *x509.CertPool.
func GetPublicCAPool(api string, tlsConfig *tls.Config) (*x509.CertPool, error) {

	cadata, err := GetPublicCA(api, tlsConfig)
	if err != nil {
		return nil, err
	}

	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}

	pool.AppendCertsFromPEM(cadata)

	return pool, nil
}

// GetManifestURL returns the url of the manifest.
func GetManifestURL(api string, tlsConfig *tls.Config) ([]byte, error) {

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/_meta/manifest", api), nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Bad response status: %s", resp.Status)
	}

	defer resp.Body.Close() // nolint: errcheck
	return ioutil.ReadAll(resp.Body)
}

// GetGoogleOAuthClientID returns the Google oauth client ID used bby the platform.
func GetGoogleOAuthClientID(api string, tlsConfig *tls.Config) ([]byte, error) {

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/_meta/googleclientid", api), nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Bad response status: %s", resp.Status)
	}

	defer resp.Body.Close() // nolint: errcheck
	return ioutil.ReadAll(resp.Body)
}
