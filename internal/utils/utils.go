package utils

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
)

// ReadModel reads data from a file to a model.
func ReadModel(r io.Reader, model any) {
	err := binary.Read(r, binary.BigEndian, model)
	if err != nil {
		log.Fatal("file reading failed: ", err)
	}
}

// WriteModel writes model to a binary .fl file.
func WriteModel(w io.Writer, model any) error {
	var binBuf bytes.Buffer
	err := binary.Write(&binBuf, binary.BigEndian, model)
	if err != nil {
		return errors.New(fmt.Sprintf("error writing binary representation: %s", err))
	}

	_, err = w.Write(binBuf.Bytes())
	if err != nil {
		return errors.New(fmt.Sprintf("error writing to file: %s", err))
	}

	return nil
}
