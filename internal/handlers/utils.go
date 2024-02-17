package handlers

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver/utils"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/models"
	"io"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// printMasterData prints the master table data to stdout. If includeDetails is true, additional fields
// are included in the output.
func printMasterData(flFile *os.File, includeDetails bool) {
	if _, err := flFile.Seek(0, io.SeekStart); err != nil {
		fmt.Printf("error seeking file: %s\n", err)
		return
	}

	var model models.Course
	var data []models.Course

	for {
		err := driver.ReadModel(flFile, &model, 0, io.SeekCurrent)
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Printf("error reading data: %s\n", err)
			return
		}

		if includeDetails {
			if !model.Presence {
				data = append(data, model)
				continue
			}
		}

		data = append(data, model)
	}

	sort.Slice(data, func(i, j int) bool { return data[i].ID < data[j].ID })

	headers := []string{"ID", "TITLE", "CATEGORY", "INSTRUCTOR"}
	if includeDetails {
		headers = append(headers, "FS_ID", "PRESENCE")
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for _, entry := range data {
		stringID := strconv.Itoa(int(entry.ID))
		stringTitle := utils.ByteArrayToString(entry.Title[:])
		stringCategory := utils.ByteArrayToString(entry.Category[:])
		stringInstructor := utils.ByteArrayToString(entry.Instructor[:])

		row := []string{stringID, stringTitle, stringCategory, stringInstructor}
		if includeDetails {
			row = append(row, fmt.Sprintf("%v", entry.FirstSlaveID), fmt.Sprintf("%v", entry.Presence))
		}

		table.Append(row)
	}

	table.Render()
}

// printMasterQuery prints selected fields from the master table based on provided field queries. If all is true,
// all records are printed.
func printMasterQuery(flFile *os.File, offset int64, queries []string, all bool) {
	if all {
		_, err := flFile.Seek(0, io.SeekStart)
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
			case "FS_ID":
				headers = append(headers, "FS_ID")
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
			case "FS_ID":
				row = append(row, strconv.Itoa(int(model.FirstSlaveID)))
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
func printSlaveQuery(flFile *os.File, id int, firstID int64, queries []string, all bool) {
	_, err := flFile.Seek(0, io.SeekStart)
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
			case "ISSUED_TO", "NEXT", "PRESENCE":
				headers = append(headers, query)
			}
		}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	var offset = firstID
	if firstID == -1 {
		offset = 0
	}

	for {
		var model models.Certificate

		err := driver.ReadModel(flFile, &model, offset, io.SeekCurrent)
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Printf("Error reading slave data: %s\n", err)
			return
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

// printSlaveData prints the slave table data to stdout. If includeDetails is true, additional fields
// are included in the output.
func printSlaveData(flFile *os.File, includeDetails bool) {
	if _, err := flFile.Seek(0, io.SeekStart); err != nil {
		fmt.Printf("error seeking file: %s\n", err)
		return
	}

	var model models.Certificate
	var data []models.Certificate

	for {
		err := driver.ReadModel(flFile, &model, 0, io.SeekCurrent)
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Printf("error reading data: %s\n", err)
			return
		}

		if includeDetails {
			if !model.Presence {
				data = append(data, model)
				continue
			}
		}

		data = append(data, model)
	}

	sort.Slice(data, func(i, j int) bool { return data[i].ID < data[j].ID })

	headers := []string{"ID", "COURSE_ID", "ISSUED_TO"}
	if includeDetails {
		headers = append(headers, "NEXT", "PRESENCE")
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for _, entry := range data {
		stringID := strconv.Itoa(int(entry.ID))
		stringCourseID := strconv.Itoa(int(entry.CourseID))
		stringIssuedTo := utils.ByteArrayToString(entry.IssuedTo[:])

		row := []string{stringID, stringCourseID, stringIssuedTo}
		if includeDetails {
			row = append(row, fmt.Sprintf("%v", entry.Next), fmt.Sprintf("%v", entry.Presence))
		}

		table.Append(row)
	}

	table.Render()
}
