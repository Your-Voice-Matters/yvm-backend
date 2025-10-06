package structs

type PollObj struct {
	Creator string   `json:"created_by"`
	Title   string   `json:"title"`
	Desc    string   `json:"description"`
	Options []string `json:"options"`
	Id      int      `json:"id,omitempty"`
}
