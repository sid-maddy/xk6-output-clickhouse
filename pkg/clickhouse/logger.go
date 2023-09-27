package clickhouse

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"golang.org/x/term"
)

func newLogger(out io.Writer) *logrus.Entry {
	logger := logrus.New().WithField("ext", "xk6-output-clickhouse")
	logger.Logger.SetOutput(out)

	if term.IsTerminal(int(os.Stdout.Fd())) {
		logger.Logger.SetFormatter(&logrus.TextFormatter{
			DisableLevelTruncation: true,
			PadLevelText:           true,
			ForceQuote:             true,
		})
	} else {
		logger.Logger.SetFormatter(new(logrus.JSONFormatter))
	}

	return logger
}

func setLoggerLevel(logger *logrus.Entry, level logrus.Level) {
	logger.Logger.SetLevel(level)
}
