package extractor

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	curso_aux_admin = "215-SG-2024-25 / AUXILIAR ADMIN. (Turma 5)"
	curso_info_dom  = "220-SG-2024.02A / INFORMÁTICA BÁSICA (domingo)"
	curso_info_sab  = "220-SG-2024.02B / INFORMÁTICA BÁSICA (sábado)"
	curso_aux_elet  = "231-SG-2024.02 / AUXILIAR ELETRICISTA"
	curso_inic_prof = "250-SG-2024-25 / INIC. PROF. (Básico)"
	curso_digitacao = "255-SG-2024.02 / DIGITAÇÃO"
	curso_ajus_sold = "260-SG-2024.25 / AJUSTADOR SOLDADOR (Módulo 1)"

	op_sab_1 = "sab - 1a. Opção"
	op_sab_2 = "sab - 2a. Opção"
	op_dom_1 = "dom - 1a. Opção"
	op_dom_2 = "dom - 2a. Opção"
)

var (
	now    = time.Now()
	fields = []string{
		"usuario", "idade", "date_of_birth", "student_group",
		"lm_precadastro_option", "celular", "matematica",
		"portugues", "logica", "redacao", "media_prova", "media_final", "avaliacao", "name"}
)

type Nota struct {
	Nome         string `json:"NOME,omitempty"`
	Celular      string `json:"CELULAR,omitempty"`
	Idade        string `json:"IDADE,omitempty"`
	Sabado1a     string `json:"1ª OPCAO S,omitempty"`
	Sabado2a     string `json:"2ª OPCAO S,omitempty"`
	Domingo1a    string `json:"1ª OPCAO D,omitempty"`
	Domingo2a    string `json:"2ª OPCAO D,omitempty"`
	Matematica   string `json:"MATEMATICA (0 - 10),omitempty"`
	Portugues    string `json:"PORTUGUES (0 - 10),omitempty"`
	Logica       string `json:"LOGICA (0 - 5),omitempty"`
	Redacao      string `json:"REDACAO (0 - 10),omitempty"`
	EntDigitacao string `json:"DIGITACAO,omitempty"`
	EntInicProf  string `json:"INIC PROF,omitempty"`
	EntAuxAdm    string `json:"AUX ADM,omitempty"`
	EntInfoSab   string `json:"INFOR SAB,omitempty"`
	EntInfoDom   string `json:"INFOR DOM,omitempty"`
	EntIngles    string `json:"INGLES,omitempty"`
	EntEletrica  string `json:"ELETRICA,omitempty"`
	EntAjustador string `json:"AJUSTADOR,omitempty"`
	// EntMontMicro    string  `json:"MONT MICRO,omitempty"`
	// EntClimat       string  `json:"CLIMATIZADOR,omitempty"`
	NotaProva       string             `json:"NOTA PROVA,omitempty"`
	MediaFinal      float64            `json:"MEDIA FINAL,omitempty"`
	NotaFinal       string             `json:"NOTA UNICA,omitempty"`
	SegChamada      int32              `json:"segunda_chamada,omitempty"`
	CursoApvSabado  string             `json:"Curso_Aprovado,omitempty"`
	CursoApvDomingo string             `json:"curso_aprovado_domingo,omitempty"`
	Name            string             `json:"name"`
	Entrevistas     map[string]float64 `json:"entrevistas,omitempty"`
}

func GetNotas(baseUrl string) ([]Nota, error) {

	queryParams := url.Values{}
	queryParams.Add("fields", fmt.Sprintf("[\"%s\"]", strings.Join(fields, "\",\"")))
	queryParams.Add("limit_start", "0")
	queryParams.Add("limit_page_length", "1000")

	if os.Getenv("SEGUNDA_CHAMADA") == "1" {
		queryParams.Add("filters", fmt.Sprintf(
			"[[\"%s\",\"%s\",\"%d\"]]", "segunda_chamada", "=", 0))
	}

	queryParams.Add("filters", fmt.Sprintf(
		"[[\"%s\",\"%s\",\"%s\"]]", "status", "=", "Entrevistado"))

	queryParams.Add("order_by", "usuario")

	// Create the URL with query parameters
	requestURL := fmt.Sprintf("%s?%s", baseUrl, queryParams.Encode())

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request %w", err)
	}

	req.Header.Set("Authorization", os.Getenv("LARMEIMEI_USER_API_KEY"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making the request: %+w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %+w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error received response: %s,\n request: %+v, \n %+w", responseBody, req, err)
	}

	var R SysMeimeiResponse
	err = json.Unmarshal(responseBody, &R)
	if err != nil {
		return nil, fmt.Errorf("error parsing response: \n %+v \n to: \n %+v \n err: %+w", string(responseBody), R, err)
	}
	fmt.Println("Total de fichas de entrevistas: ", len(R.Data))
	return toNotas(&R), nil
}

type SysMeimeiResponse struct {
	Data []Interview `json:"data"`
}

func toNotas(R *SysMeimeiResponse) []Nota {

	var notas = []Nota{}
	notasMap := make(map[string]Nota)

	for _, r := range R.Data {
		fullname := strings.Trim(strings.ToUpper(r.FullName), " ")
		nota := Nota{}
		if current, found := notasMap[fullname]; found {
			nota = notasMap[current.Nome]
		} else {
			gradesAverage := getGradeAverage(r.Portugues, r.Matematica, r.Logica, r.Redacao)
			nota = Nota{
				Nome:        fullname,
				Celular:     stripNonNumeric(r.Celular),
				Idade:       getAge(now, int(math.Floor(r.Idade)), r.Date_of_birth),
				Matematica:  r.Matematica,
				Portugues:   r.Portugues,
				Logica:      r.Logica,
				Redacao:     r.Redacao,
				NotaProva:   gradesAverage,
				MediaFinal:  r.MediaFinal,
				Name:        r.Name,
				Entrevistas: make(map[string]float64),
			}
		}

		switch r.Curso {
		case curso_aux_admin:
			nota.EntAuxAdm = r.Entrevista
		case curso_info_dom:
			nota.EntInfoDom = r.Entrevista
		case curso_info_sab:
			nota.EntInfoSab = r.Entrevista
		case curso_aux_elet:
			nota.EntEletrica = r.Entrevista
		case curso_inic_prof:
			nota.EntInicProf = r.Entrevista
		case curso_digitacao:
			nota.EntDigitacao = r.Entrevista
		case curso_ajus_sold:
			nota.EntAjustador = r.Entrevista
		}

		nota.Entrevistas[r.Curso] = parseFloat(r.Entrevista)

		switch r.Opcao {
		case op_sab_1:
			nota.Sabado1a = r.Curso
		case op_sab_2:
			nota.Sabado2a = r.Curso
		case op_dom_1:
			nota.Domingo1a = r.Curso
		case op_dom_2:
			nota.Domingo2a = r.Curso
		}

		notasMap[fullname] = nota
	}

	for _, v := range notasMap {
		v.NotaFinal = getFinalGrade(v.NotaProva, v)
		notas = append(notas, v)
	}

	go printGrades(notas)

	return notas
}

func getFinalGrade(avg string, r Nota) string {
	return average([]string{
		avg,
		r.EntAjustador,
		r.EntDigitacao,
		r.EntInicProf,
		r.EntAuxAdm,
		r.EntInfoSab,
		r.EntInfoDom,
		r.EntIngles,
		r.EntEletrica,
		r.EntAjustador,
	})
}

func parseFloat(input string) float64 {
	num, err := strconv.ParseFloat(strings.Replace(input, ",", ".", -1), 64)
	if err != nil {
		return 0
	}
	return num
}

func average(numbers []string) string {
	// Helper function to convert a string to a float64 number.
	// If there's a parsing error, return 0.
	parseFloat := func(input string) float64 {
		num, err := strconv.ParseFloat(strings.Replace(input, ",", ".", -1), 64)
		if err != nil {
			return 0
		}
		return num
	}

	// Calculate the sum of valid numbers and the count of valid elements
	var sum float64
	var count int
	for _, numStr := range numbers {
		num := parseFloat(numStr)
		if num != 0 {
			sum += num
			count++
		}
	}

	// Check if any valid numbers were found
	if count == 0 {
		return "0.00"
	}

	// Calculate the average
	avg := sum / float64(count)

	return fmt.Sprintf("%.2f", avg)
}

// resolve celular
func stripNonNumeric(s string) string {
	// Replace all non-numeric characters with an empty string
	// effectively stripping them from the original string
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, s)
}

// resolve idade
func getAge(now time.Time, age int, dobStr string) string {
	if age > 0 {
		return strconv.Itoa(age)
	}
	// age, err := strconv.Atoi(ageStr)
	// if err == nil && age > 0 {
	// 	// If the age field is filled and a valid positive number, return the provided age
	// 	return ageStr
	// }

	// If the age field is not filled or not a valid positive number,
	// calculate the age based on the date of birth
	dob, err := time.Parse("2006-01-02", dobStr)
	if err != nil {
		// Return an error value or handle the error as appropriate for your use case
		// For simplicity, we return an empty string here to indicate an invalid date of birth format.
		return ""
	}

	years := now.Year() - dob.Year()

	// Check if the birthday has already occurred this year by comparing month and day
	if now.Month() < dob.Month() || (now.Month() == dob.Month() && now.Day() < dob.Day()) {
		years--
	}

	// Convert the age to a string and return it
	return strconv.Itoa(years)
}

func getGradeAverage(pt, mat, log, red string) string {
	port, err := strconv.ParseFloat(
		strings.Replace(strings.TrimSpace(pt), ",", ".", 1), 64)
	if err != nil {
		port = 0
	}

	mate, err := strconv.ParseFloat(
		strings.Replace(strings.TrimSpace(mat), ",", ".", 1), 64)
	if err != nil {
		mate = 0
	}

	logi, err := strconv.ParseFloat(
		strings.Replace(strings.TrimSpace(log), ",", ".", 1), 64)
	if err != nil {
		mate = 0
	}

	reda, err := strconv.ParseFloat(
		strings.Replace(strings.TrimSpace(red), ",", ".", 1), 64)
	if err != nil {
		mate = 0
	}

	return fmt.Sprintf("%.2f", (port+mate+logi+reda)*10/35)
}

func printGrades(notas []Nota) {

	file, err := os.Create("../output/_fromErpNext.csv")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	printHeaderToCSV(notas[0], writer)

	for _, n := range notas {
		printFieldsToCSV(n, writer)
	}
}

func printHeaderToCSV(data interface{}, writer *csv.Writer) {
	value := reflect.ValueOf(data)
	if value.Kind() != reflect.Struct {
		fmt.Println("Input is not a struct")
		return
	}

	headers := []string{}

	for i := 0; i < value.NumField(); i++ {
		field := value.Type().Field(i)
		headers = append(headers, field.Name)
	}

	writer.Write(headers)
}

func printFieldsToCSV(data interface{}, writer *csv.Writer) {
	value := reflect.ValueOf(data)
	if value.Kind() != reflect.Struct {
		fmt.Println("Input is not a struct")
		return
	}

	values := []string{}

	for i := 0; i < value.NumField(); i++ {
		fieldValue := value.Field(i)
		values = append(values, fmt.Sprintf("%v", fieldValue.Interface()))
	}

	writer.Write(values)
}
