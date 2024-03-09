package config

type ServerConfig struct {
	Host            string   `json:"host"`
	Port            int      `json:"port"`
	IdleTimeout     Duration `json:"idle_timeout"`
	ReadTimeout     Duration `json:"read_timeout"`
	WriteTimeout    Duration `json:"write_timeout"`
	ShutdownDelay   Duration `json:"shutdown_delay"`
	ShutdownTimeout Duration `json:"shutdown_timeout"`
}

type Log struct {
	Severity string `json:"severity"`
	Handler  string `json:"handler"`
}
