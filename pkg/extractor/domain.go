package extractor

// PreRegistration ["full_name",
// "idade"
// ,"date_of_birth",
// "sab_op1","sab_op2",
// "dom_op1","dom_op2",
// "student_mobile_number",
// "matemática","português",
// "lógica","redação",
// "220_informática","215_digitação",
// "210_auxiliar_administrativo","235_elétrica",
// "205_iniciação_profissional","225_montagem_de_micro","260_ajustador_soldador",
// "270_inglês","245_climatização","program_sab",
// "program_sab2","program_dom","name"]
type PreRegistration struct {
	FullName                   string  `json:"full_name"`
	Idade                      float64 `json:"idade"`
	Date_of_birth              string  `json:"date_of_birth"`
	Opcao_1_sabado             string  `json:"sab_op1"`
	Opcao_2_sabado             string  `json:"sab_op2"`
	Opcao_1_domingo            string  `json:"dom_op1"`
	Opcao_2_domingo            string  `json:"dom_op2"`
	Celular                    string  `json:"student_mobile_number"`
	Matematica                 string  `json:"matemática"`
	Portugues                  string  `json:"português"`
	Logica                     string  `json:"lógica"`
	Redacao                    string  `json:"redação"`
	Iniciacao_profissional     string  `json:"205_iniciação_profissional"`
	Auxiliar_administrtativo   string  `json:"210_auxiliar_administrativo"`
	Digitação                  string  `json:"215_digitação"`
	Informática_basica_domingo string  `json:"220_informática"`
	Informatica_basica_sabado  string  `json:"220_informática_sab"`
	Montagem_de_micro          string  `json:"225_montagem_de_micro"`
	Eletrica                   string  `json:"235_elétrica"`
	Climatizador               string  `json:"245_climatização"`
	Ajustador_soldador         string  `json:"260_ajustador_soldador"`
	Ingles                     string  `json:"270_inglês"`
	SegundaChamada             int32   `json:"segunda_chamada"`
	CursoAprovadoSabado1       string  `json:"program_sab,omitempty"`
	CursoAprovadoSabado2       string  `json:"program_sab2,omitempty"`
	CursoAprovadoDomingo1      string  `json:"program_dom,omitempty"`
	Name                       string  `json:"name"`
}
