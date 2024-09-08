package logger

import (
	"io"
	"os"

	zlog "github.com/rs/zerolog"
	"github.com/samber/mo"
)

func New(writer mo.Option[io.Writer]) zlog.Logger {
	return zlog.New(writer.OrElse(os.Stdout)).With().Timestamp().Caller().Logger()
}

// GetWriter returns the new io.Writer for the logger, configured with pretty-pringing for debug.
func GetWriter(debug bool) io.Writer {
	if debug {
		return zlog.ConsoleWriter{Out: os.Stdout}
	}
	return os.Stdout
}
