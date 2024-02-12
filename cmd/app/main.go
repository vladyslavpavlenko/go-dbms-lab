package main

import (
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/models"
	"github.com/vladyslavpavlenko/go-dbms-lab/internal/utils"
	"io"
	"log"
	"os"
)

func main() {
	// open file
	file, err := os.OpenFile("courses.fl", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// write to file
	//log.Println("Writing to file...")
	//
	//var c1 models.Certificate
	//c1.ID = 1
	//c1.CourseID = 1
	//copy(c1.IssuedTo[:], "Vadym Ripa")
	//c1.Presence = true
	//c1.Next = -1
	//
	//err = utils.WriteModel(file, c1)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//log.Println("Done!")
	//
	//// read from file
	//log.Println("Reading from file...")
	//
	//var c2 models.Certificate
	//
	//utils.ReadModel(file, &c2)
	//
	//log.Println(c2)
	//
	//log.Println("Done!")

	// write to file
	log.Println("Writing to file...")

	var c1 models.Course
	c1.ID = 1
	copy(c1.Title[:], "Go DBMS Development")
	copy(c1.Instructor[:], "Vadym Ripa")
	copy(c1.Category[:], "Go")
	c1.FirstCertificateID = 1
	c1.Presence = true

	ms := []models.Course{
		{
			ID: 1,
		},
		{
			ID: 2,
		},
	}

	err = utils.WriteModel(file, ms)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		log.Fatal(err)
	}

	// read from file
	log.Println("Reading from file...")

	var c2 models.Course

	utils.ReadModel(file, &c2)

	log.Println(c2)

	log.Println("Done!")
}
