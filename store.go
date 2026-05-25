package validate

import (
	"fmt"
	"sync"
)

type store struct {
	list map[Name]Rule
	mux  sync.RWMutex
}

func newStore() *store {
	return &store{
		list: make(map[Name]Rule, 10),
	}
}

func (s *store) Append(rules ...Rule) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	for _, rule := range rules {
		if err := rule.Validate(); err != nil {
			return err
		}

		if _, ok := s.list[rule.Name]; ok {
			return fmt.Errorf("duplicate rule: %s", rule.Name)
		}

		s.list[rule.Name] = rule
	}

	return nil
}

func (s *store) Resolve(name Name) (Rule, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	rule, ok := s.list[name]
	if !ok {
		return Rule{}, false
	}
	return rule, true
}
