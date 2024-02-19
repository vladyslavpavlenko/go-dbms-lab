package handlers

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/models"
	"io"
	"strconv"
)

// CalcMaster handles calculation and printing the number of entries in the master table.
func (r *Repository) CalcMaster(_ *cobra.Command, _ []string) {
	fmt.Println(driver.NumberOfRecords(r.App.Master.Indices))
}

// CalcSlave handles calculation and printing the number of entries in the slave table.
func (r *Repository) CalcSlave(_ *cobra.Command, args []string) {
	if len(args) > 0 {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("error parsing ID: %v\n", err)
			return
		}

		address, ok := driver.GetAddressByIndex(r.App.Master.Indices, uint32(id))
		if !ok {
			fmt.Printf("the record with id %v does not exist\n", err)
			return
		}

		var course models.Course
		err = driver.ReadModel(r.App.Master.FL, &course, int64(address), io.SeekStart)
		if err != nil {
			fmt.Printf("error retrieving master model: %s\n", err)
			return
		}

		fmt.Println(driver.NumberOfSubrecords(r.App.Slave.FL, course.FirstSlaveAddress))
	} else {
		fmt.Println(driver.NumberOfRecords(r.App.Slave.Indices))
	}
}
