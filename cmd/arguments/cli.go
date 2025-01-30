package arguments

type CliArguments struct {
	ConfigPath string

	// The debug logger's verbosity level. Must be a value between 0 (All) and 5 (Fatal only). Default - 4 (Error/Fatal)
	Verbosity uint8

	NoBanner bool
}

type ExecuteWithArgsResult struct {
	Error  error
	StdOut string
	StdErr string
	Logs   string
}
