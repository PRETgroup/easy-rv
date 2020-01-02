package rvdef

//FinaliseStates will traverse the internal state machine of Policy p.
//It is responsible for setting the "true/false" value of the "Final" flag in all internal states
//(it does not care what they are currently set to, it will clear them all to "false" to begin with)
//States must be set to "Final=true" if no path exists to states of a different accepting kind.
// e.g one:
// s1(accepting) --> s2(nonaccepting) <----> s3(nonaccepting),
// s1 gets final=false, s2 gets final=true, s3 gets final=true,
// e.g. two:
// s4(accepting) --> s5(accepting) <----> s6(nonaccepting),
// s4, s5, s6 get final = false.
func (p *Policy) FinaliseStates() {
	//reset all final flags
	for i := 0; i < len(p.States); i++ {
		p.States[i].FinalStatusType = false
	}

	//for each state, map out its possible destinations using a depth-first search approach
out:
	for i := 0; i < len(p.States); i++ {
		sourceState := p.States[i]
		/* (from Wikipedia DFS)
		// A non-recursive implementation of DFS with worst-case space complexity O(|E|):
		procedure DFS-iterative(G, v) is
		let S be a stack
		S.push(v)
		while S is not empty do
			v = S.pop()
			if v is not labeled as discovered then
				label v as discovered
				for all edges from v to w in G.adjacentEdges(v) do
					S.push(w)
		*/

		S := make([]string, 0, 256) //for now assume 256 elements is enough, Go will add more if necessary
		discoveredNames := make(map[string]bool)
		S = append(S, p.States[i].Name)
		for len(S) > 0 {
			v := S[len(S)-1]
			S = S[:len(S)-1]

			//find the state
			for j := 0; j < len(p.States); j++ {
				if p.States[j].Name == v {
					currentState := p.States[j]
					if currentState.Accepting != sourceState.Accepting {
						continue out //this is not a final status type state, we're done here
					}
				}
			}
			if discoveredNames[v] == false {
				discoveredNames[v] = true

				for _, t := range p.Transitions {
					if t.Source == v {
						if discoveredNames[t.Destination] == false {
							S = append(S, t.Destination)
						}
					}
				}
			}
		}
		//if we couldn't find an escape, this is a Final Status state
		p.States[i].FinalStatusType = true
	}
}
