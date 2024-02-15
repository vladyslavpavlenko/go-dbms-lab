package handlers

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver/utils"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/models"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

// CalcMaster calculates and prints the number of entries in the master table, optionally calculating the number
// of entries in the slave table by Master entry's ID.
func (r *Repository) CalcMaster(cmd *cobra.Command, args []string) {
	if len(args) >= 1 {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing ID: %v\n", err)
			return
		}
		fmt.Printf("ID is %d\n", id)
	} else {
		fmt.Println(utils.NumberOfRecords(r.App.Master.Indices))
	}
}

// CalcSlave calculates and prints the number of entries in the slave table.
func (r *Repository) CalcSlave(cmd *cobra.Command, args []string) {
	fmt.Println(utils.NumberOfRecords(r.App.Slave.Indices))
}

// InsertMaster handles adding entries to the master table.
func (r *Repository) InsertMaster(cmd *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <id>: %v\n", err)
		return
	}
	title, category, instructor := args[1], args[2], args[3]

	indices := r.App.Master.Indices

	exists := utils.RecordExists(indices, uint32(id))
	if exists {
		fmt.Fprintf(os.Stderr, "record with ID [%d] already exists. Use update-m to update a master record.\n", id)
		return
	}

	// create and populate a new course entity
	var course models.Course

	course.ID = uint32(id)
	copy(course.Title[:], title)
	copy(course.Category[:], category)
	copy(course.Instructor[:], instructor)
	course.FirstSlaveID = -1
	course.Presence = true

	offset := utils.NumberOfRecords(r.App.Master.Indices) * r.App.Master.Size

	if err := driver.WriteModel(r.App.Master.FL, &course, int64(offset), io.SeekStart); err != nil {
		log.Println(err)
		return
	}

	r.App.Master.Indices = utils.AddMasterIndex(r.App.Master.Indices, uint32(id), uint32(offset))
	fmt.Println("New master record added:", course)
}

// InsertSlave is a placeholder for adding entries to the slave table.
func InsertSlave() {
	// Placeholder function.
}

// UtMaster prints all entries in the master table, including detailed information.
func (r *Repository) UtMaster(cmd *cobra.Command, args []string) {
	printMasterData(r.App.Master.FL, true)
}

// GetMaster prints entries from the master table based on ID and optional field names.
func (r *Repository) GetMaster(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "error: 1 argument expected, got %d\n", len(args))
		cmd.Usage()
		return
	}

	var offset int64
	var all bool

	if args[0] == "all" {
		all = true
	} else {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing ID: %v\n", err)
			return
		}

		address, ok := utils.GetAddressByIndex(r.App.Master.Indices, uint32(id))
		if !ok {
			fmt.Fprintf(os.Stderr, "record with ID %d not found\n", id)
			return
		}

		offset = int64(address)
	}

	queries := make([]string, 0, len(args)-1)
	for _, q := range args[1:] {
		queries = append(queries, strings.ToLower(q))
	}

	printMasterQuery(r.App.Master.FL, offset, queries, all)
}
