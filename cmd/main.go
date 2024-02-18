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
	masterName := "courses"
	slaveName := "certificates"

	master, err := driver.CreateTable(masterName, models.Course{})
	if err != nil {
		log.Fatal(err)
	}
	app.Master = master

	slave, err := driver.CreateTable(slaveName, models.Certificate{})
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

	err = utils.WriteServiceData(masterName, app.Master.Indices, app.Master.Junk)
	if err != nil {
		log.Fatal(err)
	}

	err = utils.WriteServiceData(slaveName, app.Slave.Indices, app.Slave.Junk)
	if err != nil {
		log.Fatal(err)
	}
}
