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

// AddIndex appends a new index and its corresponding address to the master indices list, then sorts the list.
func AddIndex(indices []driver.IndexTable, id uint32, address uint32) []driver.IndexTable {
	entry := driver.IndexTable{
		Index:   id,
		Address: address,
	}
	indices = append(indices, entry)
	fmt.Printf("Added a new record with ID [%d] at address [%d].\n", id, address)
	return SortIndices(indices)
}

// WriteIndices writes the sorted indices to the specified .ind file.
func WriteIndices(indFile *os.File, indices []driver.IndexTable) {
	sorted := SortIndices(indices)
	if err := driver.WriteModel(indFile, sorted, 0, io.SeekStart); err != nil {
		log.Printf("Error writing indices: %v\n", err)
	} else {
		log.Printf("%s written successfully.\n", indFile.Name())
	}
}

// RecordExists checks for the existence of a record with the specified ID in the master table.
func RecordExists(indices []driver.IndexTable, id uint32) bool {
	for _, entry := range indices {
		if id == entry.Index {
			return true
		}
	}
	return false
}

// SortIndices sorts the IndexTable entries in ascending order by index.
func SortIndices(indices []driver.IndexTable) []driver.IndexTable {
	sort.Slice(indices, func(i, j int) bool { return indices[i].Index < indices[j].Index })
	log.Println("Indices sorted.")
	return indices
}

// NumberOfRecords calculates the total number of records using the index table.
func NumberOfRecords(indices []driver.IndexTable) int {
	return len(indices)
}

// NumberOfSubrecords (placeholder) calculates the number of subrecords associated with a given ID.
func NumberOfSubrecords(indices []driver.IndexTable, id uint32) int {
	// Implementation pending.
	return 0
}

// ByteArrayToString converts a byte array into a string, trimming trailing null bytes.
func ByteArrayToString(bytes []byte) string {
	return strings.TrimRight(string(bytes), "\x00")
}

// GetAddressByIndex performs a binary search on the indices slice to find the address associated with the specified index.
func GetAddressByIndex(indices []driver.IndexTable, id uint32) (uint32, bool) {
	i := sort.Search(len(indices), func(i int) bool { return indices[i].Index >= id })
	if i < len(indices) && indices[i].Index == id {
		return indices[i].Address, true
	}
	return 0, false
}
