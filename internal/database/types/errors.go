package types

import "errors"

var (
	// ErrNoDocuments is returned when no documents are found
	ErrNoDocuments = errors.New("no documents found")
	
	// ErrDuplicateKey is returned when a unique constraint is violated
	ErrDuplicateKey = errors.New("duplicate key error")
	
	// ErrInvalidID is returned when an ID format is invalid
	ErrInvalidID = errors.New("invalid ID format")
	
	// ErrNotConnected is returned when database is not connected
	ErrNotConnected = errors.New("database not connected")
	
	// ErrTransactionFailed is returned when a transaction fails
	ErrTransactionFailed = errors.New("transaction failed")
	
	// ErrUnsupportedOperation is returned when an operation is not supported by the adapter
	ErrUnsupportedOperation = errors.New("operation not supported by this database adapter")
)