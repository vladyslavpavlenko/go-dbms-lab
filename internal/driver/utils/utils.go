package utils

import (
	"fmt"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/driver"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/models"
	"io"
	"log"
	"os"
	"sort"
	"strings"
)

// AddIndex appends a new index and its corresponding address to the indices list, then sorts the list.
func AddIndex(indices []driver.IndexTable, id uint32, address uint32) []driver.IndexTable {
	entry := driver.IndexTable{
		Index:   id,
		Address: address,
	}
	indices = append(indices, entry)
	log.Printf("added a new record with ID %d at address %d.\n", id, address)
	return SortIndices(indices)
}

// RemoveIndex removes an index and its corresponding address from the indices list, then sorts the list.
func RemoveIndex(indices []driver.IndexTable, id uint32) []driver.IndexTable {
	for i, entry := range indices {
		if entry.Index == id {
			indices = append(indices[:i], indices[i+1:]...)
			log.Printf("removed record with ID %d from index table\n", id)
			break
		}
	}
	return indices
}

// WriteIndices writes the sorted indices to the specified .ind file.
func WriteIndices(indFile *os.File, indices []driver.IndexTable) {
	sorted := SortIndices(indices)

	if err := indFile.Truncate(0); err != nil {
		log.Printf("error truncating file: %v\n", err)
		return
	}

	if _, err := indFile.Seek(0, io.SeekStart); err != nil {
		log.Printf("error seeking file: %v\n", err)
		return
	}

	if err := driver.WriteModel(indFile, sorted, 0, io.SeekStart); err != nil {
		log.Printf("error writing indices: %v\n", err)
	} else {
		log.Printf("%s written successfully.\n", indFile.Name())
	}
}

// WriteJunk writes the junk addresses to the specified .jk file.
func WriteJunk(jkFile *os.File, junk []uint32) {
	if err := jkFile.Truncate(0); err != nil {
		log.Printf("error truncating file: %v\n", err)
		return
	}

	if _, err := jkFile.Seek(0, io.SeekStart); err != nil {
		log.Printf("error seeking file: %v\n", err)
		return
	}

	if err := driver.WriteModel(jkFile, junk, 0, io.SeekStart); err != nil {
		log.Printf("error writing junk: %v\n", err)
	} else {
		log.Printf("%s written successfully.\n", jkFile.Name())
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

// NumberOfSubrecords calculates the number of subrecords, optionally associated with a given ID.
func NumberOfSubrecords(flFile *os.File, firstSlaveAddress int64) int {
	count := 0
	nextAddress := firstSlaveAddress

	for nextAddress != -1 {
		var slave models.Certificate
		err := driver.ReadModel(flFile, &slave, int64(nextAddress), io.SeekStart)
		if err != nil {
			fmt.Printf("error reading slave model: %v\n", err)
			break
		}
		if slave.Presence {
			count++
		}
		nextAddress = slave.Next
	}

	return count
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

// WriteServiceData writes index table and junk addresses to the service files (.ind, .jk).
func WriteServiceData(fileName string, indices []driver.IndexTable, junk []uint32) error {
	indName := fmt.Sprintf("%s.ind", fileName)
	indFile, err := os.OpenFile(indName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("error creating .ind file: %w", err)
	}
	defer indFile.Close()

	jkName := fmt.Sprintf("%s.jk", fileName)
	jkFile, err := os.OpenFile(jkName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("error creating .ind file: %w", err)
	}
	defer jkFile.Close()

	WriteIndices(indFile, indices)
	WriteJunk(jkFile, junk)

	return nil
}
