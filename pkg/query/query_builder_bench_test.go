//go:build !integration

package query

import (
	"io"
	"strings"
	"testing"

	settingslua "github.com/FinzoYFunzo/sas/pkg/settings/lua"
	"github.com/FinzoYFunzo/sas/pkg/settings/parser"
	"github.com/FinzoYFunzo/sas/pkg/text"
)

// BenchmarkApplySearchFilter_noLua measures applySearchFilter when no Lua
// engine is attached (the common fast-path — a nil check and immediate return).
func BenchmarkApplySearchFilter_noLua(b *testing.B) {
	logger := text.NewLogger(io.Discard, io.Discard, strings.NewReader(""), false, "bench")
	mockDB, mockAUR := newSasQueryBuilderMocks()

	qb := NewSourceQueryBuilder(mockAUR, logger, "", parser.ModeAny, "", false, false, false)

	// Pre-populate results so the benchmark only measures the filter path.
	qb.Execute(b.Context(), mockDB, []string{"sas"})

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		qb.applySearchFilter(qb.results)
	}
}

// BenchmarkApplySearchFilter_withLua measures applySearchFilter when a Lua
// SearchFilter hook is registered but performs a passthrough (returns all
// results). This isolates the overhead of the Lua dispatch + table marshaling.
func BenchmarkApplySearchFilter_withLua(b *testing.B) {
	logger := text.NewLogger(io.Discard, io.Discard, strings.NewReader(""), false, "bench")
	mockDB, mockAUR := newSasQueryBuilderMocks()

	e := settingslua.New()
	defer e.Close()

	if err := e.L.DoString(`
		sas.create_autocmd("SearchFilter", {
			callback = function(event)
				local out = {}
				for _, r in ipairs(event.data.results) do
					out[#out + 1] = { source = r.source, name = r.name }
				end
				return out
			end,
		})
	`); err != nil {
		b.Fatalf("lua setup: %v", err)
	}

	qb := NewSourceQueryBuilder(mockAUR, logger, "", parser.ModeAny, "", false, false, false)
	qb.SetLua(e)
	qb.Execute(b.Context(), mockDB, []string{"sas"})

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		qb.applySearchFilter(qb.results)
	}
}

// BenchmarkExecute_noLua measures the full Execute call (AUR query + sort)
// without a Lua hook.
func BenchmarkExecute_noLua(b *testing.B) {
	logger := text.NewLogger(io.Discard, io.Discard, strings.NewReader(""), false, "bench")
	mockDB, mockAUR := newSasQueryBuilderMocks()

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		qb := NewSourceQueryBuilder(mockAUR, logger, "", parser.ModeAny, "", false, false, false)
		qb.Execute(b.Context(), mockDB, []string{"sas"})
	}
}

// BenchmarkExecute_withLuaFilter measures the full Execute call when a Lua
// SearchFilter hook is in play.
func BenchmarkExecute_withLuaFilter(b *testing.B) {
	logger := text.NewLogger(io.Discard, io.Discard, strings.NewReader(""), false, "bench")
	mockDB, mockAUR := newSasQueryBuilderMocks()

	e := settingslua.New()
	defer e.Close()

	if err := e.L.DoString(`
		sas.create_autocmd("SearchFilter", {
			callback = function(event)
				local out = {}
				for _, r in ipairs(event.data.results) do
					if r.source == "aur" then
						out[#out + 1] = { source = r.source, name = r.name }
					end
				end
				return out
			end,
		})
	`); err != nil {
		b.Fatalf("lua setup: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		qb := NewSourceQueryBuilder(mockAUR, logger, "", parser.ModeAny, "", false, false, false)
		qb.SetLua(e)
		qb.Execute(b.Context(), mockDB, []string{"sas"})
	}
}
