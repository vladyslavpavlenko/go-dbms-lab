package main

import (
	"bufio"
	"fmt"
	"github.com/kballard/go-shellquote"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/config"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver/utils"
	"log"
	"os"
	"strings"
)

var app config.AppConfig

func main() {
	conn, err := driver.CreateTable("courses")
	if err != nil {
		log.Fatal(err)
	}

	app.Master = conn

	rootCmd := commands(&app)
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading command: %v\n", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "exit" {
			utils.WriteMasterIndices(app.Master.IND, app.Master.Indices)
			fmt.Println("Exiting...")
			break
		}

		args, err := shellquote.Split(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing command: %v\n", err)
			continue
		}

		rootCmd.SetArgs(args)
		if err := rootCmd.Execute(); err != nil {
			fmt.Fprintf(os.Stderr, "error executing command: %v\n", err)
		}
	}
}
