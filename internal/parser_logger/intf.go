package parser_logger

type ParserLogger interface {
	Error(...any)
	Info(...any)
	Infof(string, ...any)
}
