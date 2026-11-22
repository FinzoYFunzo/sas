//go:build !integration

package query

import (
	"io"
	"strings"
	"testing"

	aurc "github.com/Jguer/aur"

	dbmock "github.com/FinzoYFunzo/sas/pkg/db/mock"
	settingslua "github.com/FinzoYFunzo/sas/pkg/settings/lua"
	"github.com/FinzoYFunzo/sas/pkg/settings/parser"
	"github.com/FinzoYFunzo/sas/pkg/text"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSourceQueryBuilderRenderAURHook verifies that a RenderAUR hook replaces
// the default AUR row rendering while sync rows keep their default formatting.
func TestSourceQueryBuilderRenderAURHook(t *testing.T) {
	t.Parallel()

	mockDB, mockAUR := newSasQueryBuilderMocks()

	w := &strings.Builder{}
	logger := text.NewLogger(w, io.Discard, strings.NewReader(""), false, "test")

	e := settingslua.New()
	defer e.Close()

	require.NoError(t, e.L.DoString(`
		sas.create_autocmd("RenderAUR", {
			callback = function(event)
				return "AURLINE:" .. event.data.name
			end,
		})
	`))

	qb := NewSourceQueryBuilder(mockAUR, logger, "", parser.ModeAny, "", false, false, false)
	qb.SetLua(e)
	qb.Execute(t.Context(), mockDB, []string{"sas"})

	err := qb.Results(mockDB, Detailed)
	require.NoError(t, err)

	out := w.String()

	// Both AUR packages should use the hook's output.
	assert.Contains(t, out, "AURLINE:sas", "hook output for sas must appear")
	assert.Contains(t, out, "AURLINE:sas-git", "hook output for sas-git must appear")

	// Default AUR rendering includes the votes marker; hook replaces it entirely.
	assert.NotContains(t, out, "(+2461", "default AUR rendering must not appear when hook is active")

	// Sync package should keep its default rendering (no RenderSync hook registered).
	assert.Contains(t, out, "rubs-sard", "sync package must still appear with default rendering")
}

// TestSourceQueryBuilderRenderSyncHook verifies that a RenderSync hook replaces
// the default sync row rendering.
func TestSourceQueryBuilderRenderSyncHook(t *testing.T) {
	t.Parallel()

	mockDB, mockAUR := newSasQueryBuilderMocks()

	w := &strings.Builder{}
	logger := text.NewLogger(w, io.Discard, strings.NewReader(""), false, "test")

	e := settingslua.New()
	defer e.Close()

	require.NoError(t, e.L.DoString(`
		sas.create_autocmd("RenderSync", {
			callback = function(event)
				return "SYNCLINE:" .. event.data.name .. "/" .. event.data.repository
			end,
		})
	`))

	qb := NewSourceQueryBuilder(mockAUR, logger, "", parser.ModeAny, "", false, false, false)
	qb.SetLua(e)
	qb.Execute(t.Context(), mockDB, []string{"sas"})

	err := qb.Results(mockDB, Detailed)
	require.NoError(t, err)

	out := w.String()
	assert.Contains(t, out, "SYNCLINE:rubs-sard/extra", "hook output for rubs-sard must appear")
}

// TestSourceQueryBuilderRenderAURNilFallsBack verifies that a RenderAUR hook
// that returns nil causes the default formatter to be used for that row.
func TestSourceQueryBuilderRenderAURNilFallsBack(t *testing.T) {
	t.Parallel()

	mockDB, mockAUR := newSasQueryBuilderMocks()

	w := &strings.Builder{}
	logger := text.NewLogger(w, io.Discard, strings.NewReader(""), false, "test")

	e := settingslua.New()
	defer e.Close()

	require.NoError(t, e.L.DoString(`
		sas.create_autocmd("RenderAUR", {
			callback = function(event)
				return nil
			end,
		})
	`))

	qb := NewSourceQueryBuilder(mockAUR, logger, "", parser.ModeAny, "", false, false, false)
	qb.SetLua(e)
	qb.Execute(t.Context(), mockDB, []string{"sas"})

	err := qb.Results(mockDB, Detailed)
	require.NoError(t, err)

	out := w.String()
	// Default AUR rendering must appear when hook returns nil.
	assert.Contains(t, out, "(+2461", "default AUR rendering must appear when hook returns nil")
}

func TestSourceQueryBuilderRenderAURHookReceivesLocalVersionWhenSame(t *testing.T) {
	t.Parallel()

	e := settingslua.New()
	defer e.Close()

	require.NoError(t, e.L.DoString(`
		sas.create_autocmd("RenderAUR", {
			callback = function(event)
				return "LOCAL:" .. event.data.local_version
			end,
		})
	`))

	mockDB := &dbmock.DBExecutor{
		LocalPackageFn: func(string) dbmock.IPackage {
			return &dbmock.Package{PName: "sas", PVersion: "12.5.7-1"}
		},
	}
	logger := text.NewLogger(io.Discard, io.Discard, strings.NewReader(""), false, "test")

	qb := NewSourceQueryBuilder(nil, logger, "", parser.ModeAny, "", false, false, false)
	qb.SetLua(e)

	rendered := qb.renderAUR(&aurc.Pkg{Name: "sas", Version: "12.5.7-1"}, mockDB)
	require.Equal(t, "LOCAL:12.5.7-1", rendered)
}

func TestSourceQueryBuilderRenderSyncHookReceivesLocalVersionWhenSame(t *testing.T) {
	t.Parallel()

	e := settingslua.New()
	defer e.Close()

	require.NoError(t, e.L.DoString(`
		sas.create_autocmd("RenderSync", {
			callback = function(event)
				return "LOCAL:" .. event.data.local_version
			end,
		})
	`))

	mockDB := &dbmock.DBExecutor{
		LocalPackageFn: func(string) dbmock.IPackage {
			return &dbmock.Package{PName: "rubs-sard", PVersion: "0.9.34-5"}
		},
	}
	pkg := &dbmock.Package{PDB: dbmock.NewDB("extra"), PName: "rubs-sard", PVersion: "0.9.34-5"}
	logger := text.NewLogger(io.Discard, io.Discard, strings.NewReader(""), false, "test")

	qb := NewSourceQueryBuilder(nil, logger, "", parser.ModeAny, "", false, false, false)
	qb.SetLua(e)

	rendered := qb.renderSync(pkg, mockDB)
	require.Equal(t, "LOCAL:0.9.34-5", rendered)
}

func TestSourceQueryBuilderDefaultRenderUsesShortInstalledTagWhenSame(t *testing.T) {
	t.Parallel()

	mockDB := &dbmock.DBExecutor{
		LocalPackageFn: func(string) dbmock.IPackage {
			return &dbmock.Package{PName: "sas", PVersion: "12.5.7-1"}
		},
	}
	logger := text.NewLogger(io.Discard, io.Discard, strings.NewReader(""), false, "test")

	qb := NewSourceQueryBuilder(nil, logger, "", parser.ModeAny, "", false, false, false)

	rendered := qb.renderAUR(&aurc.Pkg{Name: "sas", Version: "12.5.7-1"}, mockDB)
	assert.Contains(t, rendered, "(Installed)")
	assert.NotContains(t, rendered, "(Installed: 12.5.7-1)")
}
