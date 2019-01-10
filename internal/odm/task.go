package odm

type Task struct {
	UUID            string   `json:"uuid"`
	Options         []Option `json:"options"`
	OutputDirectory string   `json:"outputDirectory"`
}
