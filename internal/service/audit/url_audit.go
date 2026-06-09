package audit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/go-retryablehttp"

	"net/http"
	"net/url"
	"time"
)

type externalAuditClient struct {
	URL    string
	client *retryablehttp.Client
}

// NewRetryableHTTPClient creates an audit observer that posts events to rawURL.
func NewRetryableHTTPClient(rawURL string) (Observer, error) {
	if rawURL == "" {
		return nil, nil
	}
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("audit url must be absolute: %q", rawURL)
	}
	client := retryablehttp.NewClient()
	client.RetryMax = 5
	client.RetryWaitMin = 500 * time.Millisecond
	client.RetryWaitMax = 10 * time.Second
	client.HTTPClient.Timeout = time.Second * 10
	client.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if err != nil {
			return true, nil
		}

		if resp == nil {
			return false, nil
		}

		switch resp.StatusCode {
		case http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout:
			return true, nil
		default:
			return false, nil
		}
	}
	return &externalAuditClient{
		client: client,
		URL:    rawURL,
	}, nil
}

func (ec *externalAuditClient) updateAudit(evAudit *EventAudit) error {
	return ec.sendAuditRequest(evAudit)
}

func (ec *externalAuditClient) sendAuditRequest(evAudit *EventAudit) error {
	m, err := json.Marshal(evAudit)
	if err != nil {
		return err
	}
	request, err := retryablehttp.NewRequest(http.MethodPost, ec.URL, m)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	resp, err := ec.client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("audit receiver returned status %d", resp.StatusCode)
	}
	return nil
}
