package rvc

import (
	"bytes"
	"encoding/xml"
	"errors"
	"strings"
	"text/template"

	"github.com/PRETgroup/easy-rv/rvdef"
)

//Converter is the struct we use to store all functions for conversion (and what we operate from)
type Converter struct {
	Funcs     []rvdef.Monitor
	Language  string
	templates *template.Template
}

//New returns a new instance of a Converter based on the provided language
func New(language string) (*Converter, error) {
	switch strings.ToLower(language) {
	case "c":
		return &Converter{Funcs: make([]rvdef.Monitor, 0), Language: "c", templates: cTemplates}, nil
		//	case "verilog":
		//		return &Converter{Funcs: make([]rvdef.Monitor, 0), Language: "verilog", templates: verilogTemplates}, nil
	default:
		return nil, errors.New("Language " + language + " is not supported")
	}
}

//AddFunction should be called for each Function in the project
func (c *Converter) AddFunction(functionbytes []byte) error {
	FB := rvdef.Monitor{}
	if err := xml.Unmarshal(functionbytes, &FB); err != nil {
		return errors.New("Couldn't unmarshal Monitor xml: " + err.Error())
	}

	c.Funcs = append(c.Funcs, FB)

	return nil
}

//OutputFile is used when returning the converted data from the iec61499
type OutputFile struct {
	Name      string
	Extension string
	Contents  []byte
}

//TemplateData is the structure used to hold data being passed into the templating engine
type TemplateData struct {
	FunctionIndex int
	Functions     []rvdef.Monitor
}

//ConvertAll converts iec61499 xml (stored as []FB) into vhdl []byte for each block (becomes []VHDLOutput struct)
//Returns nil error on success
func (c *Converter) ConvertAll() ([]OutputFile, error) {

	//first, finalise the states
	for i := 0; i < len(c.Funcs); i++ {
		for j := 0; j < len(c.Funcs[i].Policies); j++ {
			c.Funcs[i].Policies[j].FinaliseStates()
		}
	}

	finishedConversions := make([]OutputFile, 0, len(c.Funcs))

	type templateInfo struct {
		Prefix    string
		Name      string
		Extension string
	}

	var templates []templateInfo

	//convert all functions
	if c.Language == "c" {
		templates = []templateInfo{
			{"F_", "functionC", "c"},
			{"F_", "functionH", "h"},
			//{"cbmc_main_", "mainCBMCC", "c"},
		}
	}
	// if c.Language == "verilog" {
	// 	templates = []templateInfo{
	// 		{"test_F_", "functionVerilog", "sv"},
	// 	}
	// }
	for _, template := range templates {
		for i := 0; i < len(c.Funcs); i++ {

			output := &bytes.Buffer{}
			if err := c.templates.ExecuteTemplate(output, template.Name, TemplateData{FunctionIndex: i, Functions: c.Funcs}); err != nil {
				return nil, errors.New("Couldn't format template (fb) of" + c.Funcs[i].Name + ": " + err.Error())
			}

			finishedConversions = append(finishedConversions, OutputFile{Name: template.Prefix + c.Funcs[i].Name, Extension: template.Extension, Contents: output.Bytes()})
		}
	}

	return finishedConversions, nil
}
