package models

// Course is the course model.
type Course struct {
	ID                 uint32
	Title              [50]byte
	Category           [15]byte
	Instructor         [30]byte
	FirstCertificateID uint32
	Presence           bool
}

// Certificate is the certificate model.
type Certificate struct {
	ID       uint32
	CourseID uint32
	IssuedTo [30]byte
	Presence bool
	Next     int64
}
