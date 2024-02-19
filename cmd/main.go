package main

import (
	"bufio"
	"fmt"
	"github.com/adhocore/chin"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/config"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/models"
	"log"
	"os"
)

var app config.AppConfig

func main() {
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

	if slave.RequiresCompaction() {
		if driver.PromptCompactionConfirmation(masterName) {
			s := chin.New()
			go s.Start()

			s.Stop()
		} else {
			fmt.Println("Skipped")
		}
	}

	app.Master = master
	app.Slave = slave

	rootCmd := commands(&app)
	reader := bufio.NewReader(os.Stdin)

	err = run(rootCmd, reader)
	if err != nil {
		log.Fatal(err)
	}

	err = driver.WriteServiceData(masterName, app.Master.Indices, app.Master.Junk)
	if err != nil {
		log.Fatal(err)
	}

	err = driver.WriteServiceData(slaveName, app.Slave.Indices, app.Slave.Junk)
	if err != nil {
		log.Fatal(err)
	}
}
