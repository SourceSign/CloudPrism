package common

type StateStore interface {
	// Creates and/or logs in to a state store and returns its URL string.
	StoreOpen() (string, error)

	// Closes or logs out of a state store, without deleting any data.
	StoreClose() error

	// Deletes the state store, including all data when the force parameter is true.
	StoreDelete(force bool) error
}
