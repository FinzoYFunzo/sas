//go:build !integration

package settings

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var luaMetaFieldRe = regexp.MustCompile(`^---@field\s+([A-Za-z_][A-Za-z0-9_]*)`)

// luaMetaOptFields returns the field names declared on the sas.opt class in
// meta/sas.d.lua.
func luaMetaOptFields(meta string) map[string]bool {
	fields := map[string]bool{}

	inOpt := false

	for line := range strings.SplitSeq(meta, "\n") {
		switch {
		case strings.HasPrefix(line, "---@class "):
			inOpt = strings.TrimSpace(strings.TrimPrefix(line, "---@class ")) == "sas.opt"
		case inOpt:
			if m := luaMetaFieldRe.FindStringSubmatch(line); m != nil {
				fields[m[1]] = true
			}
		}
	}

	return fields
}

// TestLuaMetaCoversAllOptions guards meta/sas.d.lua against drifting from the
// lua-tagged fields of Configuration: lua-language-server users would otherwise
// lose completion and type checking for newly added options.
func TestLuaMetaCoversAllOptions(t *testing.T) {
	meta, err := os.ReadFile(filepath.Join("..", "..", "meta", "sas.d.lua"))
	require.NoError(t, err)

	metaFields := luaMetaOptFields(string(meta))
	require.NotEmpty(t, metaFields)

	configFields := map[string]bool{}

	for field := range reflect.TypeFor[Configuration]().Fields() {
		if name := field.Tag.Get("lua"); name != "" && name != "-" {
			configFields[name] = true
		}
	}

	for name := range configFields {
		assert.Contains(t, metaFields, name, "meta/sas.d.lua is missing sas.opt field %q", name)
	}

	for name := range metaFields {
		assert.True(t, configFields[name], "meta/sas.d.lua declares unknown sas.opt field %q", name)
	}
}

func TestLuaMetaOptFields(t *testing.T) {
	meta := "---@class sas.opt\n---@field a string\n---@field b string\n\n---@class sas.log\n---@field c string\n"
	assert.Equal(t, map[string]bool{"a": true, "b": true}, luaMetaOptFields(meta))
}
