package main

import (
	"encoding/json"
	"flag"
	"gopkg.in/yaml.v3"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"xray-geo-storage/storage"
)

type Config struct {
	Database struct {
		Host     string `yaml:"Host"`
		Port     string `yaml:"Port"`
		User     string `yaml:"User"`
		Password string `yaml:"Password"`
		Name     string `yaml:"Name"`
		SSLMode  string `yaml:"SSLMode"`
	} `yaml:"Database"`
	Server struct {
		Address string `yaml:"Address"`
		Port    string `yaml:"Port"`
	} `yaml:"Server"`
}

type WebhookPayload struct {
	GeoData   *storage.GeoData `json:"geo_data"`
	Hostname  string           `json:"hostname"`
	Timestamp time.Time        `json:"timestamp"`
}

var store storage.Storage
var config Config

func main() {
	var configPath string
	flag.StringVar(&configPath, "c", "config.yaml", "Path to configuration file")
	flag.Parse()

	if err := loadConfig(configPath); err != nil {
		log.Fatal("Failed to load config:", err)
	}

	var err error
	store, err = storage.NewPostgresStorage(
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
		config.Database.Name,
		config.Database.SSLMode,
	)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}
	log.Println("Successfully connected to PostgreSQL")

	defer store.Close()

	r := mux.NewRouter()
	r.HandleFunc("/webhook", handleWebhook).Methods("POST")
	r.HandleFunc("/health", handleHealth).Methods("GET")

	log.Printf("Server starting on %s:%s", config.Server.Address, config.Server.Port)
	addr := config.Server.Address + ":" + config.Server.Port
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func loadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	return nil
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	var payload WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if payload.Hostname == "" {
		http.Error(w, "Missing hostname in payload", http.StatusBadRequest)
		return
	}

	if payload.GeoData == nil {
		http.Error(w, "Missing geo_data in payload", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := store.Save(ctx, payload.Hostname, payload.GeoData); err != nil {
		log.Printf("Failed to insert data: %v", err)
		http.Error(w, "Failed to save data", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully added GeoData: ID=%d, Country=%s, City=%s, ISP=%s",
		payload.GeoData.ID,
		payload.GeoData.Country,
		payload.GeoData.City,
		payload.GeoData.ISP,
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Data saved successfully",
		"id":      payload.GeoData.ID,
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := store.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy", "storage": "disconnected"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "healthy",
		"storage":      "connected",
		"storage_type": "postgres",
		"timestamp":    time.Now().Unix(),
	})
}
