package config

type JwtConfig struct {
	Secret string `yaml:"secret"`
	Issuer string `yaml:"issuer"`
}
