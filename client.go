package adsbexchange

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	Namespace          = "feed.adsbexchange"
	headerContentType  = "Content-Type"
	headerRapidAPIKey  = "X-RapidAPI-Key"
	headerRapidAPIHost = "X-RapidAPI-Host"
)

// BindEnvs registers environment variable mappings for this feed's config namespace.
func BindEnvs(v *viper.Viper) {
	_ = v.BindEnv("feed.adsbexchange.host", "FEED_ADSBEXCHANGE_HOST")
	_ = v.BindEnv("feed.adsbexchange.key", "FEED_ADSBEXCHANGE_KEY")
}

// Client represents a client for the ADS-B Exchange API
type Client struct {
	logger *zap.Logger

	URL  string
	Host string
	Key  string
}

// NewClient creates a new ADS-B Exchange client
func NewClient(logger *zap.Logger, cfg *Config) (*Client, error) {
	if cfg.Host == "" || cfg.Key == "" {
		return nil, errors.New("adsbexchange host and key must be set")
	}

	return &Client{
		logger: logger.Named(Namespace),
		Host:   cfg.Host,
		Key:    cfg.Key,
		URL:    fmt.Sprintf("https://%s/v2", cfg.Host),
	}, nil
}

func (c *Client) Endpoints() map[string]string {
	return map[string]string{
		"aircraft_by_callsign":     "/callsign/{callsign}/",
		"aircraft_by_hex":          "/hex/{hex}",
		"aircraft_by_icao":         "/icao/{hex}/",
		"aircraft_by_registration": "/registration/{registration}/",
		"aircraft_by_squawk":       "/sqk/{squawk}/",
		"aircraft_military":        "/mil/",
		"aircraft_within_range":    "/lat/{lat}/lon/{lon}/dist/{dist}/",
	}
}

// doRequest builds, executes, and reads the body for a given path.
func (c *Client) doRequest(path string) ([]byte, error) {
	req, err := c.getRequest(path)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("making request", zap.String("url", req.URL.String()))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.logger.Error("error making request", zap.Error(err))
		return nil, err
	}
	defer closeResponseBody(resp.Body, c.logger)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("error reading response body", zap.Error(err))
		return nil, err
	}

	return body, nil
}

// fetchResult executes a request and unmarshals the body into a Result,
// returning the filtered aircraft slice.
func (c *Client) fetchResult(path string) ([]*Aircraft, error) {
	body, err := c.doRequest(path)
	if err != nil {
		return nil, err
	}

	var result Result
	if err := json.Unmarshal(body, &result); err != nil {
		c.logger.Error("error unmarshalling response body", zap.Error(err))
		return nil, err
	}

	return c.filterADSBICAO(result.Aircraft), nil
}

// fetchAircraft executes a request and unmarshals the body into a single Aircraft.
func (c *Client) fetchAircraft(path string) (*Aircraft, error) {
	body, err := c.doRequest(path)
	if err != nil {
		return nil, err
	}

	var aircraft Aircraft
	if err := json.Unmarshal(body, &aircraft); err != nil {
		c.logger.Error("error unmarshalling response body", zap.Error(err))
		return nil, err
	}

	return &aircraft, nil
}

// GetMilitaryAircraft returns all military aircraft
func (c *Client) GetMilitaryAircraft() ([]*Aircraft, error) {
	req, err := c.getRequest(c.Endpoints()["aircraft_military"])
	if err != nil {
		return nil, err
	}

	c.logger.Debug("making request", zap.String("url", req.URL.String()))
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		c.logger.Error("error making request", zap.Error(err))
		return nil, err
	}
	defer closeResponseBody(response.Body, c.logger)
	body, err := io.ReadAll(response.Body)
	if err != nil {
		c.logger.Error("error reading response body", zap.Error(err))
	}
	var result Result
	if err := json.Unmarshal(body, &result); err != nil {
		c.logger.Error("error unmarshaling response body", zap.Error(err))
		return nil, err
	}
	return c.filterADSBICAO(result.Aircraft), nil
}

// GetAircraftBySquawk returns aircraft by squawk
func (c *Client) GetAircraftBySquawk(squawk int) ([]*Aircraft, error) {
	path := c.Endpoints()["aircraft_by_squawk"]
	path = strings.Replace(path, "{squawk}", fmt.Sprintf("%d", squawk), 1)
	req, err := c.getRequest(path)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("making request", zap.String("url", req.URL.String()))
	response, err := http.DefaultClient.Do(req)

	if err != nil {
		c.logger.Error("error making request", zap.Error(err))
		return nil, err
	}

	defer closeResponseBody(response.Body, c.logger)

	body, err := io.ReadAll(response.Body)
	if err != nil {
		c.logger.Error("error reading response body", zap.Error(err))
	}

	var result Result
	if err := json.Unmarshal(body, &result); err != nil {
		c.logger.Error("error unmarshalling response body", zap.Error(err))
		return nil, err
	}

	return c.filterADSBICAO(result.Aircraft), nil
}

// GetAircraftByCallsign returns aircraft by callsign
func (c *Client) GetAircraftByCallsign(callsign string) (*Aircraft, error) {
	path := c.Endpoints()["aircraft_by_callsign"]
	path = strings.Replace(path, "{callsign}", callsign, 1)
	req, err := c.getRequest(path)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("making request", zap.String("url", req.URL.String()))
	response, err := http.DefaultClient.Do(req)

	if err != nil {
		c.logger.Error("error making request", zap.Error(err))
		return nil, err
	}

	defer closeResponseBody(response.Body, c.logger)

	body, err := io.ReadAll(response.Body)
	if err != nil {
		c.logger.Error("error reading response body", zap.Error(err))
	}

	var aircraft Aircraft
	if err := json.Unmarshal(body, &aircraft); err != nil {
		c.logger.Error("error unmarshalling response body", zap.Error(err))
		return nil, err
	}

	return &aircraft, nil
}

// GetAircraftByHex returns aircraft by hex
func (c *Client) GetAircraftByHex(hex string) (*Aircraft, error) {
	path := c.Endpoints()["aircraft_by_hex"]
	path = strings.Replace(path, "{hex}", hex, 1)
	req, err := c.getRequest(path)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("making request", zap.String("url", req.URL.String()))
	response, err := http.DefaultClient.Do(req)

	if err != nil {
		c.logger.Error("error making request", zap.Error(err))
		return nil, err
	}

	defer closeResponseBody(response.Body, c.logger)

	body, err := io.ReadAll(response.Body)
	if err != nil {
		c.logger.Error("error reading response body", zap.Error(err))
	}

	var aircraft Aircraft
	if err := json.Unmarshal(body, &aircraft); err != nil {
		c.logger.Error("error unmarshalling response body", zap.Error(err))
		return nil, err
	}

	return &aircraft, nil
}

// GetAircraftByICAO returns aircraft by ICAO
func (c *Client) GetAircraftByICAO(icao string) (*Aircraft, error) {
	path := c.Endpoints()["aircraft_by_icao"]
	path = strings.Replace(path, "{icao}", icao, 1)
	req, err := c.getRequest(path)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("making request", zap.String("url", req.URL.String()))
	response, err := http.DefaultClient.Do(req)

	if err != nil {
		c.logger.Error("error making request", zap.Error(err))
		return nil, err
	}

	defer closeResponseBody(response.Body, c.logger)

	body, err := io.ReadAll(response.Body)
	if err != nil {
		c.logger.Error("error reading response body", zap.Error(err))
	}

	var aircraft Aircraft
	if err := json.Unmarshal(body, &aircraft); err != nil {
		c.logger.Error("error unmarshalling response body", zap.Error(err))
		return nil, err
	}

	return &aircraft, nil
}

// GetAircraftByRegistration returns aircraft by registration
func (c *Client) GetAircraftByRegistration(registration string) (*Aircraft, error) {
	path := c.Endpoints()["aircraft_by_registration"]
	path = strings.Replace(path, "{registration}", registration, 1)
	req, err := c.getRequest(path)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("making request", zap.String("url", req.URL.String()))
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		c.logger.Error("error making request", zap.Error(err))
		return nil, err
	}

	defer closeResponseBody(response.Body, c.logger)

	body, err := io.ReadAll(response.Body)
	if err != nil {
		c.logger.Error("error reading response body", zap.Error(err))
	}

	var aircraft Aircraft
	if err := json.Unmarshal(body, &aircraft); err != nil {
		c.logger.Error("error unmarshalling response body", zap.Error(err))
	}
	return &aircraft, nil
}

// GetAircraftWithinRange returns aircraft within a given distance from a given latitude and longitude
func (c *Client) GetAircraftWithinRange(
	lat float64,
	lon float64,
	dist int64,
) ([]*Aircraft, error) {
	path := c.Endpoints()["aircraft_within_range"]
	path = strings.Replace(path, "{lat}", fmt.Sprintf("%f", lat), 1)
	path = strings.Replace(path, "{lon}", fmt.Sprintf("%f", lon), 1)
	path = strings.Replace(path, "{dist}", fmt.Sprintf("%d", dist), 1)
	req, err := c.getRequest(path)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("making request", zap.String("url", req.URL.String()))
	response, err := http.DefaultClient.Do(req)

	if err != nil {
		c.logger.Error("error making request", zap.Error(err))
		return nil, err
	}

	defer closeResponseBody(response.Body, c.logger)

	body, err := io.ReadAll(response.Body)
	if err != nil {
		c.logger.Error("error reading response body", zap.Error(err))
		return nil, err
	}

	var result Result
	if err := json.Unmarshal(body, &result); err != nil {
		c.logger.Error("error unmarshaling response body", zap.Error(err))
		return nil, err
	}

	return c.filterADSBICAO(result.Aircraft), nil
}

// filterADSBICAO filters out aircraft without record type 'adsb_icao'
func (c *Client) filterADSBICAO(aircraft []*Aircraft) []*Aircraft {
	c.logger.Debug("filtering aircraft without record type 'adsb_icao' or 'mlat'")

	var filtered []*Aircraft

	for _, ac := range aircraft {
		if ac.RecordType != RecordTypeADSBICAO && ac.RecordType != RecordTypeMLAT {
			continue
		}

		ac.Flight = trimWhitespace(ac.Flight)
		filtered = append(filtered, ac)
	}

	return filtered
}

func (c *Client) getRequest(path string) (*http.Request, error) {
	requestURL := fmt.Sprintf("%s%s", c.URL, path)
	req, err := http.NewRequest("GET", requestURL, nil)

	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Add(headerRapidAPIKey, c.Key)
	req.Header.Add(headerRapidAPIHost, c.Host)
	req.Header.Add(headerContentType, "application/json")
	return req, nil
}

// trimWhitespace trims whitespace from a string
func trimWhitespace(input string) string {
	return strings.TrimSpace(input)
}

// closeResponseBody closes the response body
func closeResponseBody(body io.ReadCloser, logger *zap.Logger) {
	if err := body.Close(); err != nil {
		logger.Error("error closing response body", zap.Error(err))
	}
}
