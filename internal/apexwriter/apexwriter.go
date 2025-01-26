package apexwriter

import (
	"strings"

	"github.com/apex/log"
)

// Writer is a struct that implements the io.Writer interface and can,
// therefore, be passed into log.SetOutput.  Writer must always be
// constructed by called NewWriter.
type Writer struct {
	fields log.Fielder
}

// Write implements the io.Writer interface for Writer.  Log messages
// output via this method will have their date and time information
// stripped (apex log will have its own) .
func (w *Writer) Write(p []byte) (n int, err error) {
	// trim everything we do not want to see in a log
	msg := strings.TrimRight(string(p), " \n\t")

	// split the message and return every non-empty line as a separate log entry...
	for _, s := range strings.FieldsFunc(msg, func(c rune) bool { return c == '\n' || c == '\r' }) {
		if strings.TrimSpace(s) != "" {
			// Sometimes, log-running Pulumi actions just log a "." - log those dots at debug level only
			// Besides that, Terraform actions should be shown at debug level es well
			if s == "." || strings.Contains(s, "] Terraform") {
				log.WithFields(w.fields).Debug(s)
			} else {
				log.WithFields(w.fields).Info(s)
			}
		}
	}

	return len(p), nil
}

// NewWriter creates a new Writer that can be passed to the SetOutput
// function in the log package from the standard library.
func NewWriter(fields log.Fielder) *Writer {
	writer := &Writer{
		fields: fields,
	}

	return writer
}
