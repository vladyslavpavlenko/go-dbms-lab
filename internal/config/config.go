package config

import (
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver"
)

// AppConfig holds application connections to Master and Slave files.
type AppConfig struct {
	Master *driver.Table
	Slave  *driver.Table
}
