package config

// RunConfig holds all parameters for a single generation run.
type RunConfig struct {
	Date     string
	Author   string
	Enrich   bool
	LangCode string
}
