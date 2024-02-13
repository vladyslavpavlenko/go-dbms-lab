package utils

import (
	"fmt"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver"
	"io"
	"log"
	"os"
	"sort"
	"strings"
)

// AddMasterIndex adds an index and a corresponding address of the new entry to Master.
func AddMasterIndex(indices []driver.IndexTable, id uint32, address uint32) []driver.IndexTable {
	entry := driver.IndexTable{
		Index:   id,
		Address: address,
	}

	indices = append(indices, entry)
	fmt.Printf("Added a new record with id [%d] at address [%d]\n", id, address)

	return indices
}

// WriteMasterIndices writes indices to the .ind file.
func WriteMasterIndices(indFile *os.File, indices []driver.IndexTable) {
	sorted := SortIndices(indices)

	err := driver.WriteModel(indFile, sorted, 0, io.SeekStart)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Wrote Master IND")
}

// RecordExists checks if a record with the given ID already exists in the master table.
func RecordExists(indices []driver.IndexTable, id uint32) bool {
	for _, entry := range indices {
		if id == entry.Index {
			return true
		}
	}

	return false
}

// SortIndices sorts entries of an IndexTable in ascending order by their indices.
func SortIndices(indices []driver.IndexTable) []driver.IndexTable {
	sort.Slice(indices, func(i int, j int) bool {
		return indices[i].Index < indices[j].Index
	})

	log.Println("Sorted!")
	log.Println(indices)

	return indices
}

// NumberOfRecords return the number of records in a table using index table.
func NumberOfRecords(indices []driver.IndexTable) int {
	return len(indices)
}

// NumberOfSubrecords TODO
func NumberOfSubrecords(indices []driver.IndexTable, id uint32) int {
	return 0
}

// ByteArrayToString converts a byte array to a string.
func ByteArrayToString(bytes []byte) string {
	return strings.TrimRight(string(bytes), "\x00")
}
