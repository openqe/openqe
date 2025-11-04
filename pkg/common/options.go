package common

// GlobalOptions contains global flags that are available to all commands
type GlobalOptions struct {
	// Verbose enables debug-level logging
	Verbose bool

	// Yes automatically confirms all interactive prompts
	Yes bool
}

// DefaultGlobalOptions returns a new GlobalOptions with default values
func DefaultGlobalOptions() *GlobalOptions {
	return &GlobalOptions{
		Verbose: false,
		Yes:     false,
	}
}
