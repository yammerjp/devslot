package command

import "io"

// Context provides shared resources to commands
type Context struct {
	Writer io.Writer
}
