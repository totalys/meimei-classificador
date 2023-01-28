package main

const html = `
    <html>
    <head>
        <style>
            table {
                border-collapse: collapse;
                margin: 0 auto;
            }
            th, td {
                border: 1px solid black;
                padding: 5px;
            }
        </style>
    </head>
    <body>
        <img src="logo.jpg" alt="logo" style="width: 200px;">
        <div style="text-align: center;">
            <h1>Lista de aprovados. Curso: {{.Course}}</h1>
        </div>
       
        <table>
            <tr>
                <th>Nome</th>
                <th>Idade</th>
            </tr>
            {{range .ApprovedStudents}}
            <tr>
                <td>{{.Name}}</td>
                <td>{{.Age}}</td>
            </tr>
            {{end}}
        </table>
        <div style="text-align: center;">
            <h1>Lista de Espera. Curso: {{.Course}}</h1>
        </div>
        <table>
            <tr>
                <th>Nome</th>
                <th>Idade</th>
            </tr>
            {{range .Waitlist}}
            <tr>
                <td>{{.Name}}</td>
                <td>{{.Age}}</td>
            </tr>
            {{end}}
        </table>
    </body>
    </html>`
