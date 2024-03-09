package main

import "fmt"

// Server struct
type Server struct {
	IP   string
	Port int
}

type OptionsServerFunc func(c *Server) error

// NewServer function
func NewServer(opts ...OptionsServerFunc) (*Server, error) {
	s := &Server{}
	for _, opt := range opts {
		if err := opt(s); err != nil { // call the given functions, which will modify the server
			return nil, err
		}
	}
	return s, nil // return the modified server
}

// WithIP function modifies the server's IP
func WithIP(ip string) OptionsServerFunc {
	return func(s *Server) error {
		s.IP = ip
		return nil
	}
}

func main() {
	s1, _ := NewServer(WithIP("127.0.0.1"))
	fmt.Printf("%+v\n", s1)
}
