package cmd

import "os"

// FromStdin indicates whether input is being piped to us or we should parse
// args.
func FromStdin() bool {
	stat, _ := os.Stdin.Stat()
	return stat.Mode()&os.ModeCharDevice == 0
}
