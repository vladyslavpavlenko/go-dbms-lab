package driver

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
)

// MaxJunkSize defines the maximum allowed size of the junk before recommending compaction.
const MaxJunkSize int = 12

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

// CompactMasterFile handles master file compaction.
//func CompactMasterFile(flFile, indFile, jkFile *os.File) error {
//	var model models.Course
//	var data []models.Course
//
//	// Read all the entries from the flFile
//	for {
//		err := ReadModel(flFile, &model, 0, io.SeekCurrent)
//		if err == io.EOF {
//			break
//		} else if err != nil {
//			fmt.Printf("error reading data: %s\n", err)
//			return err
//		}
//
//		if !model.Presence {
//			continue
//		}
//
//		data = append(data, model)
//	}
//
//	var index IndexTable
//	var data []models.Course
//
//	for {
//		err := ReadModel(indFile, &index, 0, io.SeekStart)
//		if err == io.EOF {
//			break
//		} else if err != nil {
//			fmt.Printf("error reading index: %s\n", err)
//			return err
//		}
//
//		// Process index to decide on compaction, perhaps adjusting `data` slice or planning rewrites
//	}
//
//	// Assuming jkFile is read to identify spaces for reuse, implement as needed
//
//	// Truncate the original file to rewrite compacted data
//	err := flFile.Truncate(0)
//	if err != nil {
//		fmt.Printf("error truncating file: %s\n", err)
//		return err
//	}
//
//	// Reset file pointer to the beginning to start rewriting data
//	_, err = flFile.Seek(0, io.SeekStart)
//	if err != nil {
//		fmt.Printf("error seeking in file: %s\n", err)
//		return err
//	}
//
//	for _, model := range data {
//		err := WriteModel(flFile, &model, 0, io.SeekStart)
//		if err != nil {
//			fmt.Printf("error writing model: %s\n", err)
//			return err
//		}
//	}
//
//	return nil
//}

// TruncateFile truncates the given file to a specific length.
func TruncateFile(file *os.File, address int64) error {
	err := file.Truncate(address)
	if err != nil {
		fmt.Printf("error truncating file: %v\n", err)
		return err
	}

	return nil
}
