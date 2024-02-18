package handlers

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver/utils"
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
				row = append(row, utils.ByteArrayToString(model.Title[:]))
			case "CATEGORY":
				row = append(row, utils.ByteArrayToString(model.Category[:]))
			case "INSTRUCTOR":
				row = append(row, utils.ByteArrayToString(model.Instructor[:]))
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
func printSlaveQuery(flFile *os.File, offset int64, id int, fsAddress int64, queries []string, all bool) {
	_, err := flFile.Seek(offset, io.SeekStart)
	if err != nil {
		fmt.Println("Failed to seek file:", err)
		return
	}

	headers := []string{"ID", "COURSE_ID", "ISSUED_TO"}

	courseIDFilter := -1
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
			fmt.Printf("rrror reading slave data: %s\n", err)
			return
		}

		if model.Presence == false {
			log.Println("checking for presence...")
			continue
		}

		if courseIDFilter != -1 && int(model.CourseID) != courseIDFilter {
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
				row = append(row, utils.ByteArrayToString(model.IssuedTo[:]))
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

func (r *Repository) deleteSubrecords(flFile *os.File, address int64) {
	for address != -1 {
		var model models.Certificate

		err := driver.ReadModel(flFile, &model, address, io.SeekStart)
		if err != nil {
			log.Printf("error reading slave record for deletion: %s\n", err)
			break
		}

		nextAddress := model.Next

		clear(model.IssuedTo[:])
		model.Next = -1
		model.Presence = false

		r.App.Slave.Junk = append(r.App.Slave.Junk, uint32(address))
		r.App.Slave.Indices = utils.RemoveIndex(r.App.Slave.Indices, model.ID)

		err = driver.WriteModel(flFile, &model, address, io.SeekStart)
		if err != nil {
			log.Println("error updating slave record to mark as deleted:", err)
			break
		}

		log.Printf("deleted slave record at address %d\n", address)

		address = nextAddress
	}
}
