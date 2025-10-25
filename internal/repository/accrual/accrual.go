package accrual

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	models "github.com/AleGaliev/gofermart/internal/model"
)

type logger interface {
	CreateResponseLog(statusCode int, large int64)
}
type HTTPSendler struct {
	client  *http.Client
	url     *url.URL
	logger  logger
	keyHash string
	path    string
}

func NewAccrualConfig(logger logger, baseURL string) *HTTPSendler {
	parsedURL, _ := url.Parse(baseURL)
	return &HTTPSendler{
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
		url: &url.URL{
			Scheme: parsedURL.Scheme,
			Host:   parsedURL.Host,
		},
		logger: logger,
		path:   "api/orders",
	}
}

func (h HTTPSendler) GetInfoOrders(orders string) (models.Order, error) {
	h.url.Path = fmt.Sprintf("%s/%s", h.path, orders)
	var order models.Order
	request, err := http.NewRequest(http.MethodGet, h.url.String(), nil)
	if err != nil {
		return order, fmt.Errorf("error creating request: %v", err)
	}
	request.Header.Add("Accept", "application/json")
	response, err := h.client.Do(request)
	if err != nil {
		return order, err
	}

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return order, err
	}
	if response.StatusCode == http.StatusNoContent {
		return models.Order{
			Status: "NEW",
		}, nil
	}

	if err = json.Unmarshal(body, &order); err != nil {
		return order, err
	}
	response.Body.Close()

	h.logger.CreateResponseLog(response.StatusCode, int64(len(body)))

	return order, nil
}
