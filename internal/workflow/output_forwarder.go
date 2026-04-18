package workflow

import "ralph/internal/runner"

type OutputForwarder struct {
	emit func(Event)
}

func NewOutputForwarder(emit func(Event)) *OutputForwarder {
	return &OutputForwarder{emit: emit}
}

func (f *OutputForwarder) Forward(outputCh <-chan runner.OutputLine) {
	for line := range outputCh {
		f.emit(EventOutput{Output: Output{Text: line.Text, IsErr: line.IsErr, Verbose: line.Verbose}})
	}
}
