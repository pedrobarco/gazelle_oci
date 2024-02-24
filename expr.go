package oci

import (
	"sort"

	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/bazelbuild/buildtools/build"
)

type SelectToolchain map[string]string

var _ rule.BzlExprValue = (*SelectToolchain)(nil)

func (s SelectToolchain) BzlExpr() build.Expr {
	defaultKey := "//conditions:default"
	keys := make([]string, 0, len(s))
	haveDefaultKey := false
	for key := range s {
		if key == defaultKey {
			haveDefaultKey = true
		} else {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	if haveDefaultKey {
		keys = append(keys, defaultKey)
	}

	args := make([]*build.KeyValueExpr, 0, len(s))
	for _, key := range keys {
		value := rule.ExprFromValue(s[key])
		args = append(args, &build.KeyValueExpr{
			Key:   &build.StringExpr{Value: key},
			Value: value,
		})
	}
	sel := &build.CallExpr{
		X:    &build.Ident{Name: "select"},
		List: []build.Expr{&build.DictExpr{List: args, ForceMultiLine: true}},
	}
	return sel
}
