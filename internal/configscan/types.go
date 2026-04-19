package configscan

// Placeholder represents a single ${VAR} or ${VAR:default} reference found in a config file.
type Placeholder struct {
	Name         string
	DefaultValue *string // nil means no default (required); non-nil means optional with that default
	Required     bool
	SourceFile   string
	SourcePath   string // YAML key path (e.g. spring.datasource.url) or .properties key
}

// ScanResult holds placeholders extracted from one or more config files.
type ScanResult struct {
	Required []Placeholder
	Optional []Placeholder
}
