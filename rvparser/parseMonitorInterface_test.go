package rvparser

import (
	"testing"

	"github.com/PRETgroup/easy-rv/rvdef"
)

var interfaceTests = []ParseTest{

	{
		Name: "events typo 1",
		Input: `monitor testBlock;
					interface of testBlock {
						bool inEvent;
						outEvent;
					}`,
		Err: ErrInvalidType,
	},
	{
		Name: "events typo 2",
		Input: `monitor testBlock;
					interface of testBlock {
						bool inEvent;
						bool outEvent
					}`,
		Err: ErrUnexpectedValue,
	},
	{
		Name: "data typo 1",
		Input: `monitor testBlock;
					interface of testBlock {
						bool inEvent;
						asdasd inData;
						bool outEvent;
					}`,
		Err: ErrInvalidType,
	},
	{
		Name: "data input 1",
		Input: `monitor testBlock;
					interface of testBlock {
						bool inEvent;
						bool outEvent;
					}`,
		Output: []rvdef.Monitor{
			rvdef.Monitor{
				Name: "testBlock",
				InterfaceList: []rvdef.Variable{
					rvdef.Variable{Name: "inEvent", Type: "bool", ArraySize: "", InitialValue: "", Comment: ""},
					rvdef.Variable{Name: "outEvent", Type: "bool", ArraySize: "", InitialValue: "", Comment: ""},
				},
				Policies: []rvdef.Policy(nil)},
		},
		Err: nil,
	},
	{
		Name: "data input 2",
		Input: `monitor testBlock;
					interface of testBlock {
						bool inEvent;
						bool[3] inData;
						bool outEvent;
					}`,
		Output: []rvdef.Monitor{
			rvdef.Monitor{
				Name: "testBlock",
				InterfaceList: []rvdef.Variable{
					rvdef.Variable{Name: "inEvent", Type: "bool", ArraySize: "", InitialValue: "", Comment: ""},
					rvdef.Variable{Name: "inData", Type: "bool", ArraySize: "3", InitialValue: "", Comment: ""},
					rvdef.Variable{Name: "outEvent", Type: "bool", ArraySize: "", InitialValue: "", Comment: ""},
				},
				Policies: []rvdef.Policy(nil),
			},
		},
		Err: nil,
	},
	{
		Name: "data input array typo 1",
		Input: `basicFB testBlock;
					interface of testBlock {
						event inEvent;
						bool[3 inData;
						event outEvent;
					}`,
		Err: ErrUnexpectedValue,
	},
	{
		Name: "data input 3",
		Input: `monitor testBlock;
					interface of testBlock {
						int8_t inEvent;
						bool[3] inData := [0,1,0];
						char outEvent;
					}`,
		Output: []rvdef.Monitor{
			rvdef.Monitor{
				Name: "testBlock",
				InterfaceList: rvdef.InterfaceList{
					rvdef.Variable{Name: "inEvent", Type: "int8_t", ArraySize: "", InitialValue: "", Comment: ""},
					rvdef.Variable{Name: "inData", Type: "bool", ArraySize: "3", InitialValue: "[0,1,0]", Comment: ""},
					rvdef.Variable{Name: "outEvent", Type: "char", ArraySize: "", InitialValue: "", Comment: ""},
				},
				Policies: []rvdef.Policy(nil)}},
		Err: nil,
	},
	{
		Name: "data default typo 1",
		Input: `basicFB testBlock;
					interface of testBlock {
						bool inEvent;
						bool[3] inData := 0,1,0;
						bool outEvent;
					}`,
		Err: ErrUnexpectedValue,
	},
	{
		Name: "Unexpected EOF",
		Input: `basicFB testBlock;
					interface of testBlock {
						bool inEvent;
						bool inEvent2;
						int32_t `,
		Err: ErrUnexpectedValue,
	},
}

func TestParseStringInterface(t *testing.T) {
	runParseTests(t, interfaceTests)
}
