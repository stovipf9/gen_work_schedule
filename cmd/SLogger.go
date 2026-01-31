package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/afero"
)

type SLogger struct {
	slog.Logger
	logFile afero.File
}

func NewSLogger(fs afero.Fs, logFileName string) *SLogger {
	stdoutLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logFile, err := fs.OpenFile(
		logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stdoutLogger.Warn(err.Error())
		return &SLogger{*stdoutLogger, nil}
	}

	return &SLogger{*slog.New(slog.NewJSONHandler(logFile, nil)), logFile}
}

// logger.Logger interface methods
func (s *SLogger) Print(message string) {
	s.Logger.Debug(message)
}

func (s *SLogger) Trace(message string) {
	s.Logger.Debug(message)
}

func (s *SLogger) Debug(message string) {
	s.Logger.Debug(message)
}

func (s *SLogger) Info(message string) {
	s.Logger.Info(message)
}

func (s *SLogger) Warning(message string) {
	s.Logger.Warn(message)
}

func (s *SLogger) Error(message string) {
	s.Logger.Error(message)
}

func (s *SLogger) Fatal(message string) {
	s.Logger.Error(message)
}

func (s *SLogger) Close() {
	if s.logFile != nil {
		s.logFile.Close()
	}
}

type LogLevel int

const (
	TRACE LogLevel = iota + 1
	DEBUG
	INFO
	WARNING
	ERROR
	FATAL
)

func (s *SLogger) Log(level LogLevel, err error) error {
	if err == nil {
		return nil
	}
	map[LogLevel]func(string){
		TRACE:   s.Trace,
		DEBUG:   s.Debug,
		INFO:    s.Info,
		WARNING: s.Warning,
		ERROR:   s.Error,
		FATAL:   s.Fatal,
	}[level](err.Error())
	return err
}
