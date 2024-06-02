package utils

func PanicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func Assert(expr bool, err ...error) {
	if !expr {
		if len(err) > 0 {
			panic(err[0])
		}
		panic("assertion failed")
	}
}
