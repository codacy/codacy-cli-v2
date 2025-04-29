package lizard

// Issue represents a single issue found by Lizard
type Issue struct {
	File        string
	Message     string
	Severity    string
	StartLine   int
	EndLine     int
	RuleID      string
	Description string
}

// LizardMethod represents a method/function in the Lizard results
type LizardMethod struct {
	Name     string `json:"name"`
	FromLine int    `json:"fromLine"`
	ToLine   int    `json:"toLine"`
	File     string `json:"file"`
	Nloc     int    `json:"nloc"`
	Ccn      int    `json:"ccn"`
	Params   int    `json:"params"`
	Tokens   int    `json:"tokens"`
}

// LizardFile represents a file in the Lizard results
type LizardFile struct {
	File          string  `json:"file"`
	Nloc          int     `json:"nloc"`
	MaxCcn        int     `json:"maxCcn"`
	AverageNloc   float64 `json:"averageNloc"`
	AverageCcn    float64 `json:"averageCcn"`
	AverageTokens float64 `json:"averageTokens"`
	MethodsCount  int     `json:"methodsCount"`
}

// LizardResults represents the parsed output from Lizard
type LizardResults struct {
	Methods []LizardMethod `json:"methods"`
	Files   []LizardFile   `json:"files"`
}
