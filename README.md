# meimei-classificador
Este projeto classifica os alunos nos cursos do Lar Meimei conforme suas notas. Neste repo todos os exemplos de notas são fictícios.

# chatGPT
Este projeto foi feito com ajuda do chatGPT

# Exemplo de dados de entrada

Arquivo: `input.json`

``` json

[
    {
        "NOME": "NOME DO ALUNO",
        "CELULAR": "11 9999-9999",
        "IDADE": "19",
        "Coluna1": "",
        "1ª OPCAO S": "E",
        "2ª OPCAO S": "D",
        "1ª OPCAO D": "",
        "2ª OPCAO D": "",
        "Coluna2": "",
        "MATEMATICA\n(0 - 10)": "4,0",
        "PORTUGUES\n(0 - 10)": "6,0",
        "LOGICA\n(0 - 5)": "1,0",
        "REDACAO\n(0 - 10)": "8,0",
        "Coluna3": "",
        "DIGITACAO A": "",
        "INIC PROF B": "",
        "AUX ADM C": "",
        "INFOR SAB D": "8,5",
        "INGLES E": "6,9",
        "ELETRICA F": "",
        "MONT MICRO G": "",
        "AJUSTADOR H": "",
        "Coluna4": "",
        "NOTA PROVA": "5,4",
        "NOTA UNICA": "6,9"
    },
    ...
]

```

## Preparação para execução

* Criar a pasta input, inserir o input.json com a lista de alunos conforme o padrão acima. 
* Criar a pasta output

* Adicionar o logo com o nome logo.jpg na pasta `C:\Users\<user_name>\AppData\Local\Temp` pois quando o  `wkhtmltopdf` executa, ele procura o logo nesta pasta.

* Instalar o [wkhtmltopdf](https://wkhtmltopdf.org/)

## Resultados

Na pasta output será gerado 3 arquivos para cada curso.

lista_curso.html, lista_curso.log e lista_curso.pdf