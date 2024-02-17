package driver

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// IndexTable defines the structure for index table entries, holding an index and its corresponding address.
type IndexTable struct {
	Index   uint32
	Address uint32
}

// Table encapsulates file connections and indices for a table, along with the size of its model.
type Table struct {
	FL      *os.File     // File connection for data
	IND     *os.File     // File connection for index
	Indices []IndexTable // List of indices
	Size    int          // Size of the model stored in the table
}

// NewTable initializes a new Table instance with given file connections and model size.
func NewTable(fl *os.File, ind *os.File, indices []IndexTable, model any) *Table {
	indices, _ = LoadIndices(ind)
	return &Table{
		FL:      fl,
		IND:     ind,
		Indices: indices,
		Size:    binary.Size(model),
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

	// reserve a place for garbage
	err = WriteModel(flFile, model, 0, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("error reserving a place for garbage: %w", err)
	}

	table := NewTable(flFile, indFile, []IndexTable{}, model)
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

	return indices, nil
}
