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
					sabado = a.Course
				}
				if d == 7 {
					domingo = a.Course
				}
			}
		}

		var data = struct {
			CursoAprovadoSabado  string `json:"curso_aprovado"`
			CursoAprovadoDomingo string `json:"curso_aprovado_domingo"`
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

		req.Header.Set("Authorization", os.Getenv("LARMEIMEI_TIAGO_API_KEY"))
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
