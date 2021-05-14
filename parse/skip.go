package parse

type Skip interface {
	End(r rune) bool
}

type Double struct {
	c bool
}

func (d *Double) End(r rune) bool {
	if d.c {
		d.c = false
		return false
	}
	if r == '\\' {
		d.c = true
		return false
	}
	return r == '"'
}

type Single struct {
	c bool
}

func (s *Single) End(r rune) bool {
	if s.c {
		s.c = false
		return false
	}
	if r == '\\' {
		s.c = true
		return false
	}
	return r == '\''
}

type SkipHolder struct {
	Skip
}

func (h *SkipHolder) NeedSkip(r rune) bool {
	if h.Skip == nil && r == '"' {
		h.Skip = &Double{}
		return true
	} else if h.Skip == nil && r == '\'' {
		h.Skip = &Single{}
		return true
	}

	if h.Skip != nil {
		end := h.End(r)
		if end {
			h.Skip = nil
		}
		return true
	}
	return false
}
