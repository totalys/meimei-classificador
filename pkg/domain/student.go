package domain

import "fmt"

type Student struct {
	Name       string     `json:"name"`
	Choices    []string   `json:"choices"`
	Grade      float64    `json:"grade"`
	Approved   []Approved `json:"approved"`
	Waitlisted bool       `json:"waitlisted"`
	Age        string     `json:"idade"`
	Phone      string     `json:"celular"`
	SegChamada int32      `json:"segunda_chamada"`
	DocName    string     `json:"doc_name"`
}

type Approved struct {
	IsSenai bool
	Course  string
	Days    []int
}

func (s Student) GetChoicesNames(choiceMap map[string]CourseConfig) (choiceNames []string) {

	for _, c := range s.Choices {
		choiceNames = append(choiceNames, fmt.Sprintf("%s[%s]", choiceMap[c].Name, c))
	}
	return choiceNames
}
