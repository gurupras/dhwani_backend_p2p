package types

type AudioDevice struct {
	Name       string `json:"name"`
	Identifier string `json:"id"`
	CanPlay    bool   `json:"canPlay"`
	CanRecord  bool   `json:"canRecord"`
	Default    bool   `json:"default"`
}
