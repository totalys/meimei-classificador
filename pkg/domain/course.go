package domain

type CourseConfig struct {
	Name       string `json:"name"`
	Approved   int    `json:"approved"`
	Waitlist   int    `json:"wait"`
	Days       []int  `json:"days"`
	Sala       string `json:"sala"`
	DataInicio string `json:"dataInicio"`
	// IsSenai usado para identificar cursos do Senai. Em uma versão anterior,
	// evitava que o aluno fizesse 2 cursos ao mesmo tempo.
	// Posteriormente, identificou-se que isso não era impeditivo.
	// Contudo, mantivemos a propriedade para eventuais relatórios.
	IsSenai bool `json:"is_senai"`
}

type CoursesConfigMap struct {
	Data          string `json:"data"`
	CoursesConfig map[string]CourseConfig
}
