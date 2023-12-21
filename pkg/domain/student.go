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

	// somente considera cursos cadastrados na configuração
	for _, c := range s.Choices {
		course, ok := choiceMap[c]
		if !ok {
			continue
		}
		choiceNames = append(choiceNames, fmt.Sprintf("%s[%s]", course.Name, c))
	}

	return choiceNames
}
