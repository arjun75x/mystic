// taken from https://www.davidkaya.com/sets-in-golang/
package main

var exists = struct{}{}

type set struct {
	m map[string]struct{}
}

func NewSet() *set {
	s := &set{}
	s.m = make(map[string]struct{})
	return s
}

func (s *set) Add(value string) {
	s.m[value] = exists
}

func (s *set) Remove(value string) {
	delete(s.m, value)
}

func (s *set) Contains(value string) bool {
	_, c := s.m[value]
	return c
}

func (s *set) List() []string {
	keys := make([]string, len(s.m))

	i := 0
	for k := range s.m {
		keys[i] = k
		i++
	}

	return keys
}
