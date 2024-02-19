package handlers

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/models"
	"io"
	"log"
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

	if course.FirstSlaveAddress != -1 {
		r.deleteSubrecords(r.App.Slave.FL, course.FirstSlaveAddress)
	}

	lastRecordAddress, ok := driver.GetLastRecordAddress(r.App.Master.Indices)
	if !ok {
		fmt.Printf("error getting last record address: %v\n", err)
		return
	}

	if lastRecordAddress == address {
		log.Println("deleting last entry")

		r.App.Master.Indices = driver.RemoveIndex(r.App.Master.Indices, uint32(id))

		err = driver.TruncateFile(r.App.Master.FL, int64(lastRecordAddress))
		if err != nil {
			fmt.Printf("error truncating file: %v\n", err)
		}

		log.Println("entry deleted")
		return
	}

	var lastRecord models.Course
	err = driver.ReadModel(r.App.Master.FL, &lastRecord, int64(lastRecordAddress), io.SeekStart)
	if err != nil {
		fmt.Printf("error reading last record: %v\n", err)
		return
	}

	if err := driver.WriteModel(r.App.Master.FL, &lastRecord, int64(address), io.SeekStart); err != nil {
		log.Printf("error moving last record: %v\n", err)
		return
	}

	r.App.Master.Indices = driver.RemoveIndex(r.App.Master.Indices, uint32(id))
	r.App.Master.Indices = driver.UpdateAddress(r.App.Master.Indices, lastRecord.ID, address)

	err = driver.TruncateFile(r.App.Master.FL, int64(lastRecordAddress))
	if err != nil {
		fmt.Printf("error truncating file: %v\n", err)
	}

	log.Printf("model with id %d was deleted", id)
	log.Println("Updated indices:", r.App.Master.Indices)
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
		fmt.Printf("error retrieving model: %s\n", err)
		return
	}

	courseID := certificateToDelete.CourseID

	courseAddress, ok := driver.GetAddressByIndex(r.App.Master.Indices, courseID)
	if !ok {
		fmt.Printf("error retrieving course: %s\n", err)
		return
	}

	// this is the first node
	if certificateToDelete.Previous == -1 {
		var course models.Course
		err = driver.ReadModel(r.App.Master.FL, &course, int64(courseAddress), io.SeekStart)
		if err != nil {
			fmt.Printf("error retrieving course model: %s\n", err)
			return
		}

		course.FirstSlaveAddress = certificateToDelete.Next

		err = driver.WriteModel(r.App.Master.FL, &course, int64(courseAddress), io.SeekStart)
		if err != nil {
			fmt.Printf("error updating course first slave address: %v\n", err)
			return
		}

		var nextCertificate models.Certificate

		err = driver.ReadModel(r.App.Slave.FL, &nextCertificate, certificateToDelete.Next, io.SeekStart)
		if err != nil {
			fmt.Printf("error retrieving nextCertificate model: %s\n", err)
			return
		}

		nextCertificate.Previous = -1

		err = driver.WriteModel(r.App.Slave.FL, &nextCertificate, certificateToDelete.Next, io.SeekStart)
		if err != nil {
			fmt.Printf("error updating nextCertificate: %v\n", err)
			return
		}

		certificateToDelete.Presence = false
		certificateToDelete.Next = -1
		certificateToDelete.Previous = -1
		clear(certificateToDelete.IssuedTo[:])

		err = driver.WriteModel(r.App.Slave.FL, &certificateToDelete, int64(certificateToDeleteAddress), io.SeekStart)
		if err != nil {
			fmt.Printf("error updating course first slave address: %v\n", err)
			return
		}

		// Update indices and junk.
		r.App.Slave.Junk = append(r.App.Slave.Junk, certificateToDeleteAddress)
		r.App.Slave.Indices = driver.RemoveIndex(r.App.Slave.Indices, uint32(id))

		log.Println("deleted first node")
		return
	}

	// this is the middle node
	if certificateToDelete.Next != -1 {
		var previousCertificate models.Certificate

		err = driver.ReadModel(r.App.Slave.FL, &previousCertificate, certificateToDelete.Previous, io.SeekStart)
		if err != nil {
			fmt.Printf("error retrieving previousCertificate model: %s\n", err)
			return
		}

		previousCertificate.Next = certificateToDelete.Next

		err = driver.WriteModel(r.App.Slave.FL, &previousCertificate, certificateToDelete.Previous, io.SeekStart)
		if err != nil {
			fmt.Printf("error updating previousCertificate: %v\n", err)
			return
		}

		var nextCertificate models.Certificate

		err = driver.ReadModel(r.App.Slave.FL, &nextCertificate, certificateToDelete.Next, io.SeekStart)
		if err != nil {
			fmt.Printf("error retrieving nextCertificate model: %s\n", err)
			return
		}

		nextCertificate.Previous = certificateToDelete.Previous

		err = driver.WriteModel(r.App.Slave.FL, &nextCertificate, certificateToDelete.Next, io.SeekStart)
		if err != nil {
			fmt.Printf("error updating nextCertificate: %v\n", err)
			return
		}

		certificateToDelete.Presence = false
		certificateToDelete.Next = -1
		certificateToDelete.Previous = -1
		clear(certificateToDelete.IssuedTo[:])

		err = driver.WriteModel(r.App.Slave.FL, &certificateToDelete, int64(certificateToDeleteAddress), io.SeekStart)
		if err != nil {
			fmt.Printf("error updating certificateToDelete: %v\n", err)
			return
		}

		// Update indices and junk.
		r.App.Slave.Junk = append(r.App.Slave.Junk, certificateToDeleteAddress)
		r.App.Slave.Indices = driver.RemoveIndex(r.App.Slave.Indices, uint32(id))

		log.Println("deleted middle node")
		return
	}

	if certificateToDelete.Previous != -1 && certificateToDelete.Next == -1 {
		var previousCertificate models.Certificate

		err = driver.ReadModel(r.App.Slave.FL, &previousCertificate, certificateToDelete.Previous, io.SeekStart)
		if err != nil {
			fmt.Printf("error retrieving previousCertificate model: %s\n", err)
			return
		}

		previousCertificate.Next = -1

		err = driver.WriteModel(r.App.Slave.FL, &previousCertificate, certificateToDelete.Previous, io.SeekStart)
		if err != nil {
			fmt.Printf("error updating previousCertificate: %v\n", err)
			return
		}

		certificateToDelete.Presence = false
		certificateToDelete.Next = -1
		certificateToDelete.Previous = -1
		clear(certificateToDelete.IssuedTo[:])

		err = driver.WriteModel(r.App.Slave.FL, &certificateToDelete, int64(certificateToDeleteAddress), io.SeekStart)
		if err != nil {
			fmt.Printf("error updating certificateToDelete: %v\n", err)
			return
		}

		// Update indices and junk.
		r.App.Slave.Junk = append(r.App.Slave.Junk, certificateToDeleteAddress)
		r.App.Slave.Indices = driver.RemoveIndex(r.App.Slave.Indices, uint32(id))

		log.Println("deleted last node")
		return
	}
}
