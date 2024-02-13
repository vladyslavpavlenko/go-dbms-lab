package handlers

import (
	"fmt"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver/utils"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/models"
	"io"
	"log"
	"os"
	"strconv"
)

// CalcMaster handles the "calc-m [id]" command for printing the number of entries in the master table.
func (r *Repository) CalcMaster(cmd *cobra.Command, args []string) {
	var id int
	var err error

	indices := r.App.Master.Indices

	if len(args) >= 1 {
		id, err = strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing ID: %v\n", err)
			return
		}
		fmt.Printf("ID is %d\n", id)
	} else {
		fmt.Println(utils.NumberOfRecords(indices))
	}
}

// CalcSlave handles the "calc-s" command for printing the number of entries in the slave table.
func (r *Repository) CalcSlave(cmd *cobra.Command, args []string) {
	indices := r.App.Slave.Indices

	fmt.Println(utils.NumberOfRecords(indices))
}

// InsertMaster handles the "insert-m <id> <title> <category> <instructor>" command for adding
// entries to the master table.
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

	offset := utils.NumberOfRecords(indices) * r.App.Master.Size // 108

	err = driver.WriteModel(r.App.Master.FL, course, int64(offset), io.SeekStart)
	if err != nil {
		log.Println(err)
		return
	}

	r.App.Master.Indices = utils.AddMasterIndex(indices, uint32(id), uint32(offset))

	fmt.Println(r.App.Master.Indices)
	fmt.Println(course)
}

// UtMaster handles the "ut-m" command for printing entries in the master table.
func (r *Repository) UtMaster(cmd *cobra.Command, args []string) {
	r.printMasterTable(true)
}

// GetMasterAll handles the "get-m -all" command for printing all entries in the master table.
func (r *Repository) GetMasterAll(cmd *cobra.Command, args []string) {
	r.printMasterTable(false)
}

// InsertSlave handles insert-s command for adding entries to the slave table.
func InsertSlave() {

}

func (r *Repository) printMasterTable(includeDetails bool) {
	flFile := r.App.Master.FL

	if _, err := flFile.Seek(0, io.SeekStart); err != nil {
		fmt.Fprintf(os.Stderr, "error seeking file: %s\n", err)
		return
	}

	var model models.Course
	var data []models.Course

	for {
		err := driver.ReadModel(flFile, &model, 0, io.SeekCurrent)
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "error reading data: %s\n", err)
			return
		}

		if model.Presence {
			data = append(data, model)
		}
	}

	headers := []string{"[ID]", "[TITLE]", "[CATEGORY]", "[INSTRUCTOR]"}
	if includeDetails {
		headers = append(headers, "[FC_ID]", "[PRESENCE]")
	}

	headersInterface := make([]interface{}, len(headers))
	for i, v := range headers {
		headersInterface[i] = v
	}

	tbl := table.New(headersInterface...).WithWriter(os.Stdout).WithPadding(4)

	for _, entry := range data {
		stringTitle := utils.ByteArrayToString(entry.Title[:])
		stringCategory := utils.ByteArrayToString(entry.Category[:])
		stringInstructor := utils.ByteArrayToString(entry.Instructor[:])

		row := []interface{}{entry.ID, stringTitle, stringCategory, stringInstructor}
		if includeDetails {
			row = append(row, entry.FirstSlaveID, entry.Presence)
		}

		tbl.AddRow(row...)
	}

	tbl.Print()
}
