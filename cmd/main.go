package main

import (
	"bufio"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/config"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver/utils"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/models"
	"log"
	"os"
)

var app config.AppConfig

func main() {
	master, err := driver.CreateTable("courses", models.Course{})
	if err != nil {
		log.Fatal(err)
	}
	app.Master = master

	slave, err := driver.CreateTable("certificates", models.Certificate{})
	if err != nil {
		log.Fatal(err)
	}
	app.Slave = slave

	rootCmd := commands(&app)
	reader := bufio.NewReader(os.Stdin)

	err = run(rootCmd, reader)
	if err != nil {
		log.Fatal(err)
	}

	utils.WriteIndices(app.Master.IND, app.Master.Indices)
	utils.WriteIndices(app.Slave.IND, app.Slave.Indices)
}
