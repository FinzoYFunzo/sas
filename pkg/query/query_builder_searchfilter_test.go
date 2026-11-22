//go:build !integration

package query

import (
	"io"
	"strings"
	"testing"

	settingslua "github.com/FinzoYFunzo/sas/pkg/settings/lua"
	"github.com/FinzoYFunzo/sas/pkg/settings/parser"
	"github.com/FinzoYFunzo/sas/pkg/text"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSourceQueryBuilderSearchFilterHook verifies that a SearchFilter hook
// registered on the Lua engine is applied after ranking, filtering s.results
// to only the entries the callback returns.
func TestSourceQueryBuilderSearchFilterHook(t *testing.T) {
	t.Parallel()

	mockDB, mockAUR := newSasQueryBuilderMocks()

	w := &strings.Builder{}
	logger := text.NewLogger(w, io.Discard, strings.NewReader(""), false, "test")

	// Baseline: build without a Lua hook and record count.
	baselineQB := NewSourceQueryBuilder(mockAUR, logger, "", parser.ModeAny, "", false, false, false)
	baselineQB.Execute(t.Context(), mockDB, []string{"sas"})
	baselineCount := len(baselineQB.results)
	require.Greater(t, baselineCount, 0, "baseline query must return at least one result")

	// Count how many AUR results exist in the baseline.
	aurCount := 0
	for _, r := range baselineQB.results {
		if r.source == "aur" {
			aurCount++
		}
	}
	require.Greater(t, aurCount, 0, "baseline must have at least one AUR result")

	// The mock also returns a non-AUR (sync) result for "sas".
	// Confirm there is at least one so the filter actually does something.
	nonAURCount := baselineCount - aurCount
	require.Greater(t, nonAURCount, 0, "baseline must have at least one non-AUR result for the filter to exercise")

	// Build with a SearchFilter that keeps only AUR results.
	e := settingslua.New()
	defer e.Close()

	require.NoError(t, e.L.DoString(`
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
	`))

	filteredQB := NewSourceQueryBuilder(mockAUR, logger, "", parser.ModeAny, "", false, false, false)
	filteredQB.SetLua(e)
	filteredQB.Execute(t.Context(), mockDB, []string{"sas"})

	assert.Equal(t, aurCount, len(filteredQB.results),
		"filtered results should equal the number of AUR packages in the baseline")

	for _, r := range filteredQB.results {
		assert.Equal(t, "aur", r.source,
			"every result after the SearchFilter hook must be from aur")
	}
}

func TestSourceQueryBuilderSearchFilterHookSkipsUnknownResults(t *testing.T) {
	t.Parallel()

	e := settingslua.New()
	defer e.Close()

	require.NoError(t, e.L.DoString(`
		sas.create_autocmd("SearchFilter", {
			callback = function(_)
				return {
					{ source = "aur", name = "known" },
					{ source = "aur", name = "unknown" },
				}
			end,
		})
	`))

	builder := &SourceQueryBuilder{
		lua:    e,
		logger: text.NewLogger(io.Discard, io.Discard, strings.NewReader(""), true, "test"),
	}
	filtered := builder.applySearchFilter([]abstractResult{
		{source: "aur", name: "known", description: "known package"},
	})

	require.Len(t, filtered, 1)
	assert.Equal(t, "known", filtered[0].name)
}
