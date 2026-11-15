package topo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGraph_DependenciesAndDependents_Direct(t *testing.T) {
	g := New[string, struct{}]()

	// sas depends on go
	require.NoError(t, g.DependOn("sas", "go"))

	// Dependencies("sas") => {"go"}
	depsSas := g.Dependencies("sas")
	require.NotNil(t, depsSas)
	require.Len(t, depsSas, 1)
	require.True(t, depsSas["go"])

	// Dependencies("go") => {} (empty set, because "go" is in the graph but has no deps)
	depsGo := g.Dependencies("go")
	require.NotNil(t, depsGo)
	require.Len(t, depsGo, 0)

	// Dependents("go") => {"sas"}
	dependentsGo := g.Dependents("go")
	require.NotNil(t, dependentsGo)
	require.Len(t, dependentsGo, 1)
	require.True(t, dependentsGo["sas"])

	// Dependents("sas") => {} (empty set, because nothing depends on sas)
	dependentsSas := g.Dependents("sas")
	require.NotNil(t, dependentsSas)
	require.Len(t, dependentsSas, 0)
}

func TestGraph_DependenciesAndDependents_Transitive(t *testing.T) {
	g := New[string, struct{}]()

	// sas depends on go; foo depends on sas
	require.NoError(t, g.DependOn("sas", "go"))
	require.NoError(t, g.DependOn("foo", "sas"))

	// Dependencies("foo") => {"sas", "go"}
	depsFoo := g.Dependencies("foo")
	require.NotNil(t, depsFoo)
	require.Len(t, depsFoo, 2)
	require.True(t, depsFoo["sas"])
	require.True(t, depsFoo["go"])

	// Dependents("go") => {"sas", "foo"} (transitive)
	dependentsGo := g.Dependents("go")
	require.NotNil(t, dependentsGo)
	require.Len(t, dependentsGo, 2)
	require.True(t, dependentsGo["sas"])
	require.True(t, dependentsGo["foo"])

	// Dependents("sas") => {"foo"}
	dependentsSas := g.Dependents("sas")
	require.NotNil(t, dependentsSas)
	require.Len(t, dependentsSas, 1)
	require.True(t, dependentsSas["foo"])
}

func TestGraph_DependenciesAndDependents_MissingNodeReturnsNil(t *testing.T) {
	g := New[string, struct{}]()

	// For nodes not present in the graph, transitive queries return nil.
	require.Nil(t, g.Dependencies("missing"))
	require.Nil(t, g.Dependents("missing"))

	// Adding edges adds nodes; existing nodes with no deps/dependents return an empty set (non-nil).
	require.NoError(t, g.DependOn("sas", "go"))
	require.NotNil(t, g.Dependencies("go"))
	require.Len(t, g.Dependencies("go"), 0)
	require.NotNil(t, g.Dependents("sas"))
	require.Len(t, g.Dependents("sas"), 0)
}
