package fake

import "fmt"

type Storage struct {
	getHandler func(key string) ([]byte, error)
	putHandler func(key string, body []byte) error
	data map[string][]byte
}

func (s *Storage) Get(key string) ([]byte, error) {
	if s.data == nil {
		s.data = map[string][]byte{}
	}
	b, ok := s.data[key]
	if !ok {
		return nil, fmt.Errorf("not found key: %s", key)
	}
	return b, nil
}

func (s *Storage) Put(key string, body []byte) error {
	if s.data == nil {
		s.data = map[string][]byte{}
	}
	s.data[key] = body
	return nil
}

