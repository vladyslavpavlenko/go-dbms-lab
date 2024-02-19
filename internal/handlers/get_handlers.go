package handlers

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver/utils"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/models"
	"io"
	"strconv"
	"strings"
)

// GetMaster handles printing entries from the master table based on ID and optional field names.
func (r *Repository) GetMaster(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Printf("error: at least 1 argument is required, got %d\n", len(args))
		err := cmd.Usage()
		if err != nil {
			return
		}
		return
	}

	var offset int64
	var all bool

	if args[0] == "all" {
		all = true
	} else {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("error parsing ID: %v\n", err)
			return
		}

		address, ok := utils.GetAddressByIndex(r.App.Master.Indices, uint32(id))
		if !ok {
			fmt.Printf("record with ID %d not found\n", id)
			return
		}

		offset += int64(address)
	}

	queries := make([]string, 0, len(args)-1)
	for _, q := range args[1:] {
		queries = append(queries, strings.ToUpper(q))
	}

	printMasterQuery(r.App.Master.FL, offset, queries, all)
}

// GetSlave handles printing entries from the slave table based on ID and optional field names.
func (r *Repository) GetSlave(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Printf("error: at least 1 argument is required, got %d\n", len(args))
		err := cmd.Usage()
		if err != nil {
			return
		}
		return
	}

	var all = args[0] == "all"
	id := 0
	var err error

	if !all {
		id, err = strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("error parsing ID: %v\n", err)
			return
		}

		exists := utils.RecordExists(r.App.Slave.Indices, uint32(id))
		if !exists {
			fmt.Printf("slave record with ID %d does not exist\n", id)
			return
		}
	}

	var courseID int
	var fsAddress int64 = -1
	var queries []string
	var offset int64

	if len(args) > 1 {
		queries = make([]string, 0, len(args)-1)
		for _, q := range args[1:] {
			queries = append(queries, strings.ToUpper(q))
		}

		if all {
			courseID, err = strconv.Atoi(queries[0])
			if err != nil {
				courseID = -1
			} else {
				exists := utils.RecordExists(r.App.Master.Indices, uint32(courseID))
				if !exists {
					fmt.Printf("master record with ID %d does not exist\n", courseID)
					return
				}

				address, ok := utils.GetAddressByIndex(r.App.Master.Indices, uint32(courseID))
				if !ok {
					fmt.Printf("record with ID %d not found\n", courseID)
					return
				}

				var model models.Course
				err = driver.ReadModel(r.App.Master.FL, &model, int64(address), io.SeekStart)
				if err != nil {
					fmt.Printf("error reading master data: %s\n", err)
					return
				}

				fsAddress = model.FirstSlaveAddress
				offset = fsAddress
			}
		}
	}

	if !all {
		address, ok := utils.GetAddressByIndex(r.App.Slave.Indices, uint32(id))
		if !ok {
			fmt.Printf("error getting index of the slave record with id %d: %s\n", id, err)
			return
		}

		offset = int64(address)
	}

	printSlaveQuery(r.App.Slave.FL, offset, queries, all)
}
