package driver

import (
	"fmt"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/models"
	"io"
	"log"
	"os"
	"sort"
	"strings"
)

// AddIndex appends a new index and its corresponding address to the indices list, then sorts the list.
func AddIndex(indices []IndexTable, id uint32, address uint32) []IndexTable {
	entry := IndexTable{
		Index:   id,
		Address: address,
	}
	indices = append(indices, entry)
	return SortIndices(indices)
}

// RemoveIndex removes an index and its corresponding address from the indices list, then sorts the list.
func RemoveIndex(indices []IndexTable, id uint32) []IndexTable {
	for i, entry := range indices {
		if entry.Index == id {
			indices = append(indices[:i], indices[i+1:]...)
			break
		}
	}
	return indices
}

// UpdateAddress updates the address of the entry in the indices list, then sorts the list.
func UpdateAddress(indices []IndexTable, id uint32, newAddress uint32) []IndexTable {
	for i, entry := range indices {
		if entry.Index == id {
			indices[i].Address = newAddress
			break
		}
	}
	return indices
}

// WriteIndices writes the sorted indices to the specified .ind file.
func WriteIndices(indFile *os.File, indices []IndexTable) {
	sorted := SortIndices(indices)

	if err := indFile.Truncate(0); err != nil {
		log.Printf("error truncating file: %v\n", err)
		return
	}

	if _, err := indFile.Seek(0, io.SeekStart); err != nil {
		log.Printf("error seeking file: %v\n", err)
		return
	}

	if err := WriteModel(indFile, sorted, 0, io.SeekStart); err != nil {
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

	if err := WriteModel(jkFile, junk, 0, io.SeekStart); err != nil {
		log.Printf("error writing junk: %v\n", err)
	} else {
		log.Printf("%s written successfully.\n", jkFile.Name())
	}
}

// RecordExists checks for the existence of a record with the specified ID in the master table.
func RecordExists(indices []IndexTable, id uint32) bool {
	for _, entry := range indices {
		if id == entry.Index {
			return true
		}
	}
	return false
}

// SortIndices sorts the IndexTable entries in ascending order by index.
func SortIndices(indices []IndexTable) []IndexTable {
	sort.Slice(indices, func(i, j int) bool { return indices[i].Index < indices[j].Index })
	return indices
}

// NumberOfRecords calculates the total number of records using the index table.
func NumberOfRecords(indices []IndexTable) int {
	return len(indices)
}

// NumberOfSubrecords calculates the number of subrecords, optionally associated with a given ID.
func NumberOfSubrecords(flFile *os.File, firstSlaveAddress int64) int {
	count := 0
	nextAddress := firstSlaveAddress

	for nextAddress != NoLink {
		var slave models.Certificate
		err := ReadModel(flFile, &slave, nextAddress, io.SeekStart)
		if err != nil {
			fmt.Printf("error reading slave model: %s\n", err)
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

// GetAddressByIndex performs a binary search on the indices slice to find the address associated with the specified ID.
func GetAddressByIndex(indices []IndexTable, id uint32) (uint32, bool) {
	i := sort.Search(len(indices), func(i int) bool { return indices[i].Index >= id })
	if i < len(indices) && indices[i].Index == id {
		return indices[i].Address, true
	}
	return 0, false
}

// WriteServiceData writes index table and junk addresses to the service files (.ind, .jk).
func WriteServiceData(fileName string, indices []IndexTable, junk []uint32, withJunk bool) error {
	indName := fmt.Sprintf("%s.ind", fileName)
	indFile, err := os.OpenFile(indName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("error creating .ind file: %w", err)
	}
	defer indFile.Close()
	WriteIndices(indFile, indices)

	if withJunk {
		jkName := fmt.Sprintf("%s.jk", fileName)
		jkFile, err := os.OpenFile(jkName, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return fmt.Errorf("error creating .ind file: %w", err)
		}
		defer jkFile.Close()
		WriteJunk(jkFile, junk)
	}

	return nil
}

// RequiresCompaction checks if the total size of the junk exceeds a predefined threshold.
func (t *Table) RequiresCompaction() bool {
	totalJunkSize := len(t.Junk)
	return totalJunkSize >= MaxJunkSize
}

// LoadIndices reads IndexTable entries from an .ind file, initializing the table's indices.
func LoadIndices(indFile *os.File) ([]IndexTable, error) {
	if _, err := indFile.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("error seeking file: %w", err)
	}

	var indices []IndexTable
	for {
		var model IndexTable
		err := ReadModel(indFile, &model, 0, io.SeekCurrent)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("error reading data: %w", err)
		}
		indices = append(indices, model)
	}

	return indices, nil
}

// LoadJunk reads junk addresses from a .jk file, initializing the unused space slice.
func LoadJunk(jkFile *os.File) ([]uint32, error) {
	if _, err := jkFile.Seek(0, io.SeekStart); err != nil {
		fmt.Printf("error reading data: %s\n", err)
		return nil, err
	}

	var junk []uint32
	for {
		var address uint32
		err := ReadModel(jkFile, &address, 0, io.SeekCurrent)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("error reading data: %w", err)
		}
		junk = append(junk, address)
	}

	return junk, nil
}

// GetLastRecordAddress returns the address of the last record in the master file.
func GetLastRecordAddress(indices []IndexTable) (uint32, bool) {
	if len(indices) == 0 {
		log.Println("no records found in the index table")
		return 0, false
	}
	lastIndex := indices[len(indices)-1]
	return lastIndex.Address, true
}
