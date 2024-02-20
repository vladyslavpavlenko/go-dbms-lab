package main

import (
	"bufio"
	"fmt"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/config"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/models"
	"log"
	"os"
)

var app config.AppConfig

func main() {
	fmt.Println("program started")

	masterName := "courses"
	slaveName := "certificates"

	master, err := driver.CreateTable(masterName, models.Course{}, false)
	if err != nil {
		log.Fatal(err)
	}

	slave, err := driver.CreateTable(slaveName, models.Certificate{}, true)
	if err != nil {
		log.Fatal(err)
	}

	app.Master = master
	app.Slave = slave

	rootCmd := commands(&app)
	reader := bufio.NewReader(os.Stdin)

	err = run(rootCmd, reader)
	if err != nil {
		log.Fatal(err)
	}

	err = driver.WriteServiceData(masterName, app.Master.Indices, app.Master.Junk, false)
	if err != nil {
		log.Fatal(err)
	}

	err = driver.WriteServiceData(slaveName, app.Slave.Indices, app.Slave.Junk, true)
	if err != nil {
		log.Fatal(err)
	}
}
