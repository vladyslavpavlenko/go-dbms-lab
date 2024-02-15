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
	"strconv"
	"strings"
)

// printMasterData prints the master table data to stdout. If includeDetails is true, additional fields
// are included in the output.
func printMasterData(flFile *os.File, includeDetails bool) {
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

		if includeDetails {
			if !model.Presence {
				data = append(data, model)
				continue
			}
		}

		data = append(data, model)
	}

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
		flFile.Seek(0, io.SeekStart)
	} else {
		flFile.Seek(offset, io.SeekStart)
	}

	headers := []string{"ID"}

	if len(queries) != 0 {
		for _, query := range queries {
			fieldName := strings.ToUpper(query)

			switch fieldName {
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
				fmt.Fprintf(os.Stderr, "field '%s' was not found\n", query)
			}
		}
	} else {
		headers = append(headers, "TITLE", "CATEGORY", "INSTRUCTOR")
	}

	if len(headers) == 1 && !slices.Contains(queries, "id") {
		fmt.Fprintln(os.Stderr, "nothing to show")
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
			fmt.Fprintf(os.Stderr, "error reading data: %s\n", err)
			return
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
