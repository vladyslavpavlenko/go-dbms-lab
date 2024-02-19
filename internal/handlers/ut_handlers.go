package handlers

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/models"
	"io"
	"os"
	"sort"
	"strconv"
)

// UtMaster handles printing of all entries in the master table, including detailed information.
func (r *Repository) UtMaster(_ *cobra.Command, _ []string) {
	flFile := r.App.Master.FL

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

		data = append(data, model)
	}

	sort.Slice(data, func(i, j int) bool { return data[i].ID < data[j].ID })

	headers := []string{"ID", "TITLE", "CATEGORY", "INSTRUCTOR", "FS_ADDRESS", "PRESENCE"}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for _, entry := range data {
		stringID := strconv.Itoa(int(entry.ID))
		stringTitle := driver.ByteArrayToString(entry.Title[:])
		stringCategory := driver.ByteArrayToString(entry.Category[:])
		stringInstructor := driver.ByteArrayToString(entry.Instructor[:])

		row := []string{stringID, stringTitle, stringCategory, stringInstructor}
		row = append(row, fmt.Sprintf("%v", entry.FirstSlaveAddress), fmt.Sprintf("%v", entry.Presence))

		table.Append(row)
	}

	table.Render()
}

// UtSlave handles printing of all entries in the slave table, including detailed information.
func (r *Repository) UtSlave(_ *cobra.Command, _ []string) {
	flFile := r.App.Slave.FL
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

		data = append(data, model)
	}

	sort.Slice(data, func(i, j int) bool { return data[i].ID < data[j].ID })

	headers := []string{"ID", "COURSE_ID", "ISSUED_TO", "PREVIOUS", "NEXT", "PRESENCE"}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for _, entry := range data {
		stringID := strconv.Itoa(int(entry.ID))
		stringCourseID := strconv.Itoa(int(entry.CourseID))
		stringIssuedTo := driver.ByteArrayToString(entry.IssuedTo[:])

		row := []string{stringID, stringCourseID, stringIssuedTo}
		row = append(row, fmt.Sprintf("%v", entry.Previous), fmt.Sprintf("%v", entry.Next), fmt.Sprintf("%v", entry.Presence))

		table.Append(row)
	}

	table.Render()
}
