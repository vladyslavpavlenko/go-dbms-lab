package handlers

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/models"
	"io"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
)

// printMasterQuery prints selected fields from the master table based on provided field queries. If all is true,
// all records are printed.
func printMasterQuery(flFile *os.File, offset int64, queries []string, all bool) {
	if all {
		_, err := flFile.Seek(offset, io.SeekStart)
		if err != nil {
			return
		}
	} else {
		_, err := flFile.Seek(offset, io.SeekStart)
		if err != nil {
			return
		}
	}

	headers := []string{"ID"}

	if len(queries) != 0 {
		for _, query := range queries {

			switch query {
			case "ID":
			case "TITLE":
				headers = append(headers, "TITLE")
			case "CATEGORY":
				headers = append(headers, "CATEGORY")
			case "INSTRUCTOR":
				headers = append(headers, "INSTRUCTOR")
			case "FS_ADDRESS":
				headers = append(headers, "FS_ADDRESS")
			case "PRESENCE":
				headers = append(headers, "PRESENCE")
			default:
				fmt.Printf("field '%s' was not found\n", strings.ToLower(query))
			}
		}
	} else {
		headers = append(headers, "TITLE", "CATEGORY", "INSTRUCTOR")
	}

	if len(headers) == 1 && !slices.Contains(queries, "id") {
		fmt.Println("nothing to show")
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for {
		var model models.Course
		err := driver.ReadModel(flFile, &model, 0, io.SeekCurrent)
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Printf("error reading data: %s\n", err)
			return
		}

		if model.Presence == false {
			log.Println("entry is not present...")
			continue
		}

		var row []string
		for _, header := range headers {
			switch strings.ToUpper(header) {
			case "ID":
				row = append(row, strconv.Itoa(int(model.ID)))
			case "TITLE":
				row = append(row, driver.ByteArrayToString(model.Title[:]))
			case "CATEGORY":
				row = append(row, driver.ByteArrayToString(model.Category[:]))
			case "INSTRUCTOR":
				row = append(row, driver.ByteArrayToString(model.Instructor[:]))
			case "FS_ADDRESS":
				row = append(row, strconv.Itoa(int(model.FirstSlaveAddress)))
			case "PRESENCE":
				row = append(row, strconv.FormatBool(model.Presence))
			}
		}

		table.Append(row)
		if !all {
			break
		}
	}

	table.Render()
}

// printSlaveQuery prints selected fields from the slave table based on provided field queries.
// If all is true, all records are printed. If courseIDFilter is not -1, it filters records by course ID.
func printSlaveQuery(flFile *os.File, offset int64, queries []string, all bool) {
	_, err := flFile.Seek(offset, io.SeekStart)
	if err != nil {
		fmt.Println(0)
		return
	}

	headers := []string{"ID", "COURSE_ID", "ISSUED_TO"}

	courseIDFilter := driver.NoLink
	if len(queries) > 0 && all {
		parsedID, err := strconv.Atoi(queries[0])
		if err == nil {
			courseIDFilter = parsedID
			queries = queries[1:]
		}
	}

	if len(queries) > 0 {
		headers = []string{"ID", "COURSE_ID"}
		for _, query := range queries {
			switch strings.ToUpper(query) {
			case "ISSUED_TO", "PREVIOUS", "NEXT", "PRESENCE":
				headers = append(headers, query)
			}
		}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for {
		var model models.Certificate
		err := driver.ReadModel(flFile, &model, 0, io.SeekCurrent)
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Printf("error reading slave data: %s\n", err)
			return
		}

		if model.Presence == false {
			log.Println("checking for presence...")
			continue
		}

		if courseIDFilter != driver.NoLink && int(model.CourseID) != courseIDFilter {
			offset = model.Next
			continue
		}

		var row []string
		for _, header := range headers {
			switch header {
			case "ID":
				row = append(row, strconv.Itoa(int(model.ID)))
			case "COURSE_ID":
				row = append(row, strconv.Itoa(int(model.CourseID)))
			case "ISSUED_TO":
				row = append(row, driver.ByteArrayToString(model.IssuedTo[:]))
			case "PREVIOUS":
				row = append(row, strconv.Itoa(int(model.Previous)))
			case "NEXT":
				row = append(row, strconv.Itoa(int(model.Next)))
			case "PRESENCE":
				row = append(row, strconv.FormatBool(model.Presence))
			}
		}

		table.Append(row)

		if !all {
			break
		}

		offset = model.Next
	}

	table.Render()
}

func deleteSubrecords(r *Repository, address int64) error {
	for address != driver.NoLink {
		var model models.Certificate

		err := driver.ReadModel(r.App.Slave.FL, &model, address, io.SeekStart)
		if err != nil {
			return fmt.Errorf("error reading slave record for deletion: %w", err)
		}

		nextAddress := model.Next

		clear(model.IssuedTo[:])
		model.Next = driver.NoLink
		model.Previous = driver.NoLink
		model.Presence = false

		r.App.Slave.Junk = append(r.App.Slave.Junk, uint32(address))
		r.App.Slave.Indices = driver.RemoveIndex(r.App.Slave.Indices, model.ID)

		err = driver.WriteModel(r.App.Slave.FL, &model, address, io.SeekStart)
		if err != nil {
			return fmt.Errorf("error updating slave record to mark as deleted: %w", err)
		}

		address = nextAddress
	}

	if r.App.Slave.RequiresCompaction() {
		updatedJunk, err := driver.CompactSlaveFile(r.App.Slave.FL, r.App.Slave.Indices, r.App.Slave.Junk)
		if err != nil {
			return fmt.Errorf("error compacting file: %w", err)
		}
		r.App.Slave.Junk = updatedJunk
	}

	return nil
}

// deleteFirstNode handles first node deletion.
func deleteFirstNode(r *Repository, certificateToDelete models.Certificate, courseAddress int64) error {
	var course models.Course
	err := driver.ReadModel(r.App.Master.FL, &course, courseAddress, io.SeekStart)
	if err != nil {
		return fmt.Errorf("error reading course: %w", err)
	}

	course.FirstSlaveAddress = certificateToDelete.Next

	err = driver.WriteModel(r.App.Master.FL, &course, courseAddress, io.SeekStart)
	if err != nil {
		return fmt.Errorf("error updating course.FirstSlaveAddress: %w", err)
	}

	var nextCertificate models.Certificate

	err = driver.ReadModel(r.App.Slave.FL, &nextCertificate, certificateToDelete.Next, io.SeekStart)
	if err != nil {
		return fmt.Errorf("error reading nextCertificate model: %w", err)
	}

	nextCertificate.Previous = driver.NoLink

	err = driver.WriteModel(r.App.Slave.FL, &nextCertificate, certificateToDelete.Next, io.SeekStart)
	if err != nil {
		return fmt.Errorf("error updating nextCertificate: %w", err)
	}

	return nil
}

// deleteMiddleNode handles middle node deletion.
func deleteMiddleNode(r *Repository, certificateToDelete models.Certificate) error {
	var previousCertificate models.Certificate

	err := driver.ReadModel(r.App.Slave.FL, &previousCertificate, certificateToDelete.Previous, io.SeekStart)
	if err != nil {
		return fmt.Errorf("error reading previousCertificate: %w", err)
	}

	previousCertificate.Next = certificateToDelete.Next

	err = driver.WriteModel(r.App.Slave.FL, &previousCertificate, certificateToDelete.Previous, io.SeekStart)
	if err != nil {
		return fmt.Errorf("error updating previousCertificate: %w", err)
	}

	var nextCertificate models.Certificate

	err = driver.ReadModel(r.App.Slave.FL, &nextCertificate, certificateToDelete.Next, io.SeekStart)
	if err != nil {
		return fmt.Errorf("error reading nextCertificate: %w", err)
	}

	nextCertificate.Previous = certificateToDelete.Previous

	err = driver.WriteModel(r.App.Slave.FL, &nextCertificate, certificateToDelete.Next, io.SeekStart)
	if err != nil {
		return fmt.Errorf("error updating nextCertificate: %w", err)
	}

	return nil
}

// deleteLastNode handles last node deletion.
func deleteLastNode(r *Repository, certificateToDelete models.Certificate) error {
	var previousCertificate models.Certificate

	err := driver.ReadModel(r.App.Slave.FL, &previousCertificate, certificateToDelete.Previous, io.SeekStart)
	if err != nil {
		return fmt.Errorf("error reading previousCertificate: %w", err)
	}

	previousCertificate.Next = driver.NoLink

	err = driver.WriteModel(r.App.Slave.FL, &previousCertificate, certificateToDelete.Previous, io.SeekStart)
	if err != nil {
		return fmt.Errorf("error updating previousCertificate: %w", err)
	}

	return nil
}
