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

var (
	now    = time.Now()
	fields = []string{"full_name", "idade", "date_of_birth", "sab_op1", "sab_op2", "dom_op1",
		"dom_op2", "student_mobile_number", "matemática", "português", "lógica", "redação",
		"220_informática",
		"205_iniciação_profissional",
		"210_auxiliar_administrativo",
		"215_digitação",
		"225_montagem_de_micro",
		"235_elétrica",
		"245_climatização",
		"260_ajustador_soldador",
		"270_inglês",
		"program_sab", "program_sab2", "program_dom", "name"}
)

type Nota struct {
	Nome            string `json:"NOME,omitempty"`
	Celular         string `json:"CELULAR,omitempty"`
	Idade           string `json:"IDADE,omitempty"`
	Sabado1a        string `json:"1ª OPCAO S,omitempty"`
	Sabado2a        string `json:"2ª OPCAO S,omitempty"`
	Domingo1a       string `json:"1ª OPCAO D,omitempty"`
	Domingo2a       string `json:"2ª OPCAO D,omitempty"`
	Matematica      string `json:"MATEMATICA (0 - 10),omitempty"`
	Portugues       string `json:"PORTUGUES (0 - 10),omitempty"`
	Logica          string `json:"LOGICA (0 - 5),omitempty"`
	Redacao         string `json:"REDACAO (0 - 10),omitempty"`
	EntDigitacao    string `json:"DIGITACAO,omitempty"`
	EntInicProf     string `json:"INIC PROF,omitempty"`
	EntAuxAdm       string `json:"AUX ADM,omitempty"`
	EntInfoSab      string `json:"INFOR SAB,omitempty"`
	EntInfoDom      string `json:"INFOR DOM,omitempty"`
	EntIngles       string `json:"INGLES,omitempty"`
	EntEletrica     string `json:"ELETRICA,omitempty"`
	EntMontMicro    string `json:"MONT MICRO,omitempty"`
	EntAjustador    string `json:"AJUSTADOR,omitempty"`
	EntClimat       string `json:"CLIMATIZADOR,omitempty"`
	NotaProva       string `json:"NOTA PROVA,omitempty"`
	NotaFinal       string `json:"NOTA UNICA,omitempty"`
	SegChamada      int32  `json:"segunda_chamada,omitempty"`
	CursoApvSabado  string `json:"Curso_Aprovado,omitempty"`
	CursoApvDomingo string `json:"curso_aprovado_domingo,omitempty"`
	Name            string `json:"name"`
}

func GetNotas(baseUrl string) (*[]Nota, error) {

	queryParams := url.Values{}
	queryParams.Add("fields", fmt.Sprintf("[\"%s\"]", strings.Join(fields, "\",\"")))
	queryParams.Add("limit_start", "0")
	queryParams.Add("limit_page_length", "200")

	if os.Getenv("SEGUNDA_CHAMADA") == "1" {
		queryParams.Add("filters", fmt.Sprintf(
			"[[\"%s\",\"%s\",\"%d\"]]", "segunda_chamada", "=", 0))
	}

	queryParams.Add("filters", fmt.Sprintf(
		"[[\"%s\",\"%s\",\"%s\"]]", "status", "=", "applied"))

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
	fmt.Println("total de alunos inscritos: ", len(R.Data))
	return toNotas(&R), nil
}

type SysMeimeiResponse struct {
	Data []Applicant `json:"data"`
}

// Enviar esse lembrete com o PDF !
// pessoal que tiver alunos cadastrados
// menores de idade por favor peçam que os pais os acompanhem para assinar
// as fichas do Senai nesse final de semana

func toNotas(R *SysMeimeiResponse) *[]Nota {
	var notas = []Nota{}

	for _, r := range R.Data {
		gradesAverage := getGradeAverage(r.Portugues, r.Matematica, r.Logica, r.Redacao)
		finalGrade := getFinalGrade(gradesAverage, r)
		nota := Nota{
			Nome:         r.FullName,
			Celular:      stripNonNumeric(r.Celular),
			Idade:        getAge(now, int(math.Floor(r.Idade)), r.Date_of_birth),
			Sabado1a:     r.Opcao_1_sabado,
			Sabado2a:     r.Opcao_2_sabado,
			Domingo1a:    r.Opcao_1_domingo,
			Domingo2a:    r.Opcao_2_domingo,
			Matematica:   r.Matematica,
			Portugues:    r.Portugues,
			Logica:       r.Logica,
			Redacao:      r.Redacao,
			EntDigitacao: r.Digitação,
			EntInicProf:  r.Iniciacao_profissional,
			EntAuxAdm:    r.Auxiliar_administrtativo,
			EntInfoSab:   r.Informatica_basica_sabado,
			EntInfoDom:   r.Informática_basica_domingo,
			EntIngles:    r.Ingles,
			EntEletrica:  r.Eletrica,
			EntMontMicro: r.Montagem_de_micro,
			EntAjustador: r.Ajustador_soldador,
			EntClimat:    r.Climatizador,
			NotaProva:    gradesAverage,
			NotaFinal:    finalGrade,
			SegChamada:   r.SegundaChamada,
			Name:         r.Name,
		}
		notas = append(notas, nota)
	}
	go printGrades(notas)
	return &notas
}

func getFinalGrade(avg string, r Applicant) string {
	return average([]string{
		avg,
		r.Digitação,
		r.Iniciacao_profissional,
		r.Auxiliar_administrtativo,
		r.Informatica_basica_sabado,
		r.Informática_basica_domingo,
		r.Ingles,
		r.Eletrica,
		r.Montagem_de_micro,
		r.Ajustador_soldador,
		r.Climatizador,
	})
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

// curl --request GET \
//   --url 'https://larmeimei.org/api/resource/Student%20Applicant?fields=%5B%22first_name%22%2C%22last_name%22%2C%22idade%22%2C%22date_of_birth%22%2C%22op%C3%A7%C3%A3o_1_s%C3%A1bado%22%2C%22op%C3%A7%C3%A3o_2_s%C3%A1bado%22%2C%22op%C3%A7%C3%A3o_1_domingo%22%2C%22op%C3%A7%C3%A3o_2_domingo%22%2C%22student_mobile_number%22%2C%22matem%C3%A1tica%22%2C%22portugu%C3%AAs%22%2C%22l%C3%B3gica%22%2C%22reda%C3%A7%C3%A3o%22%2C%22inform%C3%A1tica_b%C3%A1sica_domingo%22%2C%22digita%C3%A7%C3%A3o%22%2C%22210_auxiliar_administrtativo%22%2C%22el%C3%A9trica%22%2C%22inicia%C3%A7%C3%A3o_profissional%22%2C%22montagem_de_micro%22%2C%22ajustador_soldador%22%2C%22ingl%C3%AAs%22%2C%22inform%C3%A1tica_b%C3%A1sica_s%C3%A1bado%22%5D' \
//   --header 'Authorization: token 202a75613683489:4795b74d1265c27' \
//   --header 'Content-Type: application/json' \
//   --cookie 'system_user=yes; full_name=Guest; user_id=Guest; user_image='

//https://larmeimei.org/api/resource/Student%20Applicant?fields=%5B%22first_name%22,%22last_name%22,%22idade%22,%22date_of_birth%22,%22op%C3%A7%C3%A3o_1_s%C3%A1bado%22,%22op%C3%A7%C3%A3o_2_s%C3%A1bado%22,%22op%C3%A7%C3%A3o_1_domingo%22,%22op%C3%A7%C3%A3o_2_domingo%22,%22student_mobile_number%22,%22matem%C3%A1tica%22,%22portugu%C3%AAs%22,%22l%C3%B3gica%22,%22reda%C3%A7%C3%A3o%22,%22inform%C3%A1tica_b%C3%A1sica_domingo%22,%22digita%C3%A7%C3%A3o%22,%22210_auxiliar_administrtativo%22,%22el%C3%A9trica%22,%22inicia%C3%A7%C3%A3o_profissional%22,%22montagem_de_micro%22,%22ajustador_soldador%22,%22ingl%C3%AAs%22,%22inform%C3%A1tica_b%C3%A1sica_s%C3%A1bado%22%5D
