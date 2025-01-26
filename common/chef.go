package common

type Chef interface {
	// Creates or updates the stack.
	Up() error

	// Previews the creation or update of the stack.
	Preview() error

	// Deletes the stack and all its resources.
	Down() error

	// Deletes the stack and its resources and the underlying state store, force is used to delete non-empty state stores.
	Destroy(force bool) error

	// Return the project name.
	ProjectName() string

	// Add recipes to the stack.
	Append(recipes ...Recipe)

	// Compare resource state with the state known to exist in the actual cloud provider and update the Pulumi stack if needed
	Refresh() error

	// Return output of a stack
	Results() (map[string]interface{}, error)

	// Return deployments/updates hitory
	History() error
}
