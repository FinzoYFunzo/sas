package lua

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FinzoYFunzo/sas/pkg/text"
)

func withColorDisabled(t *testing.T) {
	t.Helper()

	original := text.UseColor
	text.UseColor = false
	t.Cleanup(func() { text.UseColor = original })
}

func newLogTestEngine(t *testing.T, debug bool) (*Engine, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	withColorDisabled(t)

	var stdout, stderr bytes.Buffer
	logger := text.NewLogger(&stdout, &stderr, strings.NewReader(""), debug, "lua")
	engine := NewWithLogger(logger)
	t.Cleanup(engine.Close)

	return engine, &stdout, &stderr
}

func TestLogInfoWarnAndErrorUseLoggerStreams(t *testing.T) {
	engine, stdout, stderr := newLogTestEngine(t, false)

	require.NoError(t, engine.L.DoString(`
		sas.log.info("info message")
		sas.log.warn("warn message")
		sas.log.error("error message")
	`))

	require.Equal(t, "==> info message\n -> warn message\n", stdout.String())
	require.Equal(t, " -> error message\n", stderr.String())
}

func TestLogDebugUsesLoggerDebugGate(t *testing.T) {
	engine, stdout, stderr := newLogTestEngine(t, false)

	require.NoError(t, engine.L.DoString(`sas.log.debug("hidden")`))
	require.Empty(t, stdout.String())
	require.Empty(t, stderr.String())

	debugLogger := text.NewLogger(stdout, stderr, strings.NewReader(""), true, "lua")
	engine.SetLogger(debugLogger)

	require.NoError(t, engine.L.DoString(`sas.log.debug("visible")`))
	require.Equal(t, "[DEBUG:lua] visible\n", stdout.String())
	require.Empty(t, stderr.String())
}

func TestLogStringifiesLuaArgumentsInOrder(t *testing.T) {
	engine, stdout, stderr := newLogTestEngine(t, false)

	require.NoError(t, engine.L.DoString(`
		local value = setmetatable({}, {
			__tostring = function()
				return "custom"
			end,
		})
		sas.log.info("pkg", 12, true, nil, value)
	`))

	require.Equal(t, "==> pkg 12 true nil custom\n", stdout.String())
	require.Empty(t, stderr.String())
}

func TestLogWithoutLoggerDoesNotPanic(t *testing.T) {
	engine := New()
	t.Cleanup(engine.Close)

	require.NoError(t, engine.L.DoString(`
		sas.log.debug("debug")
		sas.log.info("info")
		sas.log.warn("warn")
		sas.log.error("error")
	`))
}

func TestLoadUsesLuaChildLogger(t *testing.T) {
	withColorDisabled(t)

	path := writeLuaFile(t, `sas.log.debug("loaded")`)
	var stdout, stderr bytes.Buffer
	logger := text.NewLogger(&stdout, &stderr, strings.NewReader(""), true, "fallback")

	engine, err := Load(logger, path, &testConfig{})
	require.NoError(t, err)
	t.Cleanup(engine.Close)

	require.Equal(t, "[DEBUG:lua] loaded\n", stdout.String())
	require.Empty(t, stderr.String())
}
