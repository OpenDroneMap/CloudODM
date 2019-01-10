package odm

// Option is an option that can be passed to NodeODM
type Option struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}
