package jsonschema

import (
	"bytes"
	"encoding/json"
	"html/template"
)

var htmlTemplate string = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8" />
    <title>{{.Title}} </title>
    <link rel="stylesheet" href="https://unpkg.com/spectre.css/dist/spectre.min.css">
    <link rel="stylesheet" href="https://unpkg.com/spectre.css/dist/spectre-exp.min.css">
    <link rel="stylesheet" href="https://unpkg.com/spectre.css/dist/spectre-icons.min.css">
    <script src="https://cdn.jsdelivr.net/npm/@json-editor/json-editor@latest/dist/jsoneditor.min.js"></script>
	<script>
    // Set the default CSS theme and icon library globally
    JSONEditor.defaults.theme = 'spectre';
    JSONEditor.defaults.iconlib = 'spectre';
    </script>
    <style>
      .container {
        max-width:960px;
        margin: 0 auto
      }
    </style>
</head>
<body>
	<div class='container'>
    	<div id='editor_holder'></div>
		<button id='submit'>Submit (console.log)</button>
	</div>
    <script>
        // Initialize the editor with a JSON schema
        var editor = new JSONEditor(document.getElementById('editor_holder'), {
                schema: JSON.parse({{.Schema}})
            }
        );
        document.getElementById('submit').addEventListener('click', function () {
            console.log(editor.getValue());
        });
    </script>
</body>
</html>
`

func Form(name string) (string, error) {
	t, err := template.New("foo").Parse(htmlTemplate)
	if err != nil {
		return "", err
	}
	schame, err := GetSchema(name)
	if err != nil {
		return "", err
	}
	shString, err := json.Marshal(schame)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	_ = t.Execute(buf, struct {
		Title  string
		Schema string
	}{
		Title:  name,
		Schema: string(shString),
	})
	return buf.String(), nil
}
