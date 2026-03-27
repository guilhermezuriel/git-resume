package config


type RunConfig struct {
	DateFrom    string
	DateTo      string
	Author      string
	Enrich      bool
	LangCode    string
	AllBranches bool
	Consolidate bool

	Model string
}
