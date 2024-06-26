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
	_ "unicode/utf8"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/totalys/meimei-classificador/pkg/domain"
	"github.com/totalys/meimei-classificador/pkg/extractor"
	"github.com/totalys/meimei-classificador/pkg/updater"
)

const (
	baseUrl string = "https://larmeimei.org/api/resource/LM%20Interview"
)

type ClassifiedStudents struct {
	ApprovedStudents []domain.Student `json:"approved_students"`
	WaitlistStudents []domain.Student `json:"waitlist_students"`
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

	notas, err := extractor.GetNotas(baseUrl)
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
	var courseConfigsMap domain.CoursesConfigMap
	err = configDecoder.Decode(&courseConfigsMap)
	if err != nil {
		fmt.Println(err)
		return
	}

	var courseConfigs = courseConfigsMap.CoursesConfig

	students := []*domain.Student{}

	var candidatosIncompletos = make([]string, 0)

	for _, nota := range notas {
		var choices []string

		if nota.Domingo1a != "" {
			var choice = nota.Domingo1a
			choices = append(choices, choice)
		}

		if nota.Domingo2a != "" {
			var choice = nota.Domingo2a
			choices = append(choices, choice)
		}

		if nota.Sabado1a != "" {
			var choice = nota.Sabado1a
			choices = append(choices, choice)
		}

		if nota.Sabado2a != "" {
			var choice = nota.Sabado2a
			choices = append(choices, choice)
		}

		if len(choices) == 0 {
			log := fmt.Sprintf("sem curso pretendido: %s	telefone: %s", nota.Nome, nota.Celular)
			candidatosIncompletos = append(candidatosIncompletos,
				log)
			fmt.Println(log)
			continue
		}

		notaFinal, err := strconv.ParseFloat(strings.Replace(nota.NotaFinal, ",", ".", 1), 64)
		if err != nil {
			fmt.Printf("erro ao calcular a nota final. %s from %s. err: %s. O campo nota deve ser numérico\n", nota.NotaFinal, nota.Nome, err.Error())
			os.Exit(1)
		}

		students = append(students, &domain.Student{
			Name:        nota.Nome,
			Choices:     choices,
			Grade:       notaFinal,
			Approved:    []domain.Approved{},
			Waitlisted:  false,
			Age:         nota.Idade,
			Phone:       nota.Celular,
			SegChamada:  nota.SegChamada,
			DocName:     nota.Name,
			Entrevistas: nota.Entrevistas,
		})
	}

	if len(candidatosIncompletos) > 0 {
		log.Printf("Error: Alguns candidatos sem nenhum curso selecionado! Insira o curso escolhido para estes candidatos e tente novamente")
		os.Exit(1)
	}

	// Create a map to store the approved students and waitlist students for each course
	classifiedStudents := make(map[string]*ClassifiedStudents)

	// sort students by grade
	sort.Slice(students, func(i, j int) bool {
		return students[i].Grade > students[j].Grade
	})

	fmt.Printf("Total de alunos: %d\n", len(students))

	// Iterate through the list of students
	for _, student := range students {
		// Check if the student's grade is above or equal to the passing grade
		days_approved := []int{}
		days_approved_or_waitlisted := []int{}
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

					if canBeApproved(courseConfigs[choice], student, days_approved, choice) {
						student.Approved = append(student.Approved, domain.Approved{
							Course:  choice,
							Days:    courseConfigs[choice].Days,
							IsSenai: courseConfigs[choice].IsSenai,
						})
						classifiedStudents[choice].ApprovedStudents = append(classifiedStudents[choice].ApprovedStudents, *student)
						days_approved = append(days_approved, courseConfigs[choice].Days...)
						days_approved_or_waitlisted = append(days_approved_or_waitlisted, courseConfigs[choice].Days...)
					} else {
						fmt.Printf("Aluno %s seria aprovado para o curso %s, do Senai? [%t], no dia(s) %v mas não foi porque já está aprovado em %+v no(s) dia(s): %v \n",
							student.Name, choice, courseConfigs[choice].IsSenai, courseConfigs[choice].Days, student.Approved, days_approved)
					}
					continue
				} else if len(classifiedStudents[choice].WaitlistStudents) < courseConfigs[choice].Waitlist {

					if canBeWaitListed(courseConfigs[choice], days_approved_or_waitlisted) {
						student.Waitlisted = true
						classifiedStudents[choice].WaitlistStudents = append(classifiedStudents[choice].WaitlistStudents, *student)
						// days_approved_or_waitlisted = append(days_approved_or_waitlisted, courseConfigs[choice].Days...)
					} else {
						fmt.Printf("Aluno %s seria adicionado na lista de espera de %s no dia(s) %v mas não foi porque já está em uma lista de espera em pelo menos um curso nos dias %v \n",
							student.Name, choice, courseConfigs[choice].Days, days_approved)
					}

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

		c, ok := courseConfigs[course]
		if !ok {
			fmt.Printf("curso: [%s] não configurado. Será ignorado nos relatórios\n", course)
			continue
		}
		courseName := c.Name

		fmt.Printf("Curso: %s\n\n", courseName)
		fmt.Println("Nome \t Idade \t Contato \t Nota final \t link whatsapp \t 2a chamada")
		sb.WriteString(fmt.Sprintf("Curso: %s\n\n", courseName))
		sb.WriteString("Nome \t Idade \t Contato \t Nota final \t link whatsapp \t 2a chamada\n")

		// sort students by name
		classifieds := classified.ApprovedStudents
		sort.Slice(classifieds, func(i, j int) bool {
			return classifieds[i].Name < classifieds[j].Name
		})

		for _, student := range classifieds {
			fmt.Printf("%s \t %s \t %s\n", student.Name, student.Age, student.Phone)
			sb.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%d\n",
				student.Name, student.Age, student.Phone, fmt.Sprintf("%.2f", student.Grade), fmt.Sprintf("http://wa.me/55%s", student.Phone), student.SegChamada))
		}

		fmt.Println("Lista de espera:")
		sb.WriteString("\nLista de espera:\n")
		for _, student := range classified.WaitlistStudents {
			fmt.Printf("%s \t %s \t %s\n", student.Name, student.Age, student.Phone)
			sb.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%d\n",
				student.Name, student.Age, student.Phone, fmt.Sprintf("%.2f", student.Grade),
				fmt.Sprintf("http://wa.me/55%s", student.Phone), student.SegChamada))
		}
		sb.WriteString("\n\n")

		writeExcelFile(courseName, sb.String())
		sbAll.WriteString(sb.String())
		sb.Reset()
	}

	sb.WriteString(sbAll.String())
	fmt.Printf("\n\nAlunos não classificados: \n\n")
	sb.WriteString("Alunos não classificados \n\n")
	for _, student := range students {
		if len(student.Approved) == 0 && !student.Waitlisted {
			choiceNames := student.GetChoicesNames(courseConfigs)

			if len(choiceNames) == 0 {
				continue
			}

			fmt.Printf("%s \t %s \t %s \t cursos:%+v\n", student.Name, student.Age, student.Phone, choiceNames)
			sb.WriteString(fmt.Sprintf("%s \t %s \t %s \t %s \t cursos:%+v\n",
				student.Name, student.Age, student.Phone, fmt.Sprintf("%.2f", student.Grade), choiceNames))
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

	if os.Getenv("UPDATE") == "1" {
		err := updater.UpdateGrades(courseConfigs, baseUrl, students)
		if err != nil {
			log.Fatal(err)
		}
	}

}

func createReport(approvedStudents, waitlist []domain.Student, course, data, dataInicio, sala string) {

	course = strings.ReplaceAll(course, "/", " - ")
	// Parse the template with the data for approved students and waitlist
	tmpl, err := template.ParseFiles("template.html")
	if err != nil {
		log.Fatal(err)
	}
	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, struct {
		ApprovedStudents []domain.Student
		Waitlist         []domain.Student
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

func canBeWaitListed(currentChoice domain.CourseConfig, alreadyApprovedDays []int) bool {
	exists := false

	// Valida se o aluno já foi aprovado em um curso no dia do curso solicitado
	for _, val1 := range currentChoice.Days {
		for _, val2 := range alreadyApprovedDays {
			if val1 == val2 {
				exists = true
				break
			}
		}
	}
	return !exists
}

func canBeApproved(currentChoice domain.CourseConfig, student *domain.Student, alreadyApprovedDays []int, choice string) bool {
	exists := false

	// Valida se o aluno já foi aprovado em um curso no dia do curso solicitado
	for _, val1 := range currentChoice.Days {
		for _, val2 := range alreadyApprovedDays {
			if val1 == val2 {
				exists = true
				break
			}
		}
	}

	// Valida se o aluno já foi aprovado neste curso que ele quer fazer
	for _, courseApproved := range student.Approved {
		if strings.Split(currentChoice.Name, "/")[0] == strings.Split(courseApproved.Course, "/")[0] {
			exists = true
			break
		}
	}

	// Valida se o aluno não tirou 0 na entrevista do curso em que está tentando se inscrever
	if student.Entrevistas[choice] == 0 {
		fmt.Printf("Aluno %s seria aprovado em %s mas não foi porque tirou %f na entrevista.", student.Name, currentChoice.Name, student.Entrevistas[choice])
		exists = true
	}

	return !exists
}

func checkInconsistencies(courseConfigs map[string]domain.CourseConfig, classifiedStudents map[string]*ClassifiedStudents) {

	fmt.Println("Verificando possíveis inconsistências: Listando alunos classificados em mais de um curso para o mesmo dia da semana:")
	stds := make(map[string][]int)
	for course, c := range courseConfigs {
		classifieds, ok := classifiedStudents[course]
		if !ok {
			log.Printf("nenhum aluno classificado para o curso: %s", course)
			continue
		}
		for _, s := range classifieds.ApprovedStudents {
			stds[s.Name] = append(stds[s.Name], c.Days...)
		}
	}

	var multiCourseSameDayStudents []domain.Student
	for s, c := range stds {
		if len(c) > 1 {
			for i := 0; i < len(c); i++ {
				for j := i + 1; j < len(c); j++ {
					if c[i] == c[j] {
						multiCourseSameDayStudents = append(multiCourseSameDayStudents, domain.Student{Name: s})
						break
					}
				}
			}
		}
	}

	fmt.Println("Alunos aprovados em mais de um curso no mesmo dia:", multiCourseSameDayStudents)
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
