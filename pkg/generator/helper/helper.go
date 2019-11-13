package helper

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/SAP/jenkins-library/pkg/config"
)

type stepInfo struct {
	CobraCmdFuncName string
	CreateCmdVar     string
	ExportPrefix     string
	FlagsFunc        string
	Long             string
	Metadata         []config.StepParameters
	OSImport         bool
	Short            string
	StepFunc         string
	StepName         string
}

type enumInfo struct {
	Name   string
	Values []string
}

//StepGoTemplate ...
const stepGoTemplate = `package cmd

import (
	{{if .OSImport}}"os"{{end}}

	{{if .ExportPrefix}}{{ .ExportPrefix }} "github.com/SAP/jenkins-library/cmd"{{end}}
	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/spf13/cobra"
)

type {{ .StepName }}Options struct {
	{{- range $key, $value := .Metadata }}
	{{if eq $value.Type "enum"}}{{ $value.Name}} string{{else}}{{ $value.Name | golangName }} {{ $value.Type }} {{end}}` + "`json:\"{{$value.Name}},omitempty\"`" + `{{end}}
}

var my{{ .StepName | title}}Options {{.StepName}}Options
var {{ .StepName }}StepConfigJSON string

// {{.CobraCmdFuncName}} {{.Short}}
func {{.CobraCmdFuncName}}() *cobra.Command {
	metadata := {{ .StepName }}Metadata()
	var {{.CreateCmdVar}} = &cobra.Command{
		Use:   "{{.StepName}}",
		Short: "{{.Short}}",
		Long: {{ $tick := "` + "`" + `" }}{{ $tick }}{{.Long | longName }}{{ $tick }},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			log.SetStepName("{{ .StepName }}")
			log.SetVerbose({{if .ExportPrefix}}{{ .ExportPrefix }}.{{end}}GeneralConfig.Verbose)
			return {{if .ExportPrefix}}{{ .ExportPrefix }}.{{end}}PrepareConfig(cmd, &metadata, "{{ .StepName }}", &my{{ .StepName | title}}Options, {{if .ExportPrefix}}{{ .ExportPrefix }}.{{end}}OpenPiperFile)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return {{.StepName}}(my{{ .StepName | title }}Options)
		},
	}

	{{.FlagsFunc}}({{.CreateCmdVar}})
	return {{.CreateCmdVar}}
}

func {{.FlagsFunc}}(cmd *cobra.Command) {
	{{- range $key, $value := .Metadata }}
	cmd.Flags().{{ $value.Type | flagType }}(&my{{ $.StepName | title }}Options.{{ $value.Name | golangName }}, "{{ $value.Name }}", {{ $value.Default }}, "{{ $value.Description }}"){{ end }}
	{{- printf "\n" }}
	{{- range $key, $value := .Metadata }}{{ if $value.Mandatory }}
	cmd.MarkFlagRequired("{{ $value.Name }}"){{ end }}{{ end }}
}

// retrieve step metadata
func {{ .StepName }}Metadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{{- range $key, $value := .Metadata }}
					{
						Name:      "{{ $value.Name }}",
						Scope:     []string{{ "{" }}{{ range $notused, $scope := $value.Scope }}"{{ $scope }}",{{ end }}{{ "}" }},
						Type:      "{{ $value.Type }}",
						Mandatory: {{ $value.Mandatory }},
					},{{ end }}
				},
			},
		},
	}
	return theMetaData
}
`

//StepTestGoTemplate ...
const stepTestGoTemplate = `package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test{{.CobraCmdFuncName}}(t *testing.T) {

	testCmd := {{.CobraCmdFuncName}}()

	// only high level testing performed - details are tested in step generation procudure
	assert.Equal(t, "{{.StepName}}", testCmd.Use, "command name incorrect")

}
`

const enumGoTemplate = `package cmd

import (
	"fmt"
	"errors"
)
type {{.Name}} int

const (
	{{$enumName := .Name}}
	UNKNOWN_{{$enumName | toUpperCase}} {{$enumName}} = iota -1
	{{- range $value := .Values}}
	{{$value}} {{$enumName}} = iota{{end}}
)

func (w {{.Name}}) String() string {
	return [...]string{
		"UNKNOWN_{{.Name | toUpperCase}}",
		{{- range $value := .Values}}
		"{{$value}}",{{end}}
	}[w]
}

func ValueOf(str string) ({{.Name}}, error) {
	switch str {
		case "UNKNOWN_{{.Name | toUpperCase}}": return UNKNOWN_{{.Name | toUpperCase}}, nil
		{{- range $value := .Values}}
		case "{{$value}}": return {{$value}}, nil{{end}}
		default : return UNKNOWN_{{.Name | toUpperCase}}, errors.New(fmt.Sprintf("Unknown {{.Name}}: '%s'", str))
	}
}
`

// ProcessMetaFiles generates step coding based on step configuration provided in yaml files
func ProcessMetaFiles(metadataFiles []string, openFile func(s string) (io.ReadCloser, error), writeFile func(filename string, data []byte, perm os.FileMode) error, exportPrefix string) error {
	for key := range metadataFiles {

		var stepData config.StepData

		configFilePath := metadataFiles[key]

		metadataFile, err := openFile(configFilePath)
		checkError(err)
		defer metadataFile.Close()

		fmt.Printf("Reading file %v\n", configFilePath)

		err = stepData.ReadPipelineStepData(metadataFile)
		checkError(err)

		fmt.Printf("Step name: %v\n", stepData.Metadata.Name)

		osImport := false
		osImport, err = setDefaultParameters(&stepData)
		checkError(err)

		myStepInfo := getStepInfo(&stepData, osImport, exportPrefix)

		step := stepTemplate(myStepInfo)
		err = writeFile(fmt.Sprintf("cmd/%v_generated.go", stepData.Metadata.Name), step, 0644)
		checkError(err)

		test := stepTestTemplate(myStepInfo)
		err = writeFile(fmt.Sprintf("cmd/%v_generated_test.go", stepData.Metadata.Name), test, 0644)
		checkError(err)
	}
	return nil
}

func processEnum(name string, values []string, writeFile func(filename string, data []byte, perm os.FileMode) error) {

	enumInfo := getEnumInfo(name, values)
	enum := enumTemplate(enumInfo)
	err := writeFile(fmt.Sprintf("cmd/%v_enum_generated.go", firstCharToLowerCase(enumInfo.Name)), enum, 0644)
	checkError(err)
}

func firstCharToLowerCase(s string) string {
	return strings.ToLower(string(s[0])) + s[1:]
}

func openMetaFile(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

func fileWriter(filename string, data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filename, data, perm)
}

func setDefaultParameters(stepData *config.StepData) (bool, error) {
	//ToDo: custom function for default handling, support all relevant parameter types
	osImportRequired := false
	for k, param := range stepData.Spec.Inputs.Parameters {

		if param.Type == "enum" {
			processEnum(param.Name, param.Values, fileWriter)
		}
		if param.Default == nil {
			switch param.Type {
			case "string":
				param.Default = fmt.Sprintf("os.Getenv(\"PIPER_%v\")", param.Name)
				osImportRequired = true
			case "bool":
				// ToDo: Check if default should be read from env
				param.Default = "false"
			case "[]string":
				// ToDo: Check if default should be read from env
				param.Default = "[]string{}"
			case "enum":
				param.Default = fmt.Sprintf("\"UNKNOWN_%v\"", param.Name)

			default:
				return false, fmt.Errorf("Meta data type not set or not known: '%v'", param.Type)
			}
		} else {
			switch param.Type {
			case "string":
				param.Default = fmt.Sprintf("\"%v\"", param.Default)
			case "bool":
				boolVal := "false"
				if param.Default.(bool) == true {
					boolVal = "true"
				}
				param.Default = boolVal
			case "[]string":
				param.Default = fmt.Sprintf("[]string{\"%v\"}", strings.Join(param.Default.([]string), "\", \""))
			case "enum":
				param.Default = fmt.Sprintf("\"%v\"", param.Default)
			default:
				return false, fmt.Errorf("Meta data type not set or not known: '%v'", param.Type)
			}
		}

		stepData.Spec.Inputs.Parameters[k] = param
	}
	return osImportRequired, nil
}

func getStepInfo(stepData *config.StepData, osImport bool, exportPrefix string) stepInfo {
	return stepInfo{
		StepName:         stepData.Metadata.Name,
		CobraCmdFuncName: fmt.Sprintf("%vCommand", strings.Title(stepData.Metadata.Name)),
		CreateCmdVar:     fmt.Sprintf("create%vCmd", strings.Title(stepData.Metadata.Name)),
		Short:            stepData.Metadata.Description,
		Long:             stepData.Metadata.LongDescription,
		Metadata:         stepData.Spec.Inputs.Parameters,
		FlagsFunc:        fmt.Sprintf("add%vFlags", strings.Title(stepData.Metadata.Name)),
		OSImport:         osImport,
		ExportPrefix:     exportPrefix,
	}
}

func getEnumInfo(name string, values []string) enumInfo {
	return enumInfo{
		Name:   name,
		Values: values,
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Printf("Error occured: %v\n", err)
		os.Exit(1)
	}
}

// MetadataFiles provides a list of all step metadata files
func MetadataFiles(sourceDirectory string) ([]string, error) {

	var metadataFiles []string

	err := filepath.Walk(sourceDirectory, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".yaml" {
			metadataFiles = append(metadataFiles, path)
		}
		return nil
	})
	if err != nil {
		return metadataFiles, nil
	}
	return metadataFiles, nil
}

func stepTemplate(myStepInfo stepInfo) []byte {

	funcMap := template.FuncMap{
		"flagType":   flagType,
		"golangName": golangName,
		"title":      strings.Title,
		"longName":   longName,
	}

	tmpl, err := template.New("step").Funcs(funcMap).Parse(stepGoTemplate)
	checkError(err)

	var generatedCode bytes.Buffer
	err = tmpl.Execute(&generatedCode, myStepInfo)
	checkError(err)

	return generatedCode.Bytes()
}

func stepTestTemplate(myStepInfo stepInfo) []byte {

	funcMap := template.FuncMap{
		"flagType":   flagType,
		"golangName": golangName,
		"title":      strings.Title,
	}

	tmpl, err := template.New("stepTest").Funcs(funcMap).Parse(stepTestGoTemplate)
	checkError(err)

	var generatedCode bytes.Buffer
	err = tmpl.Execute(&generatedCode, myStepInfo)
	checkError(err)

	return generatedCode.Bytes()
}

func enumTemplate(enumInfo enumInfo) []byte {

	funcMap := template.FuncMap{
		"name":        flagType,
		"golangName":  golangName,
		"title":       strings.Title,
		"toUpperCase": strings.ToUpper,
	}

	tmpl, err := template.New("enum").Funcs(funcMap).Parse(enumGoTemplate)
	checkError(err)

	var generatedCode bytes.Buffer
	err = tmpl.Execute(&generatedCode, enumInfo)
	checkError(err)

	return generatedCode.Bytes()
}

func longName(long string) string {
	l := strings.ReplaceAll(long, "`", "` + \"`\" + `")
	l = strings.TrimSpace(l)
	return l
}

func golangName(name string) string {
	properName := strings.Replace(name, "Api", "API", -1)
	properName = strings.Replace(properName, "api", "API", -1)
	properName = strings.Replace(properName, "Url", "URL", -1)
	properName = strings.Replace(properName, "Id", "ID", -1)
	properName = strings.Replace(properName, "Json", "JSON", -1)
	properName = strings.Replace(properName, "json", "JSON", -1)
	return strings.Title(properName)
}

func flagType(paramType string) string {
	var theFlagType string
	switch paramType {
	case "bool":
		theFlagType = "BoolVar"
	case "string":
		theFlagType = "StringVar"
	case "[]string":
		theFlagType = "StringSliceVar"
	case "enum":
		theFlagType = "StringVar"
	default:
		fmt.Printf("Meta data type not set or not known: '%v'\n", paramType)
		os.Exit(1)
	}
	return theFlagType
}
