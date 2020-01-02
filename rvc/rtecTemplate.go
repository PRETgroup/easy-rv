package rvc

import (
	"text/template"

	"github.com/PRETgroup/stcompilerlib"
)

const rvcCTemplate = `{{define "_policyUpd"}}{{$block := .}}
//output policies
{{range $polI, $pol := $block.Policies}}{{$pfbMon := getPolicyMonInfo $block $polI}}
//POLICY {{$pol.Name}} BEGIN
//This will run the monitor for {{$block.Name}}'s policy {{$pol.Name}}
void {{$block.Name}}_run_monitor_{{$pol.Name}}(monitorvars_{{$block.Name}}_t* me, io_{{$block.Name}}_t* io) {
	//advance timers
	{{range $varI, $var := $pfbMon.Policy.GetDTimers}}
	me->{{$var.Name}}++;{{end}}

	//select transition to advance state
	switch(me->_policy_{{$pol.Name}}_state) {
		{{range $sti, $st := $pol.States}}case POLICY_STATE_{{$block.Name}}_{{$pol.Name}}_{{$st.Name}}:
			{{range $tri, $tr := $pfbMon.Policy.Transitions}}{{if eq $tr.Source $st.Name}}{{/*
			*/}}
			if({{$cond := getCECCTransitionCondition $block (compileExpression $tr.STGuard)}}{{$cond.IfCond}}) {
				//transition {{$tr.Source}} -> {{$tr.Destination}} on {{$tr.Condition}}
				me->_policy_{{$pol.Name}}_state = POLICY_STATE_{{$block.Name}}_{{$pol.Name}}_{{$tr.Destination}};
				//set expressions
				{{range $exi, $ex := $tr.Expressions}}
				me->{{$ex.VarName}} = {{$ex.Value}};{{end}}
				break;
			} {{end}}{{end}}
			
			//ensure a transition was taken in this state
			//assert(false && "{{$block.Name}}_{{$pol.Name}}_{{$st.Name}} must take a transition"); //if we are still here, then no transition was taken and we are no longer satisfying liveness

			break;

		{{end}}
	}
}
{{end}}
//OUTPUT POLICY {{/* $pol.Name */}} END
{{end}}

{{define "functionH"}}{{$block := index .Functions .FunctionIndex}}{{$blocks := .Functions}}
//This file should be called F_{{$block.Name}}.h
//This is autogenerated code. Edit by hand at your peril!

#include <stdint.h>
#include <stdbool.h>
#include <assert.h>

//the dtimer_t type
typedef uint64_t dtimer_t;

//For each policy, we need an enum type for the state machine
{{range $polI, $pol := $block.Policies}}
enum {{$block.Name}}_policy_{{$pol.Name}}_states { {{if len $pol.States}}{{range $index, $state := $pol.States}}{{if $index}}, {{end}}
	POLICY_STATE_{{$block.Name}}_{{$pol.Name}}_{{$state.Name}}{{end}}{{else}}POLICY_STATE_{{$block.Name}}_{{$pol.Name}}_unknown{{end}}
};
{{end}}

//IO to the function {{$block.Name}}
typedef struct {
	{{range $index, $var := $block.InterfaceList}}{{$var.Type}} {{$var.Name}}{{if $var.ArraySize}}[{{$var.ArraySize}}]{{end}};
	{{end}}
} io_{{$block.Name}}_t;

//monitor state and vars:
typedef struct {
	{{range $polI, $pol := $block.Policies}}enum {{$block.Name}}_policy_{{$pol.Name}}_states _policy_{{$pol.Name}}_state;
	//internal vars
	{{range $vari, $var := $pol.InternalVars}}{{if not $var.Constant}}{{$var.Type}} {{$var.Name}}{{if $var.ArraySize}}[{{$var.ArraySize}}]{{end}};
	{{end}}{{end}}
	{{end}}
} monitorvars_{{$block.Name}}_t;

{{range $polI, $pol := $block.Policies -}}
{{range $varI, $var := $pol.InternalVars -}}
{{if $var.Constant}}#define CONST_{{$pol.Name}}_{{$var.Name}} {{$var.InitialValue}}{{end}}
{{end}}{{end}}

//This function is provided in "F_{{$block.Name}}.c"
//It sets up the variable structures to their initial values
void {{$block.Name}}_init_all_vars(monitorvars_{{$block.Name}}_t* me, io_{{$block.Name}}_t* io);

//This function is provided in "F_{{$block.Name}}.c"
//It will run the synthesised monitor and call the controller function
void {{$block.Name}}_run_via_monitor(monitorvars_{{$block.Name}}_t* me, io_{{$block.Name}}_t* io);

//This function is provided from the user
//It is the controller function
extern void {{$block.Name}}_run(io_{{$block.Name}}_t* inputs);

//monitor functions

{{range $polI, $pol := $block.Policies}}
//This function is provided in "F_{{$block.Name}}.c"
//It will run the monitor for {{$block.Name}}'s policy {{$pol.Name}}
void {{$block.Name}}_run_monitor_{{$pol.Name}}(monitorvars_{{$block.Name}}_t* me, io_{{$block.Name}}_t* io);

//This function is provided in "F_{{$block.Name}}.c"
//It will check the state of the monitor monitor code
//It returns one of the following:
//0: always true (safe)
//1: currently true (safe)
//2: currently false (unsafe)
//3: always false (unsafe)
uint8_t {{$block.Name}}_check_rv_status_{{$pol.Name}}(monitorvars_{{$block.Name}}_t* me);

{{end}}
{{end}}

{{define "functionC"}}{{$block := index .Functions .FunctionIndex}}{{$blocks := .Functions}}
//This file should be called F_{{$block.Name}}.c
//This is autogenerated code. Edit by hand at your peril!
#include "F_{{$block.Name}}.h"

void {{$block.Name}}_init_all_vars(monitorvars_{{$block.Name}}_t* me, io_{{$block.Name}}_t* io) {
	//set any IO vars with default values
	{{range $index, $var := $block.InterfaceList}}{{if $var.InitialValue}}{{$initialArray := $var.GetInitialArray}}{{if $initialArray}}{{range $initialIndex, $initialValue := $initialArray}}inputs->{{$var.Name}}[{{$initialIndex}}] = {{$initialValue}};
	{{end}}{{else}}inputs->{{$var.Name}} = {{$var.InitialValue}};
	{{end}}{{end}}{{end}}

	{{if $block.Policies}}{{range $polI, $pol := $block.Policies}}
	me->_policy_{{$pol.Name}}_state = {{if $pol.States}}POLICY_STATE_{{$block.Name}}_{{$pol.Name}}_{{(index $pol.States 0).Name}}{{else}}POLICY_STATE_{{$block.Name}}_{{$pol.Name}}_unknown{{end}};
	//input policy internal vars
	{{range $vari, $var := $pol.InternalVars}}{{if not $var.Constant}}
	{{$initialArray := $var.GetInitialArray}}{{if $initialArray}}{{range $initialIndex, $initialValue := $initialArray}}me->{{$var.Name}}[{{$initialIndex}}] = {{$initialValue}};
	{{end}}{{else}}me->{{$var.Name}} = {{if $var.InitialValue}}{{$var.InitialValue}}{{else}}0{{end}};
	{{end}}{{end}}{{end}}
	{{end}}{{end}}
}

void {{$block.Name}}_run_via_monitor(monitorvars_{{$block.Name}}_t* me, io_{{$block.Name}}_t* io) {
	{{$block.Name}}_run(io);

	//run policies in specified order
	{{range $polI, $pol := $block.Policies}}{{$block.Name}}_run_monitor_{{$pol.Name}}(me, io);
	{{end}}
}


{{if $block.Policies}}{{template "_policyUpd" $block}}{{end}}

{{range $polI, $pol := $block.Policies}} {{$pfbMon := getPolicyMonInfo $block $polI}}
//This function is provided in "F_{{$block.Name}}.c"
//It will check the state of the monitor monitor code
//It returns one of the following:
//0: always true (safe)
//1: currently true (safe)
//2: currently false (unsafe)
//3: always false (unsafe)
uint8_t {{$block.Name}}_check_rv_status_{{$pol.Name}}(monitorvars_{{$block.Name}}_t* me) { 
	switch(me->_policy_{{$pol.Name}}_state) { 
		{{range $sti, $st := $pol.States}}case POLICY_STATE_{{$block.Name}}_{{$pol.Name}}_{{$st.Name}}:
			{{if $st.Accepting}} {{if $st.FinalStatusType}} return 0; {{else}} return 1; {{end -}}
			{{else}} {{if $st.FinalStatusType}} return 3; {{else}} return 2; {{end -}}
			{{end}}
		{{end}}
	}
}
{{end}}
{{end}}
{{define "mainCBMCC"}}{{$block := index .Functions .FunctionIndex}}{{$blocks := .Functions}}
//This file should be called cbmc_main_{{$block.Name}}.c
//This is autogenerated code. Edit by hand at your peril!

//It can be used with the cbmc model checker
//Call it using the following command: 
//$ cbmc cbmc_main_{{$block.Name}}.c{{range $blockI, $block := $blocks}} F_{{$block.Name}}.c{{end}}

{{range $blockI, $block := $blocks}}
#include "F_{{$block.Name}}.h"{{end}}
#include <stdio.h>
#include <stdint.h>

int main() {

{{range $blockI, $block := $blocks}}
	//I/O and state for {{$block.Name}}
	monitorvars_{{$block.Name}}_t enf_{{$block.Name}};
    inputs_{{$block.Name}}_t inputs_{{$block.Name}};
    outputs_{{$block.Name}}_t outputs_{{$block.Name}};

	//set values to known state
    {{$block.Name}}_init_all_vars(&enf_{{$block.Name}}, &inputs_{{$block.Name}}, &outputs_{{$block.Name}});

	//introduce nondeterminism
    //a nondet_xxxxx function name tells cbmc that it could be anything, but must be unique
    //randomise inputs
	{{range $inputI, $input := $block.InputVars -}}
	inputs_{{$block.Name}}.{{$input.Name}} = nondet_{{$block.Name}}_input_{{$inputI}}();
	{{end}}

	//randomise monitor state, i.e. clock values and position (excepting violation state)
	{{range $polI, $pol := $block.Policies -}}
	{{range $varI, $var := $pol.InternalVars -}}
	{{if not $var.Constant}}enf_{{$block.Name}}.{{$var.Name}} = nondet_{{$block.Name}}_enf_{{$pol.Name}}_{{$varI}}();{{end}}
	{{end}}
	enf_{{$block.Name}}._policy_{{$pol.Name}}_state = nondet_{{$block.Name}}_enf_{{$pol.Name}}_state() % {{len $pol.States}};
	{{end}}

    //run the monitor (i.e. tell CBMC to check this out)
	{{$block.Name}}_run_via_monitor(&enf_{{$block.Name}}, &inputs_{{$block.Name}}, &outputs_{{$block.Name}});
{{end}}
}

{{range $blockI, $block := $blocks}}
void {{$block.Name}}_run(inputs_{{$block.Name}}_t *inputs, outputs_{{$block.Name}}_t *outputs) {
    //randomise controller

    {{range $outputI, $output := $block.OutputVars -}}
	outputs->{{$output.Name}} = nondet_{{$block.Name}}_output_{{$outputI}}();
	{{end}} 
}
{{end}}
{{end}}
`

var cTemplateFuncMap = template.FuncMap{
	"getCECCTransitionCondition": getCECCTransitionCondition,

	"getPolicyMonInfo": getPolicyMonInfo,

	"compileExpression": stcompilerlib.CCompileExpression,

	"sub": sub,
}

var cTemplates = template.Must(template.New("").Funcs(cTemplateFuncMap).Parse(rvcCTemplate))
