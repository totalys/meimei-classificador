package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
)

type Nota struct {
	Nome          string `json:"NOME,omitempty"`
	Celular       string `json:"CELULAR,omitempty"`
	Idade         string `json:"IDADE,omitempty"`
	Sabado1a      string `json:"1ª OPCAO S,omitempty"`
	Sabado2a      string `json:"2ª OPCAO S,omitempty"`
	Domingo1a     string `json:"1ª OPCAO D,omitempty"`
	Domingo2a     string `json:"2ª OPCAO D,omitempty"`
	Matematica    string `json:"MATEMATICA (0 - 10),omitempty"`
	Portugues     string `json:"PORTUGUES (0 - 10),omitempty"`
	Logica        string `json:"LOGICA (0 - 5),omitempty"`
	Redacao       string `json:"REDACAO (0 - 10),omitempty"`
	EntDigitacaoA string `json:"DIGITACAO A,omitempty"`
	EntInicProfB  string `json:"INIC PROF B,omitempty"`
	EntAUXADMC    string `json:"AUX ADM C,omitempty"`
	EntInfoSabD   string `json:"INFOR SAB D,omitempty"`
	EntInglesE    string `json:"INGLES E,omitempty"`
	EntEletricaF  string `json:"ELETRICA F,omitempty"`
	EntMontMicroG string `json:"MONT MICRO G,omitempty"`
	EntAjustadorH string `json:"AJUSTADOR H,omitempty"`
	NotaProva     string `json:"NOTA PROVA,omitempty"`
	NotaFinal     string `json:"NOTA UNICA,omitempty"`
}

type CoursesConfigMap struct {
	CoursesConfig map[string]CourseConfig
}

type CourseConfig struct {
	Name     string `json:"name"`
	Approved int    `json:"approved"`
	Waitlist int    `json:"wait"`
	Days     []int  `json:"days"`
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
	data, err := ioutil.ReadFile(src)
	checkErr(err)
	// Write data to dst
	err = ioutil.WriteFile(dst, data, 0644)
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
		// handle error
	}
	copy("../resources/logo.jpg", "../output/logo.jpg")

	file, err := os.Open("../input/input.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	decoder := json.NewDecoder(file)

	notas := []Nota{}
	err = decoder.Decode(&notas)
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

	for _, nota := range notas {
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
			continue
		}

		notaFinal, err := strconv.ParseFloat(strings.Replace(nota.NotaFinal, ",", ".", 1), 64)
		if err != nil {
			fmt.Printf("failed converting NotaFinal %s from %s. err: %s\n", nota.NotaFinal, nota.Nome, err.Error())
			continue
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

	for course, classified := range classifiedStudents {
		fmt.Printf("Curso: %s\n\n", courseConfigs[course].Name)
		fmt.Println("Nome \t Idade \t Contato \t Nota final")
		sb.WriteString(fmt.Sprintf("Curso: %s\n\n", courseConfigs[course].Name))
		sb.WriteString("Nome \t Idade \t Contato \t Nota final\n")

		// sort students by name
		classifieds := classified.ApprovedStudents
		sort.Slice(classifieds, func(i, j int) bool {
			return classifieds[i].Name < classifieds[j].Name
		})

		for _, student := range classifieds {
			fmt.Printf("%s \t %s \t %s\n", student.Name, student.Age, student.Phone)
			sb.WriteString(fmt.Sprintf("%s \t %s \t %s \t %s\n", student.Name, student.Age, student.Phone, fmt.Sprintf("%.2f", student.Grade)))

		}
		fmt.Println("Lista de espera:")
		sb.WriteString("Lista de espera:\n")
		for _, student := range classified.WaitlistStudents {
			fmt.Printf("%s \t %s \t %s\n", student.Name, student.Age, student.Phone)
			sb.WriteString(fmt.Sprintf("%s \t %s \t %s \t %s\n", student.Name, student.Age, student.Phone, fmt.Sprintf("%.2f", student.Grade)))

		}
		sb.WriteString("\n\n")
	}

	fmt.Printf("\n\nAlunos não classificados: \n\n")
	sb.WriteString("Alunos não classificados \n\n")
	for _, student := range students {
		if student.Approved == "" && !student.Waitlisted {
			fmt.Printf("%s \t %s \t %s \t cursos:%+v\n", student.Name, student.Age, student.Phone, student.Choices)
			sb.WriteString(fmt.Sprintf("%s \t %s \t %s \t %s \t cursos:%+v\n", student.Name, student.Age, student.Phone, fmt.Sprintf("%.2f", student.Grade), student.Choices))
		}
	}

	aprovados := "../output/aprovados.csv"
	err = ioutil.WriteFile(aprovados, []byte(sb.String()), 0644)
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
		fmt.Printf("\nCurso: %s\n\n", courseConfigs[course].Name)

		createReport(classified.ApprovedStudents,
			classified.WaitlistStudents, courseConfigs[course].Name)
	}
}

func createReport(approvedStudents, waitlist []Student, course string) {
	// Parse the template with the data for approved students and waitlist
	tmpl, err := template.New("test").Parse(html)
	if err != nil {
		log.Fatal(err)
	}
	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, struct {
		ApprovedStudents []Student
		Waitlist         []Student
		Course           string
	}{ApprovedStudents: approvedStudents,
		Waitlist: waitlist,
		Course:   course}); err != nil {
		log.Fatal(err)
	}

	// Convert the HTML to PDF using wkhtmltopdf

	// cmd := exec.Command("wkhtmltopdf.exe", "-", fmt.Sprintf("../output/lista_%s.pdf", course))
	// cmd.Stdin = bytes.NewReader(tpl.Bytes())
	// out, err := cmd.CombinedOutput()
	// if err != nil {
	// 	log.Println(err)
	// }
	// ioutil.WriteFile(fmt.Sprintf("../output/lista_%s.log", course), out, 0644)

	if err := ioutil.WriteFile(fmt.Sprintf("../output/lista_%s.html", course), tpl.Bytes(), 0644); err != nil {
		panic(err)
	}

	toPdf(fmt.Sprintf("../output/lista_%s.html", course), fmt.Sprintf("../output/lista_%s.pdf", course), course)
}

func toPdf(htmlFileName, pdfFileName, course string) {

	file, err := os.Open(htmlFileName)
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(file)
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		panic(err)
	}

	// pageOptions := wkhtmltopdf.NewPageOptions()
	// pageOptions.EnableLocalFileAccess.Set(true)
	// pageOptions.NoImages.Set(true)
	// pageOptions.Encoding.Set("utf-8")

	page := wkhtmltopdf.NewPageReader(reader)
	page.EnableLocalFileAccess.Set(true)
	page.NoImages.Set(false)
	page.Encoding.Set("utf-8")
	page.LoadMediaErrorHandling.Set("ignore")

	pdfg.AddPage(page)
	pdfg.Create()
	if err != nil {
		log.Fatal(err)
	}

	// Write buffer contents to file on disk
	err = pdfg.WriteFile(pdfFileName)
	if err != nil {
		log.Fatal(err)
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
