package accrual

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	models "github.com/AleGaliev/gofermart/internal/model"
)

func TestHTTPSendler_GetInfoOrders(t *testing.T) {
	type fields struct {
		client  *http.Client
		url     *url.URL
		logger  logger
		keyHash string
		path    string
	}
	type args struct {
		orders string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    models.Order
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := HTTPSendler{
				client:  tt.fields.client,
				url:     tt.fields.url,
				logger:  tt.fields.logger,
				keyHash: tt.fields.keyHash,
				path:    tt.fields.path,
			}
			got, err := h.GetInfoOrders(tt.args.orders)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInfoOrders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetInfoOrders() got = %v, want %v", got, tt.want)
			}
		})
	}
}
