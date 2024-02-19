package models

// Master represents a master record.
type Master struct {
	FirstSlaveAddress int64
	Presence          bool
}

// Slave represents a slave record.
type Slave struct {
	Previous int64
	Next     int64
	Presence bool
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
