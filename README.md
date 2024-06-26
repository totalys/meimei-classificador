# meimei-classificador
Este projeto classifica os alunos nos cursos do Lar Meimei conforme suas notas. Neste repo todos os exemplos de notas são fictícios.

## chatGPT
Este projeto foi feito com ajuda do chatGPT

## Exemplo de dados de entrada

Vide domain.go Struct: Applicant
Os dados extraídos do ErpNext.

### Api do ErpNext

url: https://larmeimei.org/api/resource/LM%20Interview

domínio lar Meimei: larmeimei.org
doc_type: Student Applicant => urlEncoded: LM%20Interview

**Headers**

"Content-Type": "application/json"
"Authorization", your_api_key

**Query params**

* Controle de paginação

    * "limit_start": "0"
    * "limit_page_length": "200"
* Campos a serem extraídos
    * "fields": [todos os campos do documento separados por vírgula]
 
## Documentação

[ErpNex API](https://frappeframework.com/docs/user/en/api/rest)


## Preparação para execução

* Configurar as envs:
*    LARMEIMEI_[seu_nome]_API_KEY
*    UPDATE: 0 ou 1 para ajustar o status do aluno como aprovado no curso
*    SEGUNDA_CHAMADA: 0 ou 1 para situações de segunda_chamada.

* Criar a pasta output

* Adicionar o logo com o nome logo.jpg na pasta `C:\Users\<user_name>\AppData\Local\Temp` pois quando o  `wkhtmltopdf` executa, ele procura o logo nesta pasta.

* Instalar o [wkhtmltopdf](https://wkhtmltopdf.org/)

## Resultados

Na pasta output serão gerados 4 arquivos para cada curso.

lista_curso.html, lista_curso.log, lista_curso.pdf e excel_aprovados.xlsx com o contato e link para whatsapp para facilitar a criação das listas.

## Para rodar o classificador

$ `cd app`<br>

$ `go run classificador.go`

## Para buildar e criar um executável

$ `go build -o meimei_classificador.exe classificador.go`

A pasta /output será toda deletada e os arquivos serão gerados conforme as regras de input (extrator do erpNext) e configuration.json

## Envs

ENV LARMEIMEI_USER_API_KEY: api key and secret. ex: bearer [api_key]:[api_secret] criada na área de segurança do perfil de usuário do ErpNext

$ `export LARMEIMEI_USER_API_KEY=bearer XXXX:XXXX`
