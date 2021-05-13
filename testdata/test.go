package testdata

type x struct {
}

func (x) s() {

}

func (x x) ss() {

}

func s(
	f1,
	f2 func(some string) (m string),
	m map[string]int,
	x *x,
	q x,
	s []string,
	v interface{},
	n [1]string) {

}

func (x *x) sss() {

}
