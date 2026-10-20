// Package config holds the command-line / environment configuration shared
// by the batch runner (cmd/tfcalc) and the TAUDIS loader (cmd/taudisload).
package config

import (
	"flag"
	"os"
)

// EnvDB is the environment variable giving the SQLite TAUDIS database
// location. The -db flag, when given, overrides it.
const EnvDB = "TFCALC_DB"

// DefaultDB is the database location used when neither the flag nor the
// environment variable is set.
const DefaultDB = "taudis.db"

// Config is the shared configuration.
type Config struct {
	// DBPath is the SQLite database file holding the taudis table.
	DBPath string
}

// Register declares the shared flags on fs. Defaults come from the
// environment, so the precedence is flag > environment > built-in default.
// Read the returned Config after fs.Parse.
func Register(fs *flag.FlagSet) *Config {
	cfg := &Config{}
	def := os.Getenv(EnvDB)
	if def == "" {
		def = DefaultDB
	}
	fs.StringVar(&cfg.DBPath, "db", def, "SQLite TAUDIS database file (env "+EnvDB+")")
	return cfg
}
