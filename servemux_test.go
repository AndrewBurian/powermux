package powermux

import "testing"

func TestServeMux_String(t *testing.T) {
	s := NewServeMux()
	s.Route("/a/b/c").Get(nil)
	s.Route("/d/e/f").Get(nil).Post(nil)
	s.Route("/a/b").Delete(nil).NotFound(nil)

	//str := s.String()
	//t.Error("\n"+str)
}