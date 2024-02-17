package models

type Master struct {
	FirstSlaveAddress int64
	Presence          bool
}

type Slave struct {
	Presence bool
	Next     int64
}

// Course is the course model.
type Course struct {
	ID         uint32
	Title      [50]byte
	Category   [15]byte
	Instructor [30]byte
	Master
}

// Certificate is the certificate model.
type Certificate struct {
	ID       uint32
	CourseID uint32
	IssuedTo [30]byte
	Slave
}
