package config

// Config holds CLI configuration that can be populated via env / unmarshal.
// Fields must be exported (capitalized) so reflection-based unmarshalers can set them.
type Config struct {
	Debug     bool   `mapstructure:"debug" json:"debug"`
	Directory string `mapstructure:"directory" json:"directory"`
}

// CliConfig is the package-level configuration instance used by the CLI.
var CliConfig Config
