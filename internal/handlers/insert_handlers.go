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

// InsertMaster handles adding entries to the master table.
func (r *Repository) InsertMaster(_ *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("error parsing ID: %v\n", err)
		return
	}
	title, category, instructor := args[1], args[2], args[3]

	indices := r.App.Master.Indices

	exists := driver.RecordExists(indices, uint32(id))
	if exists {
		fmt.Printf("record with ID %d already exists. Use update-m to update a master record\n", id)
		return
	}

	var course models.Course

	course.ID = uint32(id)
	copy(course.Title[:], title)
	copy(course.Category[:], category)
	copy(course.Instructor[:], instructor)
	course.FirstSlaveAddress = driver.NoLink
	course.Presence = true

	offset, _ := r.App.Master.FL.Seek(int64(len(r.App.Master.Indices)*r.App.Master.Size), io.SeekStart)

	if err := driver.WriteModel(r.App.Master.FL, &course, offset, io.SeekStart); err != nil {
		log.Println(err)
		return
	}

	r.App.Master.Indices = driver.AddIndex(r.App.Master.Indices, uint32(id), uint32(offset))

	fmt.Println("OK")
}

// InsertSlave handles adding entries to the slave table.
func (r *Repository) InsertSlave(_ *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("error parsing ID: %v\n", err)
		return
	}

	courseID, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Printf("error parsing course ID: %v\n", err)
		return
	}

	issuedTo := args[2]

	exists := driver.RecordExists(r.App.Slave.Indices, uint32(id))
	if exists {
		fmt.Printf("record with ID %d already exists. Use update-s to update a slave record.\n", id)
		return
	}

	var course models.Course

	masterAddress, ok := driver.GetAddressByIndex(r.App.Master.Indices, uint32(courseID))
	if !ok {
		fmt.Printf("the master record with ID %d was not found\n", courseID)
		return
	}

	err = driver.ReadModel(r.App.Master.FL, &course, int64(masterAddress), io.SeekStart)
	if err != nil {
		fmt.Printf("error retrieving master model: %s\n", err)
		return
	}

	var newCertificate models.Certificate
	newCertificate.ID = uint32(id)
	newCertificate.CourseID = uint32(courseID)
	newCertificate.Presence = true
	copy(newCertificate.IssuedTo[:], issuedTo)
	newCertificate.Next = driver.NoLink
	newCertificate.Previous = driver.NoLink

	var offset int64
	if len(r.App.Slave.Junk) > 0 {
		offset = int64(r.App.Slave.Junk[0])
		r.App.Slave.Junk = r.App.Slave.Junk[1:]
	} else {
		offset, _ = r.App.Slave.FL.Seek(0, io.SeekEnd)
	}

	if course.FirstSlaveAddress == driver.NoLink {
		course.FirstSlaveAddress = offset // first slave
	} else {
		// find the last certificate in the list to update its Next to the new certificate's offset.
		var lastCertificate models.Certificate
		currentOffset := course.FirstSlaveAddress
		var prevOffset int64 = driver.NoLink

		for currentOffset != driver.NoLink {
			err := driver.ReadModel(r.App.Slave.FL, &lastCertificate, currentOffset, io.SeekStart)
			if err != nil {
				fmt.Printf("error reading slave model: %s\n", err)
				return
			}

			prevOffset = currentOffset
			currentOffset = lastCertificate.Next
		}

		if prevOffset != driver.NoLink {
			lastCertificate.Next = offset
			newCertificate.Previous = prevOffset

			if err := driver.WriteModel(r.App.Slave.FL, &lastCertificate, prevOffset, io.SeekStart); err != nil {
				fmt.Printf("error updating last slave record: %s\n", err)
				return
			}
		}
	}

	if err := driver.WriteModel(r.App.Slave.FL, &newCertificate, offset, io.SeekStart); err != nil {
		log.Println(err)
		return
	}

	// Correctly update the course's first slave address if this is the first slave or if needed.
	if err := driver.WriteModel(r.App.Master.FL, &course, int64(masterAddress), io.SeekStart); err != nil {
		fmt.Printf("error updating master record with ID %d: %s\n", courseID, err)
		return
	}

	// Update indices with the correct offset after potentially using junk space or appending.
	r.App.Slave.Indices = driver.AddIndex(r.App.Slave.Indices, uint32(id), uint32(offset))

	fmt.Println("OK")
}
