package models

type Config struct {
	Network           string `toml:"network"`
	IsExternalNetwork bool   `toml:"is_external_network"`
}
