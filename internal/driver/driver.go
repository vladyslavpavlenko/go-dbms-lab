package driver

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
)

// IndexTable defines the structure for index table entries, holding an index and its corresponding address.
type IndexTable struct {
	Index   uint32
	Address uint32
}

// Table encapsulates file connection and indices for a table, along with the size of its model.
type Table struct {
	FL      *os.File     // File connection for data
	Indices []IndexTable // List of indices
	Junk    []uint32     // List of addresses of unused space
	Size    int          // Size of the model stored in the table
}

// NewTable initializes a new Table instance with given file connections and model size.
func NewTable(fl *os.File, ind *os.File, jk *os.File, model any) *Table {
	size := binary.Size(model)

	indices, err := LoadIndices(ind)
	if err != nil {
		log.Fatal(err)
	}

	junk, err := LoadJunk(jk)
	if err != nil {
		log.Fatal(err)
	}

	return &Table{
		FL:      fl,
		Indices: indices,
		Junk:    junk,
		Size:    size,
	}
}

// CreateTable creates files for a new table (.fl and .ind) based on the given name and model, returning the Table instance.
func CreateTable(name string, model any) (*Table, error) {
	flName := fmt.Sprintf("%s.fl", name)
	flFile, err := os.OpenFile(flName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("error creating .fl file: %w", err)
	}

	indName := fmt.Sprintf("%s.ind", name)
	indFile, err := os.OpenFile(indName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("error creating .ind file: %w", err)
	}
	defer indFile.Close()

	jkName := fmt.Sprintf("%s.jk", name)
	jkFile, err := os.OpenFile(jkName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("error creating .ind file: %w", err)
	}
	defer jkFile.Close()

	table := NewTable(flFile, indFile, jkFile, model)
	return table, nil
}

// ReadModel reads a model from the specified file at a given offset and position.
func ReadModel(file *os.File, model any, offset int64, whence int) error {
	if _, err := file.Seek(offset, whence); err != nil {
		return err
	}
	return binary.Read(file, binary.BigEndian, model)
}

// WriteModel writes a model's binary representation to a file at the specified offset and position.
func WriteModel(file *os.File, model any, offset int64, whence int) error {
	if _, err := file.Seek(offset, whence); err != nil {
		return fmt.Errorf("error seeking file: %w", err)
	}

	var binBuf bytes.Buffer
	if err := binary.Write(&binBuf, binary.BigEndian, model); err != nil {
		return fmt.Errorf("error writing binary representation: %w", err)
	}

	if _, err := file.Write(binBuf.Bytes()); err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}

// LoadIndices reads IndexTable entries from an .ind file, initializing the table's indices.
func LoadIndices(indFile *os.File) ([]IndexTable, error) {
	if _, err := indFile.Seek(0, io.SeekStart); err != nil {
		fmt.Printf("error reading data: %s\n", err)
		return nil, err
	}

	var indices []IndexTable
	for {
		var model IndexTable
		err := ReadModel(indFile, &model, 0, io.SeekCurrent)
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Printf("error reading data: %s\n", err)
			return nil, err
		}
		indices = append(indices, model)
	}

	log.Println("indices loaded: ", indices)

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
			fmt.Printf("error reading data: %s\n", err)
			return nil, err
		}
		junk = append(junk, address)
	}

	log.Println("junk loaded: ", junk)

	return junk, nil
}
