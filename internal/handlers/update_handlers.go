package handlers

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver/utils"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/models"
	"io"
	"log"
	"strconv"
)

// UpdateMaster handles updating fields of the master entry by its ID.
func (r *Repository) UpdateMaster(_ *cobra.Command, args []string) {
	if len(args) < 2 {
		fmt.Printf("error: at least 2 arguments are required, got %d\n", len(args))
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("error parsing ID: %v\n", err)
		return
	}

	address, ok := utils.GetAddressByIndex(r.App.Master.Indices, uint32(id))
	if !ok {
		fmt.Printf("the record with ID %d was not found\n", id)
		return
	}

	var course models.Course
	err = driver.ReadModel(r.App.Master.FL, &course, int64(address), io.SeekStart)
	if err != nil {
		fmt.Printf("error retrieving model: %s\n", err)
		return
	}

	if len(args) > 1 && args[1] != "-" {
		clear(course.Title[:])
		copy(course.Title[:], args[1])
	}

	if len(args) > 2 && args[2] != "-" {
		clear(course.Category[:])
		copy(course.Category[:], args[2])
	}

	if len(args) > 3 && args[3] != "-" {
		clear(course.Instructor[:])
		copy(course.Instructor[:], args[3])
	}

	if err := driver.WriteModel(r.App.Master.FL, &course, int64(address), io.SeekStart); err != nil {
		fmt.Printf("error updating record: %s\n", err)
		return
	}

	log.Println("Master record updated:", course)
}

// UpdateSlave handles updating fields of the slave entry by its ID.
func (r *Repository) UpdateSlave(_ *cobra.Command, args []string) {
	if len(args) < 2 {
		fmt.Printf("error: at least 2 arguments are required, got %d\n", len(args))
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("error parsing ID: %v\n", err)
		return
	}

	address, ok := utils.GetAddressByIndex(r.App.Slave.Indices, uint32(id))
	if !ok {
		fmt.Printf("the record with ID %d was not found\n", id)
		return
	}

	var certificate models.Certificate
	err = driver.ReadModel(r.App.Slave.FL, &certificate, int64(address), io.SeekStart)
	if err != nil {
		fmt.Printf("error retrieving model: %s\n", err)
		return
	}

	if len(args) > 1 && args[1] != "*" {
		clear(certificate.IssuedTo[:])
		copy(certificate.IssuedTo[:], args[1])
	} else {
		fmt.Println("nothing to update")
		return
	}

	if err := driver.WriteModel(r.App.Slave.FL, &certificate, int64(address), io.SeekStart); err != nil {
		fmt.Printf("error updating record: %s\n", err)
		return
	}

	log.Println("Slave record updated:", certificate)
}
