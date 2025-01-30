package parser

//go:generate mockgen -source=parser.go -destination=_mock/parser.go
type CSVParser interface {
	ParseLine(record []string) (interface{}, error)
}

var ParserRegistry = map[string]CSVParser{}

func RegisterParser(parserID string, parser CSVParser) {
	ParserRegistry[parserID] = parser
}

func GetParser(parserID string) CSVParser {
	return ParserRegistry[parserID]
}
