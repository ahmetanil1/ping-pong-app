package config

import "os"

type Config struct {
	ServerPort string
	PongURL    string
}

func Load() *Config {
	return &Config{
		ServerPort: getEnv("PORT", "8080"),
		// Gateway gibi çalışması için pong-service URL'sini de config'e ekledik
		PongURL:    getEnv("PONG_SERVICE_URL", "http://localhost:8081"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
