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
)

// CalcMaster handles the "calc-m" command for printing the number of entries in the master table.
func (r *Repository) CalcMaster(cmd *cobra.Command, args []string) {

}

// InsertMaster handles the "insert-m" command for adding entries to the master table.
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
		fmt.Fprintf(os.Stderr, "record with id [%d] already exists. Use update-m to update a master record.\n", id)
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

	err = driver.WriteModel(r.App.Master.FL, course, 0, io.SeekStart)
	if err != nil {
		log.Println(err)
		return
	}

	r.App.Master.Indices = utils.AddMasterIndex(indices, uint32(id), io.SeekStart)

	fmt.Println(r.App.Master.Indices)
	fmt.Println(course)
}

// InsertSlave handles insert-s command for adding entries to the slave table.
func InsertSlave() {

}
