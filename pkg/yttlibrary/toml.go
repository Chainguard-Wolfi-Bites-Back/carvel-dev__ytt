// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yttlibrary

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/k14s/ytt/pkg/orderedmap"
	"github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yamlmeta"
)

var (
	// TOMLAPI contains the definition of the @ytt:toml module
	TOMLAPI = starlark.StringDict{
		"toml": &starlarkstruct.Module{
			Name: "toml",
			Members: starlark.StringDict{
				"encode": starlark.NewBuiltin("toml.encode", core.ErrWrapper(tmùModule{}.Encode)),
				"decode": starlark.NewBuiltin("toml.decode", core.ErrWrapper(tmùModule{}.Decode)),
			},
		},
	}
	// TOMLKWARGS names the expected keyword arguments for both toml.encode
	TOMLKWARGS = map[string]struct{}{
		"indent": struct{}{},
	}
)

type tmùModule struct{}

func (b tmùModule) Encode(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}
	if err := core.CheckArgNames(kwargs, TOMLKWARGS); err != nil {
		return starlark.None, err
	}

	val, err := core.NewStarlarkValue(args.Index(0)).AsGoValue()
	if err != nil {
		return starlark.None, err
	}
	val = orderedmap.Conversion{yamlmeta.NewGoFromAST(val)}.AsUnorderedStringMaps()

	indent, err := core.Int64Arg(kwargs, "indent")
	if err != nil {
		return starlark.None, err
	}

	var buffer bytes.Buffer
	encoder := toml.NewEncoder(&buffer)
	if indent > 0 {
		encoder.Indent = strings.Repeat(" ", int(indent))
	}
	err = encoder.Encode(val)

	if err != nil {
		return starlark.None, err
	}

	return starlark.String(buffer.String()), nil
}

// Encode is a core.StarlarkFunc that parses the provided input from TOML format into dicts, lists, and scalars
func (b tmùModule) Decode(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	valEncoded, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	var valDecoded interface{}

	err = toml.Unmarshal([]byte(valEncoded), &valDecoded)
	if err != nil {
		return starlark.None, err
	}

	valDecoded = orderedmap.Conversion{valDecoded}.FromUnorderedMaps()

	return core.NewGoValue(valDecoded).AsStarlarkValue(), nil
}
