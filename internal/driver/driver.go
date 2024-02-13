package driver

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
)

// IndexTable holds fields of an index table.
type IndexTable struct {
	Index   uint32
	Address uint32
}

// Table holds file connections and data needed for a table.
type Table struct {
	FL      *os.File
	IND     *os.File
	Indices []IndexTable
	Offset  uint32
}

// NewTable creates a new Table.
func NewTable(fl *os.File, ind *os.File, indices []IndexTable) *Table {
	return &Table{
		FL:      fl,
		IND:     ind,
		Indices: indices,
	}
}

// CreateTable creates both .fl and .ind files to represent a table and returns a Table.
func CreateTable(name string) (*Table, error) {
	flName := fmt.Sprintf("%s.fl", name)
	flFile, err := os.OpenFile(flName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error creating .fl file: %s", err))
	}

	indName := fmt.Sprintf("%s.ind", name)
	indFile, err := os.OpenFile(indName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error creating .ind file: %s", err))
	}

	conn := NewTable(flFile, indFile, []IndexTable{})

	return conn, err
}

// ReadModel reads data from a file to a model.
func ReadModel(file *os.File, model any) error {
	err := binary.Read(file, binary.BigEndian, model)
	if err != nil {
		return errors.New(fmt.Sprintf("file reading failed: %s", err))
	}

	return nil
}

// WriteModel writes model to a binary .fl file.
func WriteModel(file *os.File, model any, offset int64, whence int) error {
	if _, err := file.Seek(offset, whence); err != nil {
		log.Fatal(err)
	}

	var binBuf bytes.Buffer
	err := binary.Write(&binBuf, binary.BigEndian, model)
	if err != nil {
		return errors.New(fmt.Sprintf("error writing binary representation: %s", err))
	}

	_, err = file.Write(binBuf.Bytes())
	if err != nil {
		return errors.New(fmt.Sprintf("error writing to file: %s", err))
	}

	return nil
}
