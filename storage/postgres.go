package storage

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresStorage struct {
	db *gorm.DB
}

type PostgresGeoData struct {
	ID          uint      `gorm:"primaryKey"`
	Hostname    string    `gorm:"column:hostname;size:100;index:idx_hostname"`
	Country     string    `gorm:"column:country;size:100;index:idx_country"`
	CountryCode string    `gorm:"column:country_code;size:10;index:idx_country_code"`
	Region      string    `gorm:"column:region;size:10;index:idx_region"`
	RegionName  string    `gorm:"column:region_name;size:100"`
	City        string    `gorm:"column:city;size:100;index:idx_city;index:idx_country_city,composite:country"`
	Zip         string    `gorm:"column:zip;size:20;index:idx_zip"`
	Lat         float64   `gorm:"column:lat;index:idx_coordinates,composite:lon"`
	Lon         float64   `gorm:"column:lon;index:idx_coordinates,composite:lat"`
	Timezone    string    `gorm:"column:timezone;size:100;index:idx_timezone"`
	ISP         string    `gorm:"column:isp;size:255;index:idx_isp"`
	Org         string    `gorm:"column:org;size:255;index:idx_org"`
	AS          string    `gorm:"column:as_info;size:255"`
	CreatedAt   time.Time `gorm:"autoCreateTime;index:idx_created_at"`
}

func (PostgresGeoData) TableName() string {
	return "geodata"
}

func NewPostgresStorage(host, port, user, password, dbname, sslmode string) (*PostgresStorage, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&PostgresGeoData{}); err != nil {
		return nil, err
	}

	// Create additional optimized indexes
	if err := createOptimizedIndexes(db); err != nil {
		return nil, err
	}

	return &PostgresStorage{db: db}, nil
}

func createOptimizedIndexes(db *gorm.DB) error {
	// Create partial indexes for common queries
	indexes := []string{
		// Index for ISP queries with country
		"CREATE INDEX IF NOT EXISTS idx_isp_country ON geodata(isp, country)",

		// Index for geographic queries
		"CREATE INDEX IF NOT EXISTS idx_geo_location ON geodata(country, region, city)",

		// BRIN index for time series data (very efficient for large tables)
		"CREATE INDEX IF NOT EXISTS idx_created_at_brin ON geodata USING BRIN (created_at)",

		// GiST index for spatial queries if needed in future
		"CREATE INDEX IF NOT EXISTS idx_lat_lon_gist ON geodata USING GIST (point(lat, lon))",
	}

	for _, idx := range indexes {
		if err := db.Exec(idx).Error; err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Analyze the table to update statistics
	if err := db.Exec("ANALYZE geodata").Error; err != nil {
		// Non-critical error, just log it
		fmt.Printf("Warning: failed to analyze table: %v\n", err)
	}

	return nil
}

func (p *PostgresStorage) Save(ctx context.Context, hostname string, data *GeoData) error {
	pgData := &PostgresGeoData{
		Hostname:    hostname,
		Country:     data.Country,
		CountryCode: data.CountryCode,
		Region:      data.Region,
		RegionName:  data.RegionName,
		City:        data.City,
		Zip:         data.Zip,
		Lat:         data.Lat,
		Lon:         data.Lon,
		Timezone:    data.Timezone,
		ISP:         data.ISP,
		Org:         data.Org,
		AS:          data.AS,
	}

	result := p.db.WithContext(ctx).Create(pgData)
	if result.Error != nil {
		return result.Error
	}

	data.ID = pgData.ID
	data.CreatedAt = pgData.CreatedAt
	return nil
}

func (p *PostgresStorage) Close() error {
	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (p *PostgresStorage) Ping(ctx context.Context) error {
	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}
