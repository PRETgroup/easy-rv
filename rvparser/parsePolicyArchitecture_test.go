package rvparser

import (
	"testing"

	"github.com/PRETgroup/easy-rv/rvdef"
)

var efbArchitectureTests = []ParseTest{
	{
		Name: "missing brace after s1",
		Input: `monitor testBlock;
				interface of testBlock{
				}
				policy of testBlock {
					states {
						s1 

					}
				}`,
		Err: ErrUnexpectedValue,
	},
	{
		Name: "AEIPolicy",
		Input: `monitor AEIPolicy;
				interface of AEIPolicy {
					bool AS, VS; //in here means that they're going from PLANT to CONTROLLER
					bool AP, VP;//out here means that they're going from CONTROLLER to PLANT
				
					uint64_t AEI_ns := 900000000;
				}
				policy AEI of AEIPolicy {
					internals {
						dtimer_t tAEI; //DTIMER increases in DISCRETE TIME continuously
					}
				
					//P3: AS or AP must be true within AEI after a ventricular event VS or VP.
				
					states {
						s1 accepting {
							//-> <destination> [on guard] [: output expression][, output expression...] ;
							-> s2 on (VS or VP): tAEI := 0;
						}
				
						s2 rejecting {
							-> s1 on (AS or AP);
							-> violation on (tAEI > AEI_ns);
						}

						violation rejecting trap;
					} 
				}`,
		Output: []rvdef.Monitor{
			rvdef.Monitor{
				Name: "AEIPolicy",
				InterfaceList: []rvdef.Variable{
					rvdef.Variable{Name: "AS", Type: "bool", ArraySize: "", InitialValue: "", Comment: ""},
					rvdef.Variable{Name: "VS", Type: "bool", ArraySize: "", InitialValue: "", Comment: ""},
					rvdef.Variable{Name: "AP", Type: "bool", ArraySize: "", InitialValue: "", Comment: ""},
					rvdef.Variable{Name: "VP", Type: "bool", ArraySize: "", InitialValue: "", Comment: ""},
					rvdef.Variable{Name: "AEI_ns", Type: "uint64_t", ArraySize: "", InitialValue: "900000000", Comment: ""},
				},
				Policies: []rvdef.Policy{
					rvdef.Policy{
						Name: "AEI",
						InternalVars: []rvdef.Variable{
							rvdef.Variable{Name: "tAEI", Type: "dtimer_t", ArraySize: "", InitialValue: "", Comment: ""},
						},
						States: []rvdef.PState{{Name: "s1", Accepting: true}, {Name: "s2", Accepting: false}, {Name: "violation", Accepting: false}},
						Transitions: []rvdef.PTransition{
							rvdef.PTransition{Source: "s1", Destination: "s2", Condition: "( VS or VP )", Expressions: []rvdef.PExpression{rvdef.PExpression{VarName: "tAEI", Value: "0"}}},
							rvdef.PTransition{Source: "s2", Destination: "s1", Condition: "( AS or AP )", Expressions: []rvdef.PExpression(nil)},
							rvdef.PTransition{Source: "s2", Destination: "violation", Condition: "( tAEI > AEI_ns )", Expressions: []rvdef.PExpression(nil)},
						},
					},
				},
			},
		},
	},
	{
		Name: "AB5Policy",
		Input: `monitor ab5;
		interface of ab5 {
			bool A;  
			bool B; 
		}
		
		policy AB5 of ab5 {
			internals {
				dtimer_t v;
			}
		
			states {
		
				//first state is initial, and represents "We're waiting for an A"
				s0 accepting {
					//if we receive neither A nor B, do nothing														
					-> s0 on (!A and !B): v := 0;
		
					//if we receive an A only, head to state s1							
					-> s1 on (A and !B): v := 0;
					
					//if we receive a !A and B then VIOLATION							
					-> violation on (!A and B);
		
					//if we receive both A and B then done
					-> done on (A and B);	
				}
		
				//s1 is "we're waiting for a B, and it needs to get here within 5 ticks"
				s1 rejecting {
					//if we receive nothing, and we aren't over-time, then we do nothing														
					-> s1 on (!A and !B and v < 5);	
		
					//if we receive a B only, head to state s0					
					-> s0 on (!A and B);
		
					//if we go overtime, or we receive another A, then VIOLATION	
					-> violation on ((v >= 5) or (A and B) or (A and !B));
				}
		
				done accepting trap;
		
				violation rejecting trap;
			}
		}`,
		Output: []rvdef.Monitor{
			rvdef.Monitor{
				Name: "ab5",
				InterfaceList: []rvdef.Variable{
					rvdef.Variable{Name: "A", Type: "bool", ArraySize: "", InitialValue: "", Comment: ""},
					rvdef.Variable{Name: "B", Type: "bool", ArraySize: "", InitialValue: "", Comment: ""},
				},
				Policies: []rvdef.Policy{
					rvdef.Policy{
						Name: "AB5",
						InternalVars: []rvdef.Variable{
							rvdef.Variable{Name: "v", Type: "dtimer_t", ArraySize: "", InitialValue: "", Comment: ""},
						},
						States: []rvdef.PState{
							rvdef.PState{Name: "s0", Accepting: true},
							rvdef.PState{Name: "s1", Accepting: false},
							rvdef.PState{Name: "done", Accepting: true},
							rvdef.PState{Name: "violation", Accepting: false},
						},
						Transitions: []rvdef.PTransition{
							rvdef.PTransition{Source: "s0", Destination: "s0", Condition: "( !A and !B )", Expressions: []rvdef.PExpression{rvdef.PExpression{VarName: "v", Value: "0"}}},
							rvdef.PTransition{Source: "s0", Destination: "s1", Condition: "( A and !B )", Expressions: []rvdef.PExpression{rvdef.PExpression{VarName: "v", Value: "0"}}},
							rvdef.PTransition{Source: "s0", Destination: "violation", Condition: "( !A and B )", Expressions: []rvdef.PExpression(nil)},
							rvdef.PTransition{Source: "s0", Destination: "done", Condition: "( A and B )", Expressions: []rvdef.PExpression(nil)},
							rvdef.PTransition{Source: "s1", Destination: "s1", Condition: "( !A and !B and v < 5 )", Expressions: []rvdef.PExpression(nil)},
							rvdef.PTransition{Source: "s1", Destination: "s0", Condition: "( !A and B )", Expressions: []rvdef.PExpression(nil)},
							rvdef.PTransition{Source: "s1", Destination: "violation", Condition: "( ( v >= 5 ) or ( A and B ) or ( A and !B ) )", Expressions: []rvdef.PExpression(nil)},
						},
					},
				},
			},
		},
	},
}

func TestParsePFBArchitecture(t *testing.T) {
	runParseTests(t, efbArchitectureTests)
}
