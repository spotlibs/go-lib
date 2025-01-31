package log

type noop struct{}

func (noop) Warning(_ Map) {}
func (noop) Error(_ Map)   {}
func (noop) Info(_ Map)    {}
