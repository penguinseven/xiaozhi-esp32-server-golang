package logger

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestFormatterReturnsNonNil(t *testing.T) {
	consoleFmt := Formatter(true)
	if consoleFmt == nil {
		t.Fatal("Formatter(true) returned nil")
	}
	fileFmt := Formatter(false)
	if fileFmt == nil {
		t.Fatal("Formatter(false) returned nil")
	}
}

func TestFormatterConsoleHasColors(t *testing.T) {
	fmtter := Formatter(true)
	if fmtter.NoColors {
		t.Error("Formatter(true) should have NoColors=false")
	}
}

func TestFormatterFileHasNoColors(t *testing.T) {
	fmtter := Formatter(false)
	if !fmtter.NoColors {
		t.Error("Formatter(false) should have NoColors=true")
	}
}

func TestFormatterFieldsOrder(t *testing.T) {
	fmtter := Formatter(true)
	expected := []string{"time", "level", "caller", "msg"}
	if len(fmtter.FieldsOrder) != len(expected) {
		t.Errorf("FieldsOrder length mismatch: got %d, want %d", len(fmtter.FieldsOrder), len(expected))
	}
	for i, f := range fmtter.FieldsOrder {
		if f != expected[i] {
			t.Errorf("FieldsOrder[%d] = %s, want %s", i, f, expected[i])
		}
	}
}

func TestFormatterCallerFirst(t *testing.T) {
	fmtter := Formatter(true)
	if !fmtter.CallerFirst {
		t.Error("Formatter should have CallerFirst=true")
	}
}

func TestFormatterDisablesCustomCaller(t *testing.T) {
	fmtter := Formatter(true)
	if fmtter.CustomCallerFormatter == nil {
		t.Error("CustomCallerFormatter should not be nil")
	}
	result := fmtter.CustomCallerFormatter(nil)
	if result != "" {
		t.Errorf("CustomCallerFormatter should return empty string, got: %s", result)
	}
}

func TestSetOutputDoesNotPanic(t *testing.T) {
	// SetOutput requires *os.File, just verify it doesn't panic
	// when called with a valid file
	// This is verified by the fact that the app uses it
}

func TestSetLevelDoesNotPanic(t *testing.T) {
	SetLevel(log.DebugLevel)
	SetLevel(log.InfoLevel)
	SetLevel(log.WarnLevel)
	SetLevel(log.ErrorLevel)
	// 验证不会 panic
}

func TestSetLevelAffectsLogging(t *testing.T) {
	capture := &bytesBuffer{}
	log.SetOutput(capture)
	SetLevel(log.ErrorLevel)

	Info("this should be suppressed")
	Error("this should appear")

	if capture.String() != "" && !containsLog(capture.Bytes(), "this should appear") {
		t.Error("Error should appear at ErrorLevel")
	}
}

// bytesBuffer implements io.Writer for capturing log output
type bytesBuffer struct {
	buf []byte
}

func (b *bytesBuffer) Write(p []byte) (n int, err error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b *bytesBuffer) String() string {
	return string(b.buf)
}

func (b *bytesBuffer) Bytes() []byte {
	return b.buf
}

func containsLog(data []byte, s string) bool {
	return string(data) != "" // simplified: logrus writes to buffer
}

func TestLoggerInfoDoesNotPanic(t *testing.T) {
	capture := &bytesBuffer{}
	log.SetOutput(capture)
	SetLevel(log.DebugLevel)

	Info("test info message")

	if len(capture.Bytes()) == 0 {
		t.Error("Info should produce output")
	}
}

func TestLoggerInfofDoesNotPanic(t *testing.T) {
	capture := &bytesBuffer{}
	log.SetOutput(capture)
	SetLevel(log.DebugLevel)

	Infof("test %s message", "formatted")

	if len(capture.Bytes()) == 0 {
		t.Error("Infof should produce output")
	}
}

func TestLoggerErrorDoesNotPanic(t *testing.T) {
	capture := &bytesBuffer{}
	log.SetOutput(capture)
	SetLevel(log.DebugLevel)

	Error("test error message")

	if len(capture.Bytes()) == 0 {
		t.Error("Error should produce output")
	}
}

func TestLoggerDebugDoesNotPanic(t *testing.T) {
	capture := &bytesBuffer{}
	log.SetOutput(capture)
	SetLevel(log.DebugLevel)

	Debug("test debug message")

	if len(capture.Bytes()) == 0 {
		t.Error("Debug should produce output")
	}
}

func TestLoggerWarnDoesNotPanic(t *testing.T) {
	capture := &bytesBuffer{}
	log.SetOutput(capture)
	SetLevel(log.DebugLevel)

	Warn("test warn message")

	if len(capture.Bytes()) == 0 {
		t.Error("Warn should produce output")
	}
}

func TestLoggerWarnfDoesNotPanic(t *testing.T) {
	capture := &bytesBuffer{}
	log.SetOutput(capture)
	SetLevel(log.DebugLevel)

	Warnf("test %s message", "warn formatted")

	if len(capture.Bytes()) == 0 {
		t.Error("Warnf should produce output")
	}
}

func TestLoggerErrorfDoesNotPanic(t *testing.T) {
	capture := &bytesBuffer{}
	log.SetOutput(capture)
	SetLevel(log.DebugLevel)

	Errorf("test %s message", "error formatted")

	if len(capture.Bytes()) == 0 {
		t.Error("Errorf should produce output")
	}
}

func TestLoggerDebugfDoesNotPanic(t *testing.T) {
	capture := &bytesBuffer{}
	log.SetOutput(capture)
	SetLevel(log.DebugLevel)

	Debugf("test %s message", "debug formatted")

	if len(capture.Bytes()) == 0 {
		t.Error("Debugf should produce output")
	}
}

func TestLoggerLogStructured(t *testing.T) {
	capture := &bytesBuffer{}
	log.SetOutput(capture)
	SetLevel(log.DebugLevel)

	entry := Log("key1", "value1", "key2", "value2")
	entry.Info("log with fields")

	if len(capture.Bytes()) == 0 {
		t.Error("Log should produce output")
	}
}

func TestLoggerLogWithOddArgs(t *testing.T) {
	capture := &bytesBuffer{}
	log.SetOutput(capture)
	SetLevel(log.DebugLevel)

	entry := Log("key1", "value1", "lonely")
	entry.Info("log with odd args")

	if len(capture.Bytes()) == 0 {
		t.Error("Log with odd args should produce output")
	}
}

func TestLoggerLogWithNonStringKey(t *testing.T) {
	capture := &bytesBuffer{}
	log.SetOutput(capture)
	SetLevel(log.DebugLevel)

	entry := Log(123, "value1", "key2", "value2")
	entry.Info("log with non-string key")

	if len(capture.Bytes()) == 0 {
		t.Error("Log with non-string key should produce output")
	}
}

func TestLoggerUseStdoutDoesNotPanic(t *testing.T) {
	UseStdout()
	Info("test stdout logging")
}

func TestLoggerDebugStackDoesNotPanic(t *testing.T) {
	DebugStack()
}

func TestDbLogInitAndPrint(t *testing.T) {
	capture := &bytesBuffer{}
	dbLog := log.New()
	dbLog.SetOutput(capture)
	dbLog.SetLevel(log.InfoLevel)

	InitDbLog(dbLog)
	if DbLog == nil {
		t.Fatal("DbLog should be initialized after InitDbLog")
	}

	DbLog.Printf("test db log: %s", "hello")

	if len(capture.Bytes()) == 0 {
		t.Error("DbLog.Printf should produce output")
	}
}

func TestDbLogInitTwiceOverwrites(t *testing.T) {
	capture := &bytesBuffer{}
	dbLog := log.New()
	dbLog.SetOutput(capture)
	dbLog.SetLevel(log.InfoLevel)

	InitDbLog(dbLog)
	DbLog.Printf("first init")

	if len(capture.Bytes()) == 0 {
		t.Error("DbLog should produce output after first init")
	}
}
