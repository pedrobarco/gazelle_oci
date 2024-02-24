package oci

import (
	"flag"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/repo"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/pedrobarco/gazelle_oci/internal/module"
	myrule "github.com/pedrobarco/gazelle_oci/internal/rule"
)

const (
	languageName string = "oci"
)

// Directives
const (
	// Base image to be used when creating OCI image.
	// #gazelle:oci_base_image lang base_image
	baseImageDirective string = "oci_base_image"
)

type ociConfig struct {
	baseImageSet         map[string]label.Label
	moduleToApparentName func(string) string
}

var langs = []string{"go", "*"}

// SupportsLang returns whether the provided lang is recognized by the version of
// gazelle_oci being used. This avoids incompatibility between new versions of
// Gazelle and old version of gazelle_oci.
func (cfg *ociConfig) SupportsLang(lang string) bool {
	return slices.Contains(langs, lang)
}

// GetBaseImage returns the label to be used as the base image for oci_image.
func (cfg *ociConfig) GetBaseImage(lang string) label.Label {
	if v, ok := cfg.baseImageSet[lang]; ok {
		return v
	}

	if v, ok := cfg.baseImageSet["*"]; ok {
		return v
	}

	return label.NoLabel
}

// GetRepoName returns the apparent name for the original repo.
func (cfg *ociConfig) GetRepoName(repo string) string {
	name := cfg.moduleToApparentName(repo)
	if name == "" {
		switch repo {
		case "rules_go":
			// The legacy name used in WORKSPACE
			return "io_bazel_rules_go"
		default:
			return repo
		}
	}
	return name
}

var kinds = map[string]rule.KindInfo{
	"pkg_tar": {
		NonEmptyAttrs:  map[string]bool{"srcs": true},
		MergeableAttrs: map[string]bool{"srcs": true},
	},
	"oci_image": {
		NonEmptyAttrs:  map[string]bool{"tars": true, "base": true, "entrypoint": true},
		MergeableAttrs: map[string]bool{"tars": true},
	},
	"platform_transition_filegroup": {
		NonEmptyAttrs:  map[string]bool{"srcs": true},
		MergeableAttrs: map[string]bool{"srcs": true, "target_platfrom": true},
	},
	"oci_tarball": {
		NonEmptyAttrs:  map[string]bool{"image": true},
		MergeableAttrs: map[string]bool{"repo_tags": true},
	},
}

type ociImageLang struct{}

var _ language.Language = (*ociImageLang)(nil)
var _ language.ModuleAwareLanguage = (*ociImageLang)(nil)

func NewLanguage() language.Language {
	return &ociImageLang{}
}

// GetConfig gets the ociConfig config from the Gazelle config.
func (e *ociImageLang) GetConfig(c *config.Config) *ociConfig {
	return c.Exts[e.Name()].(*ociConfig)
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
func (e *ociImageLang) KnownDirectives() []string {
	return []string{baseImageDirective}
}

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
func (e *ociImageLang) Configure(c *config.Config, rel string, f *rule.File) {
	var cfg *ociConfig
	if _, ok := c.Exts[e.Name()]; !ok {
		cfg = &ociConfig{
			baseImageSet: make(map[string]label.Label),
		}
	} else {
		cfg = e.GetConfig(c)
	}
	c.Exts[e.Name()] = cfg

	if rel == "" {
		moduleToApparentName, err := module.ExtractModuleToApparentNameMapping(c.RepoRoot)
		if err != nil {
			log.Fatalf("could not extract module apparent names: %v", err)
		}
		cfg.moduleToApparentName = moduleToApparentName
	}

	if f == nil {
		return
	}

	for _, directive := range f.Directives {
		if directive.Key == baseImageDirective {
			split := strings.Split(directive.Value, " ")
			if len(split) != 2 {
				log.Fatalf("bad %s, should be gazelle:%s <lang> <base_image>",
					baseImageDirective, baseImageDirective,
				)
			}

			lang := split[0]
			if !cfg.SupportsLang(lang) {
				log.Fatalf("bad %s, language %s is not supported",
					baseImageDirective, lang,
				)
			}

			val, err := label.Parse(split[1])
			if err != nil {
				log.Fatalf("bad %s, invalid label %s: %v",
					val, baseImageDirective, err,
				)
			}

			cfg.baseImageSet[lang] = val
		}
	}
}

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

	cfg := e.GetConfig(args.Config)
	result := language.GenerateResult{}

	for _, r := range args.OtherGen {
		switch r.Kind() {
		case "go_binary":
			layer := rule.NewRule("pkg_tar", r.Name()+"_layer")
			layer.SetAttr("srcs", []string{":" + r.Name()})
			result.Gen = append(result.Gen, layer)
			result.Imports = append(result.Imports, struct{}{})

			base := cfg.GetBaseImage("go")
			if label.NoLabel.Equal(base) {
				log.Fatalf("failed to get base image for go")
			}

			image := rule.NewRule("oci_image", "image")
			image.SetAttr("base", base.String())
			image.SetAttr("entrypoint", []string{"/" + r.Name()})
			image.SetAttr("tars", []string{":" + layer.Name()})
			result.Gen = append(result.Gen, image)
			result.Imports = append(result.Imports, struct{}{})

			rulesGoRepoName := cfg.GetRepoName("rules_go")

			transition := rule.NewRule("platform_transition_filegroup", "transitioned_image")
			transition.SetAttr("srcs", []string{":" + image.Name()})
			transition.SetAttr("target_platform", myrule.SelectToolchain{
				"@platforms//cpu:arm64":  "@" + rulesGoRepoName + "//go/toolchain:linux_arm64",
				"@platforms//cpu:x86_64": "@" + rulesGoRepoName + "//go/toolchain:linux_amd64",
			})
			result.Gen = append(result.Gen, transition)
			result.Imports = append(result.Imports, struct{}{})

			tarball := rule.NewRule("oci_tarball", "tarball")
			tarball.SetAttr("image", ":"+transition.Name())
			// TODO: support directive to configure registry path and repo tags
			tarball.SetAttr("repo_tags", []string{r.Name() + ":latest"})
			result.Gen = append(result.Gen, tarball)
			result.Imports = append(result.Imports, struct{}{})
		default:
			continue
		}
	}

	return result
}

// Loads returns .bzl files and symbols they define. Every rule generated by
// GenerateRules, now or in the past, should be loadable from one of these
// files.
//
// Deprecated: Implement ModuleAwareLanguage's ApparentLoads.
func (e *ociImageLang) Loads() []rule.LoadInfo {
	panic("ApparentLoads should be called instead")
}

// Fix repairs deprecated usage of language-specific rules in f. This is
// called before the file is indexed. Unless c.ShouldFix is true, fixes
// that delete or rename rules should not be performed.
func (e *ociImageLang) Fix(c *config.Config, f *rule.File) {}

// ApparentLoads returns .bzl files and symbols they define. Every rule
// generated by GenerateRules, now or in the past, should be loadable from
// one of these files.
//
// The moduleToApparentName argument is a function that resolves a given
// Bazel module name to the apparent repository name configured for this
// module in the MODULE.bazel file, or the empty string if there is no such
// module or the MODULE.bazel file doesn't exist. Languages should use the
// non-empty value returned by this function to form the repository part of
// the load statements they return and fall back to using the legacy
// WORKSPACE name otherwise.
//
// See https://bazel.build/external/overview#concepts for more information
// on repository names.
//
// Example: For a project with these lines in its MODULE.bazel file:
//
//	bazel_dep(name = "rules_go", version = "0.38.1", repo_name = "my_rules_go")
//	bazel_dep(name = "gazelle", version = "0.27.0")
//
// moduleToApparentName["rules_go"] == "my_rules_go"
// moduleToApparentName["gazelle"] == "gazelle"
// moduleToApparentName["foobar"] == ""
func (*ociImageLang) ApparentLoads(moduleToApparentName func(string) string) []rule.LoadInfo {
	rulesOci := moduleToApparentName("rules_oci")
	if rulesOci == "" {
		rulesOci = "rules_oci"
	}

	aspectBazelLib := moduleToApparentName("aspect_bazel_lib")
	if aspectBazelLib == "" {
		aspectBazelLib = "aspect_bazel_lib"
	}

	rulesPkg := moduleToApparentName("rules_pkg")
	if rulesPkg == "" {
		rulesPkg = "rules_pkg"
	}

	return []rule.LoadInfo{
		{
			Name:    fmt.Sprintf("@%s//lib:transitions.bzl", aspectBazelLib),
			Symbols: []string{"platform_transition_filegroup"},
		},
		{
			Name:    fmt.Sprintf("@%s//oci:defs.bzl", rulesOci),
			Symbols: []string{"oci_image", "oci_tarball"},
		},
		{
			Name:    fmt.Sprintf("@%s//pkg:tar.bzl", rulesPkg),
			Symbols: []string{"pkg_tar"},
		},
	}
}
