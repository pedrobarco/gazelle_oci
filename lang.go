package oci

import (
	"flag"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/repo"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

const (
	languageName = "oci"
)

var kinds = map[string]rule.KindInfo{
	"pkg_tar": {
		NonEmptyAttrs:  map[string]bool{"srcs": true},
		MergeableAttrs: map[string]bool{"srcs": true},
	},
	"oci_image": {
		NonEmptyAttrs:  map[string]bool{"tars": true, "base": true, "entrypoint": true},
		MergeableAttrs: map[string]bool{"tars": true},
	},
}

type ociImageLang struct{}

var _ language.Language = (*ociImageLang)(nil)

func NewLanguage() language.Language {
	return &ociImageLang{}
}

// RegisterFlags registers command-line flags used by the extension. This
// method is called once with the root configuration when Gazelle
// starts. RegisterFlags may set an initial values in Config.Exts. When flags
// are set, they should modify these values.
func (e *ociImageLang) RegisterFlags(fs *flag.FlagSet, cmd string, c *config.Config) {}

// CheckFlags validates the configuration after command line flags are parsed.
// This is called once with the root configuration when Gazelle starts.
// CheckFlags may set default values in flags or make implied changes.
func (e *ociImageLang) CheckFlags(fs *flag.FlagSet, c *config.Config) error { return nil }

// KnownDirectives returns a list of directive keys that this Configurer can
// interpret. Gazelle prints errors for directives that are not recoginized by
// any Configurer.
func (e *ociImageLang) KnownDirectives() []string { return nil }

// Configure modifies the configuration using directives and other information
// extracted from a build file. Configure is called in each directory.
//
// c is the configuration for the current directory. It starts out as a copy
// of the configuration for the parent directory.
//
// rel is the slash-separated relative path from the repository root to
// the current directory. It is "" for the root directory itself.
//
// f is the build file for the current directory or nil if there is no
// existing build file.
func (e *ociImageLang) Configure(c *config.Config, rel string, f *rule.File) {}

// Name returns the name of the language. This should be a prefix of the
// kinds of rules generated by the language, e.g., "go" for the Go extension
// since it generates "go_library" rules.
func (e *ociImageLang) Name() string {
	return languageName
}

// Imports returns a list of ImportSpecs that can be used to import the rule
// r. This is used to populate RuleIndex.
//
// If nil is returned, the rule will not be indexed. If any non-nil slice is
// returned, including an empty slice, the rule will be indexed.
func (e *ociImageLang) Imports(c *config.Config, r *rule.Rule, f *rule.File) []resolve.ImportSpec {
	return nil
}

// Embeds returns a list of labels of rules that the given rule embeds. If
// a rule is embedded by another importable rule of the same language, only
// the embedding rule will be indexed. The embedding rule will inherit
// the imports of the embedded rule.
func (e *ociImageLang) Embeds(r *rule.Rule, from label.Label) []label.Label {
	return nil
}

// Resolve translates imported libraries for a given rule into Bazel
// dependencies. Information about imported libraries is returned for each
// rule generated by language.GenerateRules in
// language.GenerateResult.Imports. Resolve generates a "deps" attribute (or
// the appropriate language-specific equivalent) for each import according to
// language-specific rules and heuristics.
func (e *ociImageLang) Resolve(c *config.Config, ix *resolve.RuleIndex, rc *repo.RemoteCache, r *rule.Rule, imports interface{}, from label.Label) {
}

// Kinds returns a map of maps rule names (kinds) and information on how to
// match and merge attributes that may be found in rules of those kinds. All
// kinds of rules generated for this language may be found here.
func (e *ociImageLang) Kinds() map[string]rule.KindInfo {
	return kinds
}

// GenerateRules extracts build metadata from source files in a directory.
// GenerateRules is called in each directory where an update is requested
// in depth-first post-order.
//
// args contains the arguments for GenerateRules. This is passed as a
// struct to avoid breaking implementations in the future when new
// fields are added.
//
// A GenerateResult struct is returned. Optional fields may be added to this
// type in the future.
//
// Any non-fatal errors this function encounters should be logged using
// log.Print.
func (e *ociImageLang) GenerateRules(args language.GenerateArgs) language.GenerateResult {
	if args.File == nil {
		return language.GenerateResult{}
	}

	var rules []*rule.Rule
	var imports []interface{}

	for _, r := range args.File.Rules {
		switch r.Kind() {
		case "go_binary":
			layer := rule.NewRule("pkg_tar", r.Name()+"_layer")
			layer.SetAttr("srcs", []string{":" + r.Name()})

			image := rule.NewRule("oci_image", "image")
			image.SetAttr("base", "@distroless_base")
			image.SetAttr("entrypoint", []string{"/" + r.Name()})
			image.SetAttr("tars", []string{":" + r.Name()})

			rules = append(rules, layer, image)
			imports = append(imports, nil, nil)
		default:
			continue
		}
	}

	return language.GenerateResult{
		Gen:     rules,
		Imports: imports,
	}
}

// Loads returns .bzl files and symbols they define. Every rule generated by
// GenerateRules, now or in the past, should be loadable from one of these
// files.
//
// Deprecated: Implement ModuleAwareLanguage's ApparentLoads.
func (e *ociImageLang) Loads() []rule.LoadInfo {
	return []rule.LoadInfo{
		{
			Name:    "@rules_pkg//pkg:tar.bzl",
			Symbols: []string{"pkg_tar"},
		},
		{
			Name:    "@rules_oci//oci:defs.bzl",
			Symbols: []string{"oci_image"},
		},
	}
}

// Fix repairs deprecated usage of language-specific rules in f. This is
// called before the file is indexed. Unless c.ShouldFix is true, fixes
// that delete or rename rules should not be performed.
func (e *ociImageLang) Fix(c *config.Config, f *rule.File) {}
