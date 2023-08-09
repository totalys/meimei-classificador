package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
)

const (
	baseUrl string = "https://larmeimei.org/api/resource/Student%20Applicant"
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
	EntMontMicro string `json:"MONT MICRO,omitempty"`
	EntAjustador string `json:"AJUSTADOR,omitempty"`
	EntClimat    string `json:"CLIMATIZADOR,omitempty"`
	NotaProva    string `json:"NOTA PROVA,omitempty"`
	NotaFinal    string `json:"NOTA UNICA,omitempty"`
}

type CoursesConfigMap struct {
	Data          string `json:"data"`
	CoursesConfig map[string]CourseConfig
}

type CourseConfig struct {
	Name       string `json:"name"`
	Approved   int    `json:"approved"`
	Waitlist   int    `json:"wait"`
	Days       []int  `json:"days"`
	Sala       string `json:"sala"`
	DataInicio string `json:"dataInicio"`
}

type Student struct {
	Name       string   `json:"name"`
	Choices    []string `json:"choices"`
	Grade      float64  `json:"grade"`
	Approved   string   `json:"approved"`
	Waitlisted bool     `json:"waitlisted"`
	Age        string   `json:"idade"`
	Phone      string   `json:"celular"`
}

func (s Student) GetChoicesNames(choiceMap map[string]CourseConfig) (choiceNames []string) {

	for _, c := range s.Choices {
		choiceNames = append(choiceNames, fmt.Sprintf("%s[%s]", choiceMap[c].Name, c))
	}
	return choiceNames
}

type ClassifiedStudents struct {
	ApprovedStudents []Student `json:"approved_students"`
	WaitlistStudents []Student `json:"waitlist_students"`
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func copy(src string, dst string) {
	// Read all content of src to data, may cause OOM for a large file.
	data, err := os.ReadFile(src)
	checkErr(err)
	// Write data to dst
	err = os.WriteFile(dst, data, 0644)
	checkErr(err)
}

func main() {
	// deleting old results
	err := os.RemoveAll("../output/")
	if err != nil {
		panic(err)
	}
	folderPath := "../output"
	err = os.Mkdir(folderPath, os.ModePerm)
	if err != nil {
		panic(err)
	}
	copy("../resources/logo.jpg", "../output/logo.jpg")

	file, err := os.Open("../input/input.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	//decoder := json.NewDecoder(file)
	// notas := []Nota{}
	// err = decoder.Decode(&notas)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	notas, err := GetNotas(baseUrl)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Open the config file
	configFile, err := os.Open("../config/config.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer configFile.Close()

	// Decode the config file into a map
	configDecoder := json.NewDecoder(configFile)
	var courseConfigsMap CoursesConfigMap
	err = configDecoder.Decode(&courseConfigsMap)
	if err != nil {
		fmt.Println(err)
		return
	}

	var courseConfigs = courseConfigsMap.CoursesConfig

	students := []*Student{}

	for _, nota := range *notas {
		var choices []string

		if nota.Sabado1a != "" {
			choices = append(choices, nota.Sabado1a)
		}
		if nota.Sabado2a != "" {
			choices = append(choices, nota.Sabado2a)
		}
		if nota.Domingo1a != "" {
			choices = append(choices, nota.Domingo1a)
		}
		if nota.Domingo2a != "" {
			choices = append(choices, nota.Domingo2a)
		}
		if len(choices) == 0 {
			log.Printf("Error: Candidata(o):%s sem nenhum curso selecionado! Insira o curso escolhido e tente novamente", nota.Nome)
			os.Exit(1)
		}

		notaFinal, err := strconv.ParseFloat(strings.Replace(nota.NotaFinal, ",", ".", 1), 64)
		if err != nil {
			fmt.Printf("erro ao calcular a nota final. %s from %s. err: %s. O campo nota deve ser numérico\n", nota.NotaFinal, nota.Nome, err.Error())
			os.Exit(1)
		}
		students = append(students, &Student{
			Name:       nota.Nome,
			Choices:    choices,
			Grade:      notaFinal,
			Approved:   "",
			Waitlisted: false,
			Age:        nota.Idade,
			Phone:      nota.Celular,
		})
	}

	// Create a map to store the approved students and waitlist students for each course
	classifiedStudents := make(map[string]*ClassifiedStudents)

	// sort students by grade
	sort.Slice(students, func(i, j int) bool {
		return students[i].Grade > students[j].Grade
	})

	// Iterate through the list of students
	for _, student := range students {
		// Check if the student's grade is above or equal to the passing grade
		days_approved := []int{}
		if student.Grade >= 0 {
			// Iterate through the student's choices
			for _, choice := range student.Choices {
				// Check if the course has already been added to the map
				if _, ok := classifiedStudents[choice]; !ok {
					// If not, create a new entry in the map
					classifiedStudents[choice] = &ClassifiedStudents{}
				}
				// Append the student to the approved students list for their course
				if len(classifiedStudents[choice].ApprovedStudents) < courseConfigs[choice].Approved {
					// if  contans(courseConfigs[choice].Day)
					if canBeApproved(courseConfigs[choice].Days, days_approved) {
						if student.Approved != "" {
							student.Approved = fmt.Sprintf("%s|%s", student.Approved, choice)
						} else {
							student.Approved = choice
						}

						classifiedStudents[choice].ApprovedStudents = append(classifiedStudents[choice].ApprovedStudents, *student)
						days_approved = append(days_approved, courseConfigs[choice].Days...)
					} else {
						fmt.Printf("Aluno %s seria aprovado para o curso %s no dia %v mas não foi porque já está aprovado em %s nos dias: %v \n",
							student.Name, choice, courseConfigs[choice].Days, student.Approved, days_approved)
					}
					continue
				} else if len(classifiedStudents[choice].WaitlistStudents) < courseConfigs[choice].Waitlist {
					student.Waitlisted = true
					classifiedStudents[choice].WaitlistStudents = append(classifiedStudents[choice].WaitlistStudents, *student)
					continue
				}
			}
		}
	}
	fmt.Println("Lista para divulgação")
	for course, classified := range classifiedStudents {
		fmt.Printf("\nCurso: %s\n\n", courseConfigs[course].Name)
		fmt.Println("Nome \t Idade")

		// sort students by grade
		classifieds := classified.ApprovedStudents
		sort.Slice(classifieds, func(i, j int) bool {
			return classifieds[i].Name < classifieds[j].Name
		})

		for _, student := range classifieds {
			fmt.Printf("%s \t %s\n", student.Name, student.Age)

		}
		fmt.Println("Lista de espera:")

		for _, student := range classified.WaitlistStudents {
			fmt.Printf("%s \t %s\n", student.Name, student.Age)

		}
	}

	fmt.Println("Lista para coordenadores")
	sb := strings.Builder{}
	sbAll := strings.Builder{}

	for course, classified := range classifiedStudents {
		fmt.Printf("Curso: %s\n\n", courseConfigs[course].Name)
		fmt.Println("Nome \t Idade \t Contato \t Nota final \t link whatsapp")
		sb.WriteString(fmt.Sprintf("Curso: %s\n\n", courseConfigs[course].Name))
		sb.WriteString("Nome \t Idade \t Contato \t Nota final \t link whatsapp\n")

		// sort students by name
		classifieds := classified.ApprovedStudents
		sort.Slice(classifieds, func(i, j int) bool {
			return classifieds[i].Name < classifieds[j].Name
		})

		for _, student := range classifieds {
			fmt.Printf("%s \t %s \t %s\n", student.Name, student.Age, student.Phone)
			sb.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\n",
				student.Name, student.Age, student.Phone, fmt.Sprintf("%.2f", student.Grade), fmt.Sprintf("http://wa.me/55%s", student.Phone)))
		}

		fmt.Println("Lista de espera:")
		sb.WriteString("\nLista de espera:\n")
		for _, student := range classified.WaitlistStudents {
			fmt.Printf("%s \t %s \t %s\n", student.Name, student.Age, student.Phone)
			sb.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\n",
				student.Name, student.Age, student.Phone, fmt.Sprintf("%.2f", student.Grade), fmt.Sprintf("http://wa.me/55%s", student.Phone)))
		}
		sb.WriteString("\n\n")

		go writeExcelFile(courseConfigs[course].Name, sb.String())
		sbAll.WriteString(sb.String())
		sb.Reset()
	}
	sb.WriteString(sbAll.String())
	fmt.Printf("\n\nAlunos não classificados: \n\n")
	sb.WriteString("Alunos não classificados \n\n")
	for _, student := range students {
		if student.Approved == "" && !student.Waitlisted {
			choiceNames := student.GetChoicesNames(courseConfigs)
			fmt.Printf("%s \t %s \t %s \t cursos:%+v\n", student.Name, student.Age, student.Phone, choiceNames)
			sb.WriteString(fmt.Sprintf("%s \t %s \t %s \t %s \t cursos:%+v\n", student.Name, student.Age, student.Phone, fmt.Sprintf("%.2f", student.Grade), choiceNames))
		}
	}

	aprovados := "../output/aprovados.csv"
	err = os.WriteFile(aprovados, []byte(sb.String()), 0644)
	if err != nil {
		fmt.Println(err)
	}

	b, err := json.Marshal(classifiedStudents)
	if err != nil {
		panic(err)
	}

	// Write the JSON string to a file
	f, err := os.Create("../output/results.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = f.Write(b)
	if err != nil {
		panic(err)
	}

	for course, classified := range classifiedStudents {
		createReport(
			classified.ApprovedStudents,
			classified.WaitlistStudents,
			courseConfigs[course].Name,
			courseConfigsMap.Data,
			courseConfigs[course].DataInicio,
			courseConfigs[course].Sala)
	}

	checkInconsistencies(courseConfigs, classifiedStudents)

}

func createReport(approvedStudents, waitlist []Student, course, data, dataInicio, sala string) {

	// Parse the template with the data for approved students and waitlist
	tmpl, err := template.ParseFiles("template.html")
	if err != nil {
		log.Fatal(err)
	}
	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, struct {
		ApprovedStudents []Student
		Waitlist         []Student
		Course           string
		Data             string
		DataInicio       string
		Sala             string
	}{ApprovedStudents: approvedStudents,
		Waitlist:   waitlist,
		Course:     strings.ToUpper(course),
		Data:       data,
		DataInicio: dataInicio,
		Sala:       sala}); err != nil {
		log.Fatal(err)
	}

	// Convert the HTML to PDF using wkhtmltopdf
	cmd := exec.Command("wkhtmltopdf.exe",
		"--enable-local-file-access",
		"--encoding",
		"utf-8",
		"-", fmt.Sprintf("../output/lista_%s.pdf", course))
	cmd.Stdin = bytes.NewReader(tpl.Bytes())
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err)
	}
	os.WriteFile(fmt.Sprintf("../output/lista_%s.log", course), out, 0644)

	if err := os.WriteFile(fmt.Sprintf("../output/lista_%s.html", course), tpl.Bytes(), 0644); err != nil {
		panic(err)
	}
}

func canBeApproved(currentChoiceDays, alreadyApprovedDays []int) bool {
	exists := false
	for _, val1 := range currentChoiceDays {
		for _, val2 := range alreadyApprovedDays {
			if val1 == val2 {
				exists = true
				break
			}
		}
	}
	return !exists
}

func checkInconsistencies(courseConfigs map[string]CourseConfig, classifiedStudents map[string]*ClassifiedStudents) {

	fmt.Println("Verificando possíveis inconsistências: Listando alunos classificados em mais de um curso para o mesmo dia da semana:")
	stds := make(map[string][]int)
	for course, c := range courseConfigs {
		for _, s := range classifiedStudents[course].ApprovedStudents {
			stds[s.Name] = append(stds[s.Name], c.Days...)
		}
	}

	var multiCourseSameDayStudents []Student
	for s, c := range stds {
		if len(c) > 1 {
			for i := 0; i < len(c); i++ {
				for j := i + 1; j < len(c); j++ {
					if c[i] == c[j] {
						multiCourseSameDayStudents = append(multiCourseSameDayStudents, Student{Name: s})
						break
					}
				}
			}
		}
	}

	fmt.Println("Students with more than one course on the same day:", multiCourseSameDayStudents)
}

func writeExcelFile(curso, content string) {

	// Create a new Excel file
	file := excelize.NewFile()
	lines := strings.Split(content, "\n")

	sheetName := curso
	file.NewSheet(sheetName)
	file.SetColWidth(sheetName, "B", "B", 60)
	file.SetColWidth(sheetName, "C", "C", 10)
	file.SetColWidth(sheetName, "D", "D", 15)
	file.SetColWidth(sheetName, "E", "E", 10)
	file.SetColWidth(sheetName, "F", "F", 30)
	for rowIdx, line := range lines {
		colls := strings.Split(line, "\t")

		for collIdx, c := range colls {
			cell := excelize.ToAlphaString(collIdx+1) + fmt.Sprintf("%d", rowIdx+1)
			file.SetCellValue(sheetName, cell, c)

		}

		// creates the whatsapp link
		if len(colls) > 4 {
			url := colls[len(colls)-1]
			cell := excelize.ToAlphaString(len(colls)) + fmt.Sprintf("%d", rowIdx+1)
			file.SetCellHyperLink(sheetName, cell, url, "External")

			style, _ := file.NewStyle(`{"font":{"color":"#1265BE","underline":"single"}}`)
			file.SetCellStyle(sheetName, cell, cell, style)
		}
	}

	// delete default folder created and not used.
	file.DeleteSheet("Sheet1")

	outputFileName := fmt.Sprintf("../output/excel_aprovados_%s.xlsx", curso)
	if err := file.SaveAs(outputFileName); err != nil {
		log.Println("error saving excel file", err)
	}

	fmt.Printf("Excel file saved as %s\n", outputFileName)

}
