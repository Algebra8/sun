package sun

import (
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// Module builtins is a Starlark module of Python-like builtins functions.
var Module = &starlarkstruct.Module{
	Name: "builtins",
	Members: starlark.StringDict{
		"map":      starlark.NewBuiltin("map", map_),
		"next":     starlark.NewBuiltin("next", next),
		"filter":   starlark.NewBuiltin("filter", filter),
		"callable": starlark.NewBuiltin("callable", callable),
		"hex":      starlark.NewBuiltin("hex", hex),
		"oct":      starlark.NewBuiltin("oct", oct),
		"bin":      starlark.NewBuiltin("bin", bin),
	},
}

// Module itertools is a Starlark module of Python's itertools module.
var ItertoolsModule = &starlarkstruct.Module{
	Name: "itertools",
	Members: starlark.StringDict{
		"count": starlark.NewBuiltin("itertools.count", count_),
	},
}
