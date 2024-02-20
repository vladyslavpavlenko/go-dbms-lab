package handlers

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/models"
	"io"
	"strconv"
)

// DeleteMaster handles deletion of the master record by its ID.
func (r *Repository) DeleteMaster(_ *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("error parsing ID: %v\n", err)
		return
	}

	address, ok := driver.GetAddressByIndex(r.App.Master.Indices, uint32(id))
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

	if course.FirstSlaveAddress != driver.NoLink {
		err = deleteSubrecords(r, course.FirstSlaveAddress)
		if err != nil {
			return
		}
	}

	lastRecordAddress, ok := driver.GetLastRecordAddress(r.App.Master.Indices)
	if !ok {
		fmt.Printf("error getting last record address: %v\n", err)
		return
	}

	if lastRecordAddress == address {
		r.App.Master.Indices = driver.RemoveIndex(r.App.Master.Indices, uint32(id))

		err = driver.TruncateFile(r.App.Master.FL, int64(lastRecordAddress))
		if err != nil {
			fmt.Printf("error truncating file: %v\n", err)
		}

		return
	}

	var lastRecord models.Course
	err = driver.MoveModel(r.App.Master.FL, &lastRecord, int64(lastRecordAddress), int64(address))
	if err != nil {
		fmt.Printf("error moving entry: %v\n", err)
		return
	}

	r.App.Master.Indices = driver.RemoveIndex(r.App.Master.Indices, uint32(id))
	r.App.Master.Indices = driver.UpdateAddress(r.App.Master.Indices, lastRecord.ID, address)

	err = driver.TruncateFile(r.App.Master.FL, int64(lastRecordAddress))
	if err != nil {
		fmt.Printf("error truncating file: %v\n", err)
	}

	fmt.Println("OK")
}

// DeleteSlave handles deletion of the slave record by its ID.
func (r *Repository) DeleteSlave(_ *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("error parsing ID: %v\n", err)
		return
	}

	certificateToDeleteAddress, ok := driver.GetAddressByIndex(r.App.Slave.Indices, uint32(id))
	if !ok {
		fmt.Printf("the slave record with ID %d was not found\n", id)
		return
	}

	var certificateToDelete models.Certificate
	err = driver.ReadModel(r.App.Slave.FL, &certificateToDelete, int64(certificateToDeleteAddress), io.SeekStart)
	if err != nil {
		fmt.Printf("error reading certificate: %s\n", err)
		return
	}

	courseID := certificateToDelete.CourseID

	courseAddress, ok := driver.GetAddressByIndex(r.App.Master.Indices, courseID)
	if !ok {
		fmt.Printf("error reading course: %s\n", err)
		return
	}

	if certificateToDelete.Previous == driver.NoLink && certificateToDelete.Next != driver.NoLink {
		err := deleteFirstNode(r, certificateToDelete, int64(courseAddress))
		if err != nil {
			fmt.Println(err)
			return
		}
	} else if certificateToDelete.Previous != driver.NoLink && certificateToDelete.Next != driver.NoLink {
		err := deleteMiddleNode(r, certificateToDelete)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else if certificateToDelete.Previous != driver.NoLink && certificateToDelete.Next == driver.NoLink {
		err := deleteLastNode(r, certificateToDelete)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	certificateToDelete.Presence = false
	certificateToDelete.Next = driver.NoLink
	certificateToDelete.Previous = driver.NoLink
	clear(certificateToDelete.IssuedTo[:])

	err = driver.WriteModel(r.App.Slave.FL, &certificateToDelete, int64(certificateToDeleteAddress), io.SeekStart)
	if err != nil {
		fmt.Printf("error updating certificateToDelete: %v\n", err)
		return
	}

	// update indices and junk
	r.App.Slave.Junk = append(r.App.Slave.Junk, certificateToDeleteAddress)
	r.App.Slave.Indices = driver.RemoveIndex(r.App.Slave.Indices, uint32(id))

	if r.App.Slave.RequiresCompaction() {
		updatedJunk, err := driver.CompactSlaveFile(r.App.Slave.FL, r.App.Slave.Indices, r.App.Slave.Junk)
		if err != nil {
			fmt.Println("error compacting file:", err)
			return
		}
		r.App.Slave.Junk = updatedJunk
	}

	fmt.Println("OK")
}
