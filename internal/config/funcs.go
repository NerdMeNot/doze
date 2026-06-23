package config

import (
	"os"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

// stdlibFunctions is the function set available in doze.hcl expressions. It is a
// curated slice of the go-cty standard library (the same family Terraform builds
// on) plus a small doze-specific `env` helper for reading host environment
// variables.
func stdlibFunctions() map[string]function.Function {
	return map[string]function.Function{
		// strings
		"upper":      stdlib.UpperFunc,
		"lower":      stdlib.LowerFunc,
		"title":      stdlib.TitleFunc,
		"trim":       stdlib.TrimFunc,
		"trimspace":  stdlib.TrimSpaceFunc,
		"trimprefix": stdlib.TrimPrefixFunc,
		"trimsuffix": stdlib.TrimSuffixFunc,
		"replace":    stdlib.ReplaceFunc,
		"split":      stdlib.SplitFunc,
		"join":       stdlib.JoinFunc,
		"format":     stdlib.FormatFunc,
		"formatlist": stdlib.FormatListFunc,
		"regex":      stdlib.RegexFunc,
		"substr":     stdlib.SubstrFunc,
		"strlen":     stdlib.StrlenFunc,
		"chomp":      stdlib.ChompFunc,
		"indent":     stdlib.IndentFunc,

		// collections
		"concat":   stdlib.ConcatFunc,
		"length":   stdlib.LengthFunc,
		"contains": stdlib.ContainsFunc,
		"keys":     stdlib.KeysFunc,
		"values":   stdlib.ValuesFunc,
		"lookup":   stdlib.LookupFunc,
		"merge":    stdlib.MergeFunc,
		"flatten":  stdlib.FlattenFunc,
		"distinct": stdlib.DistinctFunc,
		"sort":     stdlib.SortFunc,
		"reverse":  stdlib.ReverseListFunc,
		"slice":    stdlib.SliceFunc,
		"element":  stdlib.ElementFunc,
		"coalesce": stdlib.CoalesceFunc,
		"range":    stdlib.RangeFunc,

		// numbers
		"abs":      stdlib.AbsoluteFunc,
		"ceil":     stdlib.CeilFunc,
		"floor":    stdlib.FloorFunc,
		"max":      stdlib.MaxFunc,
		"min":      stdlib.MinFunc,
		"parseint": stdlib.ParseIntFunc,

		// encoding
		"jsonencode": stdlib.JSONEncodeFunc,
		"jsondecode": stdlib.JSONDecodeFunc,
		"csvdecode":  stdlib.CSVDecodeFunc,

		// doze-specific
		"env": envFunc(),
	}
}

// envFunc reads a host environment variable, returning an optional default (or
// "") when it is unset. Lets config parameterize from the shell: env("PGDATA").
func envFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{Name: "name", Type: cty.String},
		},
		VarParam: &function.Parameter{Name: "default", Type: cty.String},
		Type:     function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, _ cty.Type) (cty.Value, error) {
			if v, ok := os.LookupEnv(args[0].AsString()); ok {
				return cty.StringVal(v), nil
			}
			if len(args) > 1 {
				return args[1], nil
			}
			return cty.StringVal(""), nil
		},
	})
}
