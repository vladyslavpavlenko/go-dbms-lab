package driver

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/models"
	"io"
	"log"
	"os"
	"sort"
)

const (
	MaxJunkSize = 2
	NoLink      = -1
)

// IndexTable defines the structure for index table entries, holding an index and its corresponding address.
type IndexTable struct {
	Index   uint32
	Address uint32
}

// Table encapsulates file connection and indices for a table, along with the size of its model.
type Table struct {
	FL      *os.File
	Indices []IndexTable
	Junk    []uint32
	Size    int
}

// NewTable initializes a new Table instance with given file connections and model size.
func NewTable(fl *os.File, ind *os.File, jk *os.File, model any, withJunk bool) *Table {
	size := binary.Size(model)

	indices, err := LoadIndices(ind)
	if err != nil {
		log.Fatal(err)
	}

	var junk []uint32
	if withJunk {
		junk, err = LoadJunk(jk)
		if err != nil {
			log.Fatal(err)
		}
	}

	return &Table{
		FL:      fl,
		Indices: indices,
		Junk:    junk,
		Size:    size,
	}
}

// CreateTable creates files for a new table (.fl and .ind) based on the given name and model, returning the Table instance.
func CreateTable(name string, model any, withJunk bool) (*Table, error) {
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

	var jkFile *os.File

	if withJunk {
		jkName := fmt.Sprintf("%s.jk", name)
		jkFile, err = os.OpenFile(jkName, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return nil, fmt.Errorf("error creating .ind file: %w", err)
		}
		defer jkFile.Close()
	}

	table := NewTable(flFile, indFile, jkFile, model, withJunk)
	return table, nil
}

// ReadModel reads a model from the specified file at a given offset and position.
func ReadModel(file *os.File, model any, offset int64, whence int) error {
	if _, err := file.Seek(offset, whence); err != nil {
		return fmt.Errorf("error reading model: %w", err)
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
		return fmt.Errorf("error writing model: %w", err)
	}

	if _, err := file.Write(binBuf.Bytes()); err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}

// MoveModel moves a model from one address to another.
func MoveModel(flFile *os.File, model any, oldAddress, newAddress int64) error {
	err := ReadModel(flFile, model, oldAddress, io.SeekStart)
	if err != nil {
		return fmt.Errorf("error reading last record: %w", err)
	}

	if err := WriteModel(flFile, model, newAddress, io.SeekStart); err != nil {
		return fmt.Errorf("error moving last record: %w", err)
	}

	return nil
}

// CompactSlaveFile handles slave file compaction.
func CompactSlaveFile(flFile *os.File, indices []IndexTable, junk []uint32) ([]uint32, error) {
	sort.Slice(junk, func(i, j int) bool {
		return junk[i] < junk[j]
	})

	sort.Slice(indices, func(i, j int) bool {
		return indices[i].Address > indices[j].Address
	})

	var model models.Certificate

	for i := 0; i < len(indices) && i < len(junk) && indices[i].Address > junk[i]; i++ {
		err := MoveModel(flFile, &model, int64(indices[i].Address), int64(junk[i]))
		if err != nil {
			return nil, err
		}

		UpdateAddress(indices, indices[i].Index, junk[i])

		err = updateLinkedListPointers(flFile, &model, junk[i])
		if err != nil {
			return nil, fmt.Errorf("error updating linked list pointers: %w", err)
		}
	}

	err := TruncateFile(flFile, int64(len(indices)*binary.Size(model)))
	if err != nil {
		return nil, fmt.Errorf("error trancating file: %w", err)
	}

	junk = junk[:0]

	return junk, nil
}

// updateLinkedListPointers updates Next and Previous pointers of a node's neighboring nodes to its new address.
func updateLinkedListPointers(flFile *os.File, model *models.Certificate, newAddress uint32) error {
	// update the previous node's next pointer
	if model.Previous != NoLink {
		var prevModel models.Certificate
		err := ReadModel(flFile, &prevModel, model.Previous, io.SeekStart)
		if err != nil {
			return err
		}
		prevModel.Next = int64(newAddress)
		err = WriteModel(flFile, &prevModel, model.Previous, io.SeekStart)
		if err != nil {
			return err
		}
	}

	// update the next node's previous pointer
	if model.Next != NoLink {
		var nextModel models.Certificate
		err := ReadModel(flFile, &nextModel, model.Next, io.SeekStart)
		if err != nil {
			return err
		}
		nextModel.Previous = int64(newAddress)
		err = WriteModel(flFile, &nextModel, model.Next, io.SeekStart)
		if err != nil {
			return err
		}
	}

	return nil
}

// TruncateFile truncates the given file to a specific length.
func TruncateFile(file *os.File, address int64) error {
	err := file.Truncate(address)
	if err != nil {
		return fmt.Errorf("error truncating file: %w", err)
	}

	return nil
}
