package audit

import (
	"context"

	"github.com/hashicorp/go-retryablehttp"
	"go.uber.org/zap"

	"net/http"
	"net/url"
	"time"
)

type externalAuditClient struct {
	URL    string
	client *retryablehttp.Client
}

func NewRetryableHttpClient(rawURL string) (observer, error) {
	if rawURL == "" {
		return nil, nil
	}
	_, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
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

func (ec *externalAuditClient) updateAudit(evAudit *EventAudit, appLogger *zap.Logger) {
	err := ec.sendAuditRequest(evAudit)
	if err != nil {
		appLogger.Error("Failed to send audit request", zap.Error(err))
	}
}

func (ac *externalAuditClient) sendAuditRequest(evAudit *EventAudit) error {
	request, err := retryablehttp.NewRequest(http.MethodGet, ac.URL, evAudit)
	if err != nil {
		return err
	}
	resp, err := ac.client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {

	}
	return nil
}
