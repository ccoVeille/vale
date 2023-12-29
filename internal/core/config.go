package core

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/errata-ai/ini"
	"github.com/errata-ai/vale/v2/internal/glob"
)

var (
	// ConfigDir is the default location for Vale's configuration files.
	//
	// This was introduced in v3.0.0 as a means of standardizing the location
	// of Vale's configuration files.
	//
	// This directory is relative to the user's specified `StylesPath`, which
	// can be set via the `--config` flag, the `VALE_CONFIG_PATH` environment
	// variable, or the default search process.
	//
	// NOTE: The config pipeline is stored in the top-level `.vale-config`
	// directory. See `cmd/vale/sync.go`.
	ConfigDir = "config"

	// Vocabularies are loaded in `ini.go`.
	VocabDir = filepath.Join(ConfigDir, "vocabularies")

	// Dictionaries are loaded in `spelling.go#makeSpeller`.
	DictDir = filepath.Join(ConfigDir, "dictionaries")

	// Templates are loaded in `cmd/vale/custom.go`.
	TmplDir = filepath.Join(ConfigDir, "templates")

	// Ignore files are loaded in `spelling.go#NewSpelling`.
	IgnoreDir = filepath.Join(ConfigDir, "ignore")
)

// IgnoreFiles returns a list of all user-defined ignore files.
func IgnoreFiles(stylesPath string) ([]string, error) {
	ignore := filepath.Join(stylesPath, IgnoreDir)
	return doublestar.FilepathGlob(filepath.Join(ignore, "**", "*.txt"))
}

// CLIFlags holds the values that are defined at runtime by the user.
//
// For example, `vale --minAlertLevel=error`.
type CLIFlags struct {
	AlertLevel string
	Built      string
	Glob       string
	InExt      string
	Output     string
	Path       string
	Sources    string
	Filter     string
	Local      bool
	NoExit     bool
	Normalize  bool
	Relative   bool
	Remote     bool
	Simple     bool
	Sorted     bool
	Wrap       bool
	Version    bool
	Help       bool
}

// Config holds the configuration values from both the CLI and `.vale.ini`.
type Config struct {
	// General configuration
	BlockIgnores   map[string][]string        // A list of blocks to ignore
	Checks         []string                   // All checks to load
	Formats        map[string]string          // A map of unknown -> known formats
	Asciidoctor    map[string]string          // A map of asciidoctor attributes
	FormatToLang   map[string]string          // A map of format to lang ID
	GBaseStyles    []string                   // Global base style
	GChecks        map[string]bool            // Global checks
	IgnoredClasses []string                   // A list of HTML classes to ignore
	IgnoredScopes  []string                   // A list of HTML tags to ignore
	MinAlertLevel  int                        // Lowest alert level to display
	RuleToLevel    map[string]string          // Single-rule level changes
	SBaseStyles    map[string][]string        // Syntax-specific base styles
	SChecks        map[string]map[string]bool // Syntax-specific checks
	SkippedScopes  []string                   // A list of HTML blocks to ignore
	Stylesheets    map[string]string          // XSLT stylesheet
	StylesPath     string                     // Directory with Rule.yml files
	TokenIgnores   map[string][]string        // A list of tokens to ignore
	WordTemplate   string                     // The template used in YAML -> regexp list conversions
	RootINI        string                     // the path to the project's .vale.ini file

	DictionaryPath string // Location to search for dictionaries.

	Vocabularies []Vocabulary         `json:"-"`
	FallbackPath string               `json:"-"`
	SecToPat     map[string]glob.Glob `json:"-"`
	Styles       []string             `json:"-"`
	Paths        []string             `json:"-"`
	Root         string               `json:"-"`

	NLPEndpoint string // An external API to call for NLP-related work.

	// Command-line configuration
	Flags *CLIFlags `json:"-"`

	StyleKeys []string `json:"-"`
	RuleKeys  []string `json:"-"`
}

// NewConfig initializes a Config with its default values.
func NewConfig(flags *CLIFlags) (*Config, error) {
	var cfg Config

	cfg.BlockIgnores = make(map[string][]string)
	cfg.Flags = flags
	cfg.Formats = make(map[string]string)
	cfg.Asciidoctor = make(map[string]string)
	cfg.GChecks = make(map[string]bool)
	cfg.MinAlertLevel = 1
	cfg.RuleToLevel = make(map[string]string)
	cfg.SBaseStyles = make(map[string][]string)
	cfg.SChecks = make(map[string]map[string]bool)
	cfg.SecToPat = make(map[string]glob.Glob)
	cfg.Stylesheets = make(map[string]string)
	cfg.TokenIgnores = make(map[string][]string)
	cfg.Paths = []string{""}
	cfg.FormatToLang = make(map[string]string)

	return &cfg, nil
}

func (c *Config) String() string {
	c.StylesPath = filepath.ToSlash(c.StylesPath)
	b, _ := json.MarshalIndent(c, "", "  ")
	return string(b)
}

// Get the user-defined packages from a `.vale.ini` file.
func GetPackages(src string) ([]string, error) {
	packages := []string{}

	uCfg, err := ini.Load(src)
	if err != nil {
		return packages, err
	}

	core := uCfg.Section("")
	return core.Key("Packages").Strings(","), nil
}

func pipeConfig(cfg *Config) ([]string, error) {
	var sources []string

	pipeline := filepath.Join(cfg.StylesPath, ".vale-config")
	if IsDir(pipeline) && len(cfg.Flags.Sources) == 0 {
		configs, err := os.ReadDir(pipeline)
		if err != nil {
			return sources, err
		}

		for _, config := range configs {
			if config.IsDir() {
				continue
			}
			sources = append(sources, filepath.Join(pipeline, config.Name()))
		}
		sources = append(sources, cfg.Flags.Path)
	}

	return sources, nil
}
