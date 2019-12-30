package rvdef

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/PRETgroup/stcompilerlib"
)

//FBECCGuardToSTExpression converts a given FB's guard into a STExpression parsetree
func FBECCGuardToSTExpression(pName, guard string) ([]stcompilerlib.STInstruction, *stcompilerlib.STParseError) {
	return stcompilerlib.ParseString(pName, guard)
}

//PSTTransition is a container struct for a PTransition and its ST translated guard
type PSTTransition struct {
	PTransition
	STGuard stcompilerlib.STExpression
}

//A PMonitorPolicy is what goes inside a PMonitor, it is derived from a Policy
type PMonitorPolicy struct {
	InternalVars []Variable
	States       []PState
	Transitions  []PSTTransition
}

//GetDTimers returns all DTIMERS in a PMonitorPolicy
func (pol PMonitorPolicy) GetDTimers() []Variable {
	dTimers := make([]Variable, 0)
	for _, v := range pol.InternalVars {
		if strings.ToLower(v.Type) == "dtimer_t" {
			dTimers = append(dTimers, v)
		}
	}
	return dTimers
}

//GetTransitionsForSource returns a slice of all transitions in this PMonitorPolicy
//which have a source as "src"
func (pol PMonitorPolicy) GetTransitionsForSource(src string) []PSTTransition {
	nTrans := make([]PSTTransition, 0)
	for _, tr := range pol.Transitions {
		if tr.Source == src {
			nTrans = append(nTrans, tr)
		}
	}
	return nTrans
}

//DoesExpressionInvolveTime returns true if a given expression uses time
func (pol PMonitorPolicy) DoesExpressionInvolveTime(expr stcompilerlib.STExpression) bool {
	op := expr.HasOperator()
	if op == nil {
		return VariablesContain(pol.GetDTimers(), expr.HasValue())
	}
	for _, arg := range expr.GetArguments() {
		if pol.DoesExpressionInvolveTime(arg) {
			return true
		}
	}
	return false
}

//A PMonitor will store a given input and output policy and can derive the enforcers required to uphold them
type PMonitor struct {
	interfaceList InterfaceList
	Name          string
	Policy        PMonitorPolicy
}

//MakePMonitor will convert a given policy to an enforcer for that policy
func MakePMonitor(il InterfaceList, p Policy) (*PMonitor, error) {
	//make the enforcer
	enf := &PMonitor{interfaceList: il, Name: p.Name}
	//first, convert policy transitions
	outpTr, err := p.GetPSTTransitions()
	if err != nil {
		return nil, err
	}

	enf.Policy = PMonitorPolicy{
		InternalVars: p.InternalVars,
		States:       p.States,
		Transitions:  outpTr,
	}

	enf.Policy.RemoveNilTransitions()
	enf.Policy.RemoveDuplicateTransitions()

	return enf, nil
}

//RemoveNilTransitions will do a search through a policies transitions and remove any that have nil guards
func (pol *PMonitorPolicy) RemoveNilTransitions() {
	for i := 0; i < len(pol.Transitions); i++ {
		for j := i; j < len(pol.Transitions); j++ {
			if pol.Transitions[j].STGuard == nil || pol.Transitions[j].Condition == "" || stcompilerlib.STCompileExpression(pol.Transitions[j].STGuard) == "" {
				pol.Transitions = append(pol.Transitions[:j], pol.Transitions[j+1:]...)
				j--
			}
		}
	}
}

//RemoveDuplicateTransitions will do a search through a policies transitions and remove any that are simple duplicates
//(i.e. every field the same and in the same order).
func (pol *PMonitorPolicy) RemoveDuplicateTransitions() {
	for i := 0; i < len(pol.Transitions); i++ {
		for j := i + 1; j < len(pol.Transitions); j++ {
			if reflect.DeepEqual(pol.Transitions[i], pol.Transitions[j]) {
				pol.Transitions = append(pol.Transitions[:j], pol.Transitions[j+1:]...)
				j--
			}
		}
	}
}

//CURRENTLY UNUSED
//RemoveAlwaysTrueTransitions will do a search through a policies transitions and remove any that are just "true"
func (pol *PMonitorPolicy) RemoveAlwaysTrueTransitions() {
	for i := 0; i < len(pol.Transitions); i++ {
		for j := i; j < len(pol.Transitions); j++ {
			if val := pol.Transitions[j].STGuard.HasValue(); val == "true" || val == "1" {
				pol.Transitions = append(pol.Transitions[:j], pol.Transitions[j+1:]...)
				j--
			}
		}
	}
}

//VariablesContain returns true if a list of variables contains a given name
func VariablesContain(vars []Variable, name string) bool {
	for i := 0; i < len(vars); i++ {
		if vars[i].Name == name {
			return true
		}
	}
	return false
}

func stringSliceContains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

//GetPSTTransitions will convert all internal PTransitions into PSTTransitions (i.e. PTransitions with a ST symbolic tree condition)
func (p *Policy) GetPSTTransitions() ([]PSTTransition, error) {
	stTrans := make([]PSTTransition, len(p.Transitions))
	for i := 0; i < len(p.Transitions); i++ {
		stguard, err := FBECCGuardToSTExpression(p.Name, p.Transitions[i].Condition)
		if err != nil {
			return nil, err
		}
		if len(stguard) != 1 {
			return nil, fmt.Errorf("Incompatible policy guard (wrong number of expressions)")
		}
		expr, ok := stguard[0].(stcompilerlib.STExpression)
		if !ok {
			return nil, fmt.Errorf("Incompatible policy guard (not an expression)")
		}
		stTrans[i] = PSTTransition{
			PTransition: p.Transitions[i],
			STGuard:     expr,
		}
	}
	return stTrans, nil
}

//SplitExpressionsOnOr will take a given STExpression and return a slice of STExpressions which are
//split over the "or" operators, e.g.
//[a] should become [a]
//[or a b] should become [a] [b]
//[or a [b and c]] should become [a] [b and c]
//[[a or b] and [c or d]] should become [a and c] [a and d] [b and c] [b and d]
func SplitExpressionsOnOr(expr stcompilerlib.STExpression) []stcompilerlib.STExpression {
	//IF IS OR
	//	BREAK APART
	//IF IS VALUE
	//	RETURN CURRENT
	//IF IS OTHER OPERATOR
	//	MARK LOCATION AND RECURSE

	// broken := breakIfOr(expr)
	// if len(broken) == 1 {
	// 	return broken
	// }

	op := expr.HasOperator()
	if op == nil { //if it's just a value, return
		return []stcompilerlib.STExpression{expr}
	}
	if op.GetToken() == "or" { //if it's an "or", return the arguments
		rets := make([]stcompilerlib.STExpression, 0)
		args := expr.GetArguments()
		for i := 0; i < len(args); i++ { //for each argument of the "or", return it, unless it is itself an "or" (in which case, expand further)
			arg := args[i]
			argOp := arg.HasOperator()
			if argOp == nil || argOp.GetToken() != "or" {
				rets = append(rets, arg)
				continue
			}
			args = append(args, arg.GetArguments()...)
		}
		return rets
	}

	//otherwise, things are more interesting

	//make the thing we're returning
	rets := make([]stcompilerlib.STExpressionOperator, 0)

	//build a new expression
	var nExpr stcompilerlib.STExpressionOperator

	//operator is op, arguments are args
	nExpr.Operator = op
	args := expr.GetArguments()
	nExpr.Arguments = make([]stcompilerlib.STExpression, len(args))

	rets = append(rets, nExpr)
	//for each argument in the expression operator
	for i, arg := range args {
		//get arguments to operator by calling SplitExpressionsOnOr again
		argT := SplitExpressionsOnOr(arg)
		//if argT has more than one value, it indicates that this argument was "split", and we should return two nExpr, one with each argument
		//we will increase the size of rets by a multiplyFactor, which is the size of argT
		//i.e. if we receive two arguments, and we already had two elements in rets, it indicates we need to return 4 values
		//for instance, if our original command was "(a or b) and (c or d)" we'd need to return 4 elements (a and c) (a and d) (b and c) (b and d)
		multiplyFactor := len(argT)
		//for each factor in multiplyFactor, duplicate rets[n]
		//e.g. multiplyFactor 2 on [1 2 3] becomes [1 1 2 2 3 3]
		//e.g. multiplyFactor 3 on [1 2 3] becomes [1 1 1 2 2 2 3 3 3]
		for y := 0; y < len(rets); y++ {
			for z := 1; z < multiplyFactor; z++ {

				var newElem stcompilerlib.STExpressionOperator
				copyElem := rets[y]
				newElem.Operator = copyElem.Operator
				newElem.Arguments = make([]stcompilerlib.STExpression, len(copyElem.Arguments))
				copy(newElem.Arguments, copyElem.Arguments)

				rets = append(rets, stcompilerlib.STExpressionOperator{})
				copy(rets[y+1:], rets[y:])
				rets[y] = newElem
				y++
			}
		}

		//for each argument, copy it into the return elements at the appropriate locations
		//(if we have multiple arguments, they will be chosen in a round-robin fashion)
		for j := 0; j < len(argT); j++ {
			at := argT[j]
			for k := j; k < len(rets); k += len(argT) {
				rets[k].Arguments[i] = at
			}
		}

		//expected, _ := json.MarshalIndent(rets, "\t", "\t")
		//fmt.Printf("Current:\n\t%s\n\n", expected)
	}

	//conversion for returning
	actualRets := make([]stcompilerlib.STExpression, len(rets))
	for i := 0; i < len(rets); i++ {
		actualRets[i] = rets[i]
	}
	return actualRets

}

//DeepGetValues recursively gets all values from a given stcompilerlib.STExpression
func DeepGetValues(expr stcompilerlib.STExpression) []string {
	if expr == nil {
		return nil
	}
	if val := expr.HasValue(); val != "" {
		return []string{val}
	}
	vals := make([]string, 0)
	for _, arg := range expr.GetArguments() {
		vals = append(vals, DeepGetValues(arg)...)
	}
	return vals
}
