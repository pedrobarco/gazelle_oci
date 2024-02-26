package oci

import (
	"flag"
	"log"
	"slices"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/pedrobarco/gazelle_oci/internal/module"
)

// Directives
const (
	// Base image to be used when creating OCI image.
	// #gazelle:oci_base_image lang base_image
	baseImageDirective string = "oci_base_image"
)

var langs = []string{"go", "*"}

type ociConfig struct {
	baseImageSet         map[string]label.Label
	moduleToApparentName func(string) string
}

func newOciConfig() *ociConfig {
	return &ociConfig{
		baseImageSet:         make(map[string]label.Label),
		moduleToApparentName: func(s string) string { return s },
	}
}

// getOciConfig gets the ociConfig from the Gazelle config.
func getOciConfig(c *config.Config) *ociConfig {
	return c.Exts[languageName].(*ociConfig)
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
	if f == nil {
		return
	}

	var cfg *ociConfig
	if _, ok := c.Exts[languageName]; !ok {
		cfg = newOciConfig()
	} else {
		cfg = getOciConfig(c)
	}
	c.Exts[e.Name()] = cfg

	if rel == "" {
		moduleToApparentName, err := module.ExtractModuleToApparentNameMapping(c.RepoRoot)
		if err != nil {
			log.Fatalf("could not extract module apparent names: %v", err)
		}
		cfg.moduleToApparentName = moduleToApparentName
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
