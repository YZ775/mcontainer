package status

type Status struct {
	OciVersion string `json:"ociVersion"`
	ID         string `json:"id"`
	Status     string `json:"status"`
	Pid        int    `json:"pid"`
	Bundle     string `json:"bundle"`
}
