package records

// Package records implements the Mesos variant of RecordIO decoding, whereby
// each record is prefixed by a line that indicates the length (decimal, printed
// in ASCII) of the record. The octets of the record immediately follow the
// length-line. Zero-length records are allowed.
//
// This package does not enforce any particular record format: that choice is
// left up to the caller in the form of the Unmarshaler.
