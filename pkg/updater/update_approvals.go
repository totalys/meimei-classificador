package updater

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/totalys/meimei-classificador/pkg/domain"
)

var course_student_group_map = map[string]string{
	"220 - Informática Básica - LM/Domingo": "220-SG-2024/Dom (INFORMÁTICA BÁSICA)",
	"220 - Informática Básica - LM/Sabado":  "220-SG-2024/Sáb (INFORMÁTICA BÁSICA)",
	"260 - Ajustador Soldador - LM":         "260-SG-2024 (AJUSTADOR SOLDADOR)",
	"231 - Auxiliar Eletricista - LM":       "231-SG-2024/Sem01 (AUXILIAR ELETRICISTA)",
	"225 - Montagem de Micros - LM":         "225-SG-2024 (MONTAGEM DE MICRO)",
	"215 - Auxiliar Administrativo - LM":    "215-SG-2024 (AUX.ADMINISTRATIVO)",
	"232 - Eletricista Instalador - LM":     "232-SG-2024 (ELETRICISTA INSTALADOR)",
	"255 - Digitação - LM":                  "255-SG-2024/Sem01 (DIGITAÇÃO)",
	"270 - Inglês - LM":                     "270-SG-2024 (INGLÊS)",
	"250 - Iniciação Profissional - LM":     "251-SG-2024/Sem01 (INIC.PROF.BÁSICO)",
}

func UpdateGrades(coursesConfigs map[string]domain.CourseConfig, baseUrl string, students []*domain.Student) error {
	log.Println("atualizando o Sysmeimei 2.0 ...")
	for i, s := range students {

		if i%10 == 0 {
			log.Print(" ..", i)
		}

		requestURL := fmt.Sprintf("%s/%s", baseUrl, s.DocName)

		domingo := ""
		sabado := ""

		for _, a := range s.Approved {
			for _, d := range a.Days {
				if d == 6 {
					sabado = course_student_group_map[a.Course]
				}
				if d == 7 {
					domingo = course_student_group_map[a.Course]
				}
			}
		}

		if sabado == "" && domingo == "" {
			continue
		}

		var data = struct {
			CursoAprovadoSabado  string `json:"student_group_sab"`
			CursoAprovadoDomingo string `json:"student_group_dom"`
		}{
			CursoAprovadoSabado:  sabado,
			CursoAprovadoDomingo: domingo,
		}

		b, err := json.Marshal(&data)
		if err != nil {
			return fmt.Errorf("error marshaling approved courses %w", err)
		}

		req, err := http.NewRequest(http.MethodPut, requestURL, bytes.NewBuffer(b))
		if err != nil {
			return fmt.Errorf("error creating request %w", err)
		}

		req.Header.Set("Authorization", os.Getenv("LARMEIMEI_USER_API_KEY"))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		response, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("error making the request: %+w", err)
		}
		defer response.Body.Close()

		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("error reading response: %+w", err)
		}

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("error received response: %s,\n request: %+v, \n %+w", responseBody, req, err)
		}
	}

	log.Println("alunos aprovados atualizados no Sysmeimei 2.0")

	return nil
}
