package config

import (
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver"
)

// AppConfig holds connections and variable needed by an application.
type AppConfig struct {
	Master *driver.Table
	Slave  *driver.Table
}
