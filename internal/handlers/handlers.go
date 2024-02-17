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
	"strings"
)

// CalcMaster calculates and prints the number of entries in the master table, optionally calculating the number
// of entries in the slave table by Master entry's ID.
func (r *Repository) CalcMaster(_ *cobra.Command, args []string) {
	if len(args) >= 1 {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("error parsing ID: %v\n", err)
			return
		}
		fmt.Printf("ID is %d\n", id)
	} else {
		fmt.Println(utils.NumberOfRecords(r.App.Master.Indices))
	}
}

// CalcSlave calculates and prints the number of entries in the slave table.
func (r *Repository) CalcSlave(_ *cobra.Command, _ []string) {
	fmt.Println(utils.NumberOfRecords(r.App.Slave.Indices))
}

// InsertMaster handles adding entries to the master table.
func (r *Repository) InsertMaster(_ *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("error parsing <id>: %v\n", err)
		return
	}
	title, category, instructor := args[1], args[2], args[3]

	indices := r.App.Master.Indices

	exists := utils.RecordExists(indices, uint32(id))
	if exists {
		fmt.Printf("record with ID [%d] already exists. Use update-m to update a master record\n", id)
		return
	}

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

	r.App.Master.Indices = utils.AddIndex(r.App.Master.Indices, uint32(id), uint32(offset))
	log.Println("new master record added:", course)
}

// UtMaster prints all entries in the master table, including detailed information.
func (r *Repository) UtMaster(_ *cobra.Command, _ []string) {
	printMasterData(r.App.Master.FL, true)
}

// GetMaster prints entries from the master table based on ID and optional field names.
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

		offset = int64(address)
	}

	queries := make([]string, 0, len(args)-1)
	for _, q := range args[1:] {
		queries = append(queries, strings.ToUpper(q))
	}

	printMasterQuery(r.App.Master.FL, offset, queries, all)
}

// GetSlave prints entries from the slave table based on ID and optional field names.
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
	}

	var courseID int
	var firstID int64 = -1
	var queries []string

	if len(args) > 1 {
		queries = make([]string, 0, len(args)-1)
		for _, q := range args[1:] {
			queries = append(queries, strings.ToUpper(q))
		}

		courseID, err = strconv.Atoi(queries[0])
		if err != nil {
			courseID = -1
		} else {
			exists := utils.RecordExists(r.App.Master.Indices, uint32(courseID))
			if !exists {
				fmt.Printf("master record with ID [%d] does not exist\n", courseID)
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

			firstID = model.FirstSlaveID
		}
	}

	printSlaveQuery(r.App.Slave.FL, id, firstID, queries, all)
}

// UpdateMaster updates fields of the entry by its ID.
func (r *Repository) UpdateMaster(_ *cobra.Command, args []string) {
	if len(args) < 2 {
		fmt.Printf("error: at least 2 arguments are required, got %d\n", len(args))
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("error parsing <id>: %v\n", err)
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

	if len(args) > 1 && args[1] != "*" {
		clear(course.Title[:])
		copy(course.Title[:], args[1])
	}

	if len(args) > 2 && args[2] != "*" {
		clear(course.Category[:])
		copy(course.Category[:], args[2])
	}

	if len(args) > 3 && args[3] != "*" {
		clear(course.Instructor[:])
		copy(course.Instructor[:], args[3])
	}

	if err := driver.WriteModel(r.App.Master.FL, &course, int64(address), io.SeekStart); err != nil {
		fmt.Printf("error updating record: %s\n", err)
		return
	}

	log.Println("Master record updated:", course)
}

// UpdateSlave updates fields of the entry by its ID.
func (r *Repository) UpdateSlave(_ *cobra.Command, args []string) {
	if len(args) < 2 {
		fmt.Printf("error: at least 2 arguments are required, got %d\n", len(args))
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("error parsing <id>: %v\n", err)
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

func (r *Repository) DeleteMaster(_ *cobra.Command, _ []string) {

}

// InsertSlave is a placeholder for adding entries to the slave table.
func (r *Repository) InsertSlave(_ *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("error parsing <id>: %v\n", err)
		return
	}

	courseID, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Printf("error parsing <course_id>: %v\n", err)
		return
	}

	issuedTo := args[2]

	indices := r.App.Slave.Indices
	exists := utils.RecordExists(indices, uint32(id))
	if exists {
		fmt.Printf("record with ID %d already exists. Use update-s to update a slave record.\n", id)
		return
	}

	masterAddress, ok := utils.GetAddressByIndex(r.App.Master.Indices, uint32(courseID))
	if !ok {
		fmt.Printf("the master record with ID %d was not found\n", id)
		return
	}

	var certificate models.Certificate
	certificate.ID = uint32(id)
	certificate.CourseID = uint32(courseID)
	copy(certificate.IssuedTo[:], issuedTo)
	certificate.Presence = true

	offset := utils.NumberOfRecords(r.App.Slave.Indices) * r.App.Slave.Size

	if err := driver.WriteModel(r.App.Slave.FL, &certificate, int64(offset), io.SeekStart); err != nil {
		log.Println(err)
		return
	}

	r.App.Slave.Indices = utils.AddIndex(r.App.Slave.Indices, uint32(id), uint32(offset))
	log.Println("New slave record added:", certificate)

	// update first_slave_id
	var course models.Course
	err = driver.ReadModel(r.App.Master.FL, &course, int64(masterAddress), io.SeekStart)
	if err != nil {
		fmt.Printf("error retrieving master model: %s\n", err)
		return
	}

	if course.FirstSlaveID == -1 {
		course.FirstSlaveID = int64(masterAddress)
	}

	if err := driver.WriteModel(r.App.Master.FL, &course, int64(masterAddress), io.SeekStart); err != nil {
		fmt.Printf("error updating master record with ID %d: %s\n", courseID, err)
		return
	}
}

// UtSlave prints all entries in the slave table, including detailed information.
func (r *Repository) UtSlave(_ *cobra.Command, _ []string) {
	printSlaveData(r.App.Slave.FL, true)
}
