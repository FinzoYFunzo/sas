package lua

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	BuildDir      string `json:"buildDir" lua:"build_dir"`
	RequestSplitN int    `json:"requestsplitn" lua:"request_split_n"`
	Devel         bool   `json:"devel" lua:"devel"`
	AnswerClean   string `json:"answerclean" lua:"answer_clean"`
	AnswerDiff    string `json:"answerdiff" lua:"answer_diff"`
	AnswerEdit    string `json:"answeredit" lua:"answer_edit"`
	Ignored       string `json:"-" lua:"-"`
}

func TestApply(t *testing.T) {
	t.Parallel()
	e := New()
	t.Cleanup(e.Close)

	require.NoError(t, e.L.DoString(`
		sas.opt.build_dir = "/tmp/sas"
		sas.opt.request_split_n = 200
		sas.opt.devel = true
	`))

	cfg := &testConfig{}
	unknown, errs := e.Apply(cfg)

	assert.Empty(t, unknown)
	assert.Empty(t, errs)
	assert.Equal(t, "/tmp/sas", cfg.BuildDir)
	assert.Equal(t, 200, cfg.RequestSplitN)
	assert.True(t, cfg.Devel)
}

func TestApplyUnknownAndTypeMismatch(t *testing.T) {
	t.Parallel()
	e := New()
	t.Cleanup(e.Close)

	require.NoError(t, e.L.DoString(`
		sas.opt.does_not_exist = "x"
		sas.opt.buildDir = "/tmp/nope"
		sas.opt.requestsplitn = 200
		sas.opt.Devel = true
		sas.opt.devel = "not a bool"
		sas.opt.build_dir = "/tmp/ok"
	`))

	cfg := &testConfig{}
	unknown, errs := e.Apply(cfg)

	assert.ElementsMatch(t, []string{"does_not_exist", "buildDir", "requestsplitn", "Devel"}, unknown)
	assert.Len(t, errs, 1)
	assert.Zero(t, cfg.RequestSplitN)
	assert.False(t, cfg.Devel)
	assert.Equal(t, "/tmp/ok", cfg.BuildDir)
}

func TestApplyAppliesAnswerOptionsFromLua(t *testing.T) {
	t.Parallel()
	e := New()
	t.Cleanup(e.Close)

	require.NoError(t, e.L.DoString(`
		sas.opt.answer_clean = "All"
		sas.opt.answer_diff = "None"
		sas.opt.answer_edit = "Installed"
	`))

	cfg := &testConfig{}
	unknown, errs := e.Apply(cfg)

	assert.Empty(t, unknown)
	assert.Empty(t, errs)
	assert.Equal(t, "All", cfg.AnswerClean)
	assert.Equal(t, "None", cfg.AnswerDiff)
	assert.Equal(t, "Installed", cfg.AnswerEdit)
}

type namedSliceTestConfig struct {
	BuildDir string           `lua:"build_dir"`
	Repos    []namedSliceRepo `lua:"repos"`
}

type namedSliceRepo struct {
	Name  string `lua:"name"`
	URL   string `lua:"url"`
	Depth int    `lua:"depth"`
}

func TestApplyNamedStructSlice(t *testing.T) {
	t.Parallel()
	e := New()
	t.Cleanup(e.Close)

	require.NoError(t, e.L.DoString(`
		sas.opt.build_dir = "/tmp/sas"
		sas.opt.repos = {
			["sas-pkgbuild"] = {
				url = "https://github.com/Jguer/yay-PKGBUILD",
				depth = 2,
			},
			["local-repo"] = {
				url = "file:///srv/pkgbuilds",
			},
		}
	`))

	cfg := &namedSliceTestConfig{}
	unknown, errs := e.Apply(cfg)

	assert.Empty(t, unknown)
	assert.Empty(t, errs)
	assert.Equal(t, "/tmp/sas", cfg.BuildDir)

	// The keyed table becomes a slice sorted by repo name for determinism.
	require.Len(t, cfg.Repos, 2)

	assert.Equal(t, "local-repo", cfg.Repos[0].Name)
	assert.Equal(t, "file:///srv/pkgbuilds", cfg.Repos[0].URL)
	assert.Equal(t, 0, cfg.Repos[0].Depth)

	assert.Equal(t, "sas-pkgbuild", cfg.Repos[1].Name)
	assert.Equal(t, "https://github.com/Jguer/yay-PKGBUILD", cfg.Repos[1].URL)
	assert.Equal(t, 2, cfg.Repos[1].Depth)
}

func TestApplyNamedStructSliceRejectsUnknownKey(t *testing.T) {
	t.Parallel()
	e := New()
	t.Cleanup(e.Close)

	require.NoError(t, e.L.DoString(`
		sas.opt.repos = {
			["sas-pkgbuild"] = {
				url = "https://github.com/Jguer/yay-PKGBUILD",
				nonsense = true,
			},
		}
	`))

	cfg := &namedSliceTestConfig{}
	_, errs := e.Apply(cfg)

	assert.Len(t, errs, 1)
}

func TestApplyNamedStructSliceRejectsNameKeyInsideEntry(t *testing.T) {
	t.Parallel()
	e := New()
	t.Cleanup(e.Close)

	require.NoError(t, e.L.DoString(`
		sas.opt.repos = {
			["sas-pkgbuild"] = {
				name = "something-else",
				url = "https://github.com/Jguer/yay-PKGBUILD",
			},
		}
	`))

	cfg := &namedSliceTestConfig{}
	_, errs := e.Apply(cfg)

	require.Len(t, errs, 1)
	assert.ErrorContains(t, errs[0], `unknown key "name"`)
}

type nonStructSliceTestConfig struct {
	Tags []string `lua:"tags"`
}

func TestApplyRejectsSliceOfNonStructs(t *testing.T) {
	t.Parallel()
	e := New()
	t.Cleanup(e.Close)

	require.NoError(t, e.L.DoString(`
		sas.opt.tags = { "a", "b" }
	`))

	cfg := &nonStructSliceTestConfig{}
	_, errs := e.Apply(cfg)

	require.Len(t, errs, 1)
	assert.ErrorContains(t, errs[0], "unsupported slice element kind string")
	assert.Empty(t, cfg.Tags)
}

func TestApplyNamedStructSliceRejectsArrayStyleTable(t *testing.T) {
	t.Parallel()
	e := New()
	t.Cleanup(e.Close)

	require.NoError(t, e.L.DoString(`
		sas.opt.repos = {
			{ url = "a" },
			{ url = "b" },
		}
	`))

	cfg := &namedSliceTestConfig{}
	_, errs := e.Apply(cfg)

	require.Len(t, errs, 1)
	assert.ErrorContains(t, errs[0], "entry keys must be strings")
}

func TestApplyRejectsNonPointer(t *testing.T) {
	t.Parallel()
	e := New()
	t.Cleanup(e.Close)

	_, errs := e.Apply(testConfig{})
	assert.Len(t, errs, 1)
}
