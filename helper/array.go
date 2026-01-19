package helper

type StringSet struct {
	items map[string]struct{}
}

func NewStringSet() *StringSet {
	return &StringSet{
		items: make(map[string]struct{}),
	}
}

func NewStringSetFromSlice(strings []string) *StringSet {
	s := &StringSet{
		items: make(map[string]struct{}, len(strings)),
	}
	for _, str := range strings {
		s.items[str] = struct{}{}
	}
	return s
}

func (s *StringSet) Add(str string) {
	s.items[str] = struct{}{}
}

func (s *StringSet) Exists(str string) bool {
	_, exists := s.items[str]
	return exists
}

func (s *StringSet) Size() int {
	return len(s.items)
}
