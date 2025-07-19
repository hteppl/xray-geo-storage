package storage

import (
	"context"
	"time"
)

type GeoData struct {
	ID          uint      `json:"id,omitempty"`
	Country     string    `json:"country"`
	CountryCode string    `json:"countryCode"`
	Region      string    `json:"region"`
	RegionName  string    `json:"regionName"`
	City        string    `json:"city"`
	Zip         string    `json:"zip"`
	Lat         float64   `json:"lat"`
	Lon         float64   `json:"lon"`
	Timezone    string    `json:"timezone"`
	ISP         string    `json:"isp"`
	Org         string    `json:"org"`
	AS          string    `json:"as"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
}

type Storage interface {
	Save(ctx context.Context, hostname string, data *GeoData) error
	Close() error
	Ping(ctx context.Context) error
}
