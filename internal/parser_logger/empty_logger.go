package parser_logger

func NewEmptyLogger() *_logger {
	return &_logger{}
}

type _logger struct{}

func (_ _logger) Error(a ...any) {}

func (_ _logger) Info(a ...any) {}

func (_ _logger) Infof(s string, a ...any) {}
