package models

type GenerationStat struct {
	CPE []string `json:"cpe,omitempty" mapstructure:"cpe" validate:"required,dive,cpe_generation"`
}
