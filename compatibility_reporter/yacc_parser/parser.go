package yacc_parser

type Seq struct {
	Items []string
}

type Production struct {
	Head  string
	Alter []Seq
}

type stateType int

const (
	initState            = 0
	delimFetchedState    = 1
	termFetchedState     = 2
	prepareNextProdState = 3
	endState             = 4
)

func Parse(nextToken func() token) []Production {
	var tkn token
	var prods []Production
	var p Production
	var s Seq
	var lastTerm string

	state := initState
	p.Head = nextToken().toString()

	for state != endState {
		tkn = nextToken()
		switch state {
		case initState:
			if tkn.toString() != ":" {
				panic("expect ':'")
			}
			state = delimFetchedState
		case delimFetchedState:
			_, isNt := tkn.(*nonTerminal)
			_, isKw := tkn.(*keyword)
			if !isNt && !isKw {
				panic(tkn.toString() + " is not keyword or nonterminal")
			}
			state = termFetchedState
			s.Items = append(s.Items, tkn.toString())
		case termFetchedState:
			switch v := tkn.(type) {
			case *eof:
				p.Alter = append(p.Alter, s)
				prods = append(prods, p)
				state = endState
			case *operator:
				p.Alter = append(p.Alter, s)
				s = Seq{}
				state = termFetchedState
			case *nonTerminal, *keyword:
				lastTerm = v.toString()
				state = prepareNextProdState
			}
		case prepareNextProdState:
			switch v := tkn.(type) {
			case *eof:
				s.Items = append(s.Items, lastTerm)
				p.Alter = append(p.Alter, s)
				prods = append(prods, p)
				state = endState
			case *operator:
				if v.val == "|" {
					s.Items = append(s.Items, lastTerm)
					p.Alter = append(p.Alter, s)
					s = Seq{}
				} else if v.val == ":" {
					p.Alter = append(p.Alter, s)
					s = Seq{}
					prods = append(prods, p)
					p = Production{Head: lastTerm}
				}
				state = delimFetchedState
			case *nonTerminal, *keyword:
				s.Items = append(s.Items, lastTerm)
				lastTerm = v.toString()
			}
		}
	}
	return prods
}
