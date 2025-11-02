package fixtures

type ValidModel struct {
	Name       string   `dagjsongen:"name"`
	Invocation []byte   `dagjsongen:"invocation"`
	Proofs     [][]byte `dagjsongen:"proofs"`
}

type ErrorModel struct {
	Name string `dagjsongen:"name"`
}

type InvalidModel struct {
	Name       string     `dagjsongen:"name"`
	Invocation []byte     `dagjsongen:"invocation"`
	Proofs     [][]byte   `dagjsongen:"proofs"`
	Error      ErrorModel `dagjsongen:"error"`
}

type FixturesModel struct {
	Version  string         `dagjsongen:"version"`
	Comments string         `dagjsongen:"comments"`
	Valid    []ValidModel   `dagjsongen:"valid"`
	Invalid  []InvalidModel `dagjsongen:"invalid"`
}
