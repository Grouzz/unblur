package protocol

type Config struct {
	Action string  `json:"action"` //"blur" or "deblur"
	Sigma  float64 `json:"sigma"`
	K      float64 `json:"k"`
}
