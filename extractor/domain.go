package extractor

type Applicant struct {
	First_name                 string  `json:"first_name"`
	Last_name                  string  `json:"last_name"`
	Idade                      float64 `json:"idade"`
	Date_of_birth              string  `json:"date_of_birth"`
	Opcao_1_sabado             string  `json:"opção_1_sábado"`
	Opcao_2_sabado             string  `json:"opção_2_sábado"`
	Opcao_1_domingo            string  `json:"opção_1_domingo"`
	Opcao_2_domingo            string  `json:"opção_2_domingo"`
	Celular                    string  `json:"student_mobile_number"`
	Matematica                 string  `json:"matemática"`
	Portugues                  string  `json:"português"`
	Logica                     string  `json:"lógica"`
	Redacao                    string  `json:"redação"`
	Informática_basica_domingo string  `json:"informática_básica_domingo"`
	Digitação                  string  `json:"digitação"`
	Auxiliar_administrtativo   string  `json:"210_auxiliar_administrtativo"`
	Eletrica                   string  `json:"elétrica"`
	Iniciacao_profissional     string  `json:"iniciação_profissional"`
	Montagem_de_micro          string  `json:"montagem_de_micro"`
	Ajustador_soldador         string  `json:"ajustador_soldador"`
	Ingles                     string  `json:"inglês"`
	Informatica_basica_sabado  string  `json:"informática_básica_sábado"`
	Climatizador               string  `json:"climatização"`
}
