package log

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type CustomTextFormatter struct{}

func (f *CustomTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	message := fmt.Sprintf("%s %5.5s %s\n",
		entry.Time.Format("2006-01-02 15:04:05.000"), // Date-time
		strings.ToUpper(entry.Level.String()),        // Log level
		entry.Message,                                // Log message
	)

	return []byte(message), nil
}

var Log = &logrus.Logger{
	Out:       os.Stdout,
	Level:     logrus.InfoLevel,
	Formatter: &CustomTextFormatter{},
}
