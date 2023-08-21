package tristate

type Tristate *bool

func fromBool(v bool) Tristate {
	return Tristate(&v)
}

var (
	Maybe Tristate = Tristate(nil)
	True  Tristate = fromBool(true)
	False Tristate = fromBool(false)
)
