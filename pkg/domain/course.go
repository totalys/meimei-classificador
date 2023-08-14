package domain

type CourseConfig struct {
	Name       string `json:"name"`
	Approved   int    `json:"approved"`
	Waitlist   int    `json:"wait"`
	Days       []int  `json:"days"`
	Sala       string `json:"sala"`
	DataInicio string `json:"dataInicio"`
	IsSenai    bool   `json:"is_senai"`
}

type CoursesConfigMap struct {
	Data          string `json:"data"`
	CoursesConfig map[string]CourseConfig
}
