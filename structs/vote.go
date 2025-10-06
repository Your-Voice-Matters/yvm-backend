package structs

type Vote struct {
	Votername   string `json:"votername"`
	PollID      int    `json:"pollid"`
	OptionID    int    `json:"chosenoption"`
	Description string `json:"description"`
}
