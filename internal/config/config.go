package config

type Config struct {
	Port       string
	StorageDir string
	Username   string
	Password   string
}

func DefaultConfig() *Config {
	return &Config{
		Port:       "8080",
		StorageDir: "./data",
		Username:   "admin",
		Password:   "admin",
	}
}
