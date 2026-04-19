package configscan

type Placeholder struct {
	Name         string  `json:"name"`
	DefaultValue *string `json:"defaultValue,omitempty"`
	Required     bool    `json:"required"`
	SourceFile   string  `json:"sourceFile"`
	SourcePath   string  `json:"sourcePath"`
}

type ScanResult struct {
	Required []Placeholder `json:"required"`
	Optional []Placeholder `json:"optional"`
}

func Merge(results ...*ScanResult) *ScanResult {
	merged := &ScanResult{}
	for _, result := range results {
		if result == nil {
			continue
		}
		merged.Required = append(merged.Required, result.Required...)
		merged.Optional = append(merged.Optional, result.Optional...)
	}
	return merged
}
