//args: -Ewastedassign
package testdata

import (
	"strings"
)

func p(x int) int {
	return x + 1
}

func multiple(val interface{}, times uint) interface{} {

	switch hogehoge := val.(type) {
	case int:
		return 12
	case string:
		return strings.Repeat(hogehoge, int(times))
	default:
		return nil
	}
}

func noUseParams(params string) int {
	a := 12
	println(a)
	return a
}

func f(param int) int {
	println(param)
	useOutOfIf := 1212121 // ERROR "wasted assignment"
	ret := 0
	if false {
		useOutOfIf = 200 // ERROR "reassigned, but never used afterwards"
		return 0
	} else if param == 100 {
		useOutOfIf = 100 // ERROR "wasted assignment"
		useOutOfIf = 201
		useOutOfIf = p(useOutOfIf)
		useOutOfIf += 200 // ERROR "wasted assignment"
	} else {
		useOutOfIf = 100
		useOutOfIf += 100
		useOutOfIf = p(useOutOfIf)
		useOutOfIf += 200 // ERROR "wasted assignment"
	}

	if false {
		useOutOfIf = 200 // ERROR "reassigned, but never used afterwards"
		return 0
	} else if param == 200 {
		useOutOfIf = 100 // ERROR "wasted assignment"
		useOutOfIf = 201
		useOutOfIf = p(useOutOfIf)
		useOutOfIf += 200
	} else {
		useOutOfIf = 100
		useOutOfIf += 100
		useOutOfIf = p(useOutOfIf)
		useOutOfIf += 200
	}
	println(useOutOfIf)
	useOutOfIf = 192
	useOutOfIf += 100
	useOutOfIf += 200 // ERROR "reassigned, but never used afterwards"
	return ret
}

func checkLoopTest() int {
	hoge := 12
	noUse := 1111
	println(noUse)

	noUse = 1111 // ERROR "reassigned, but never used afterwards"
	for {
		if hoge == 14 {
			break
		}
		hoge = hoge + 1
	}
	return hoge
}

func r(param int) int {
	println(param)
	useOutOfIf := 1212121
	ret := 0
	if false {
		useOutOfIf = 200 // ERROR "reassigned, but never used afterwards"
		return 0
	} else if param == 100 {
		ret = useOutOfIf
	} else if param == 200 {
		useOutOfIf = 100 // ERROR "wasted assignment"
		useOutOfIf = 100
		useOutOfIf = p(useOutOfIf)
		useOutOfIf += 200 // ERROR "wasted assignment"
	}
	useOutOfIf = 12
	println(useOutOfIf)
	useOutOfIf = 192
	useOutOfIf += 100
	useOutOfIf += 200 // ERROR "reassigned, but never used afterwards"
	return ret
}

func mugen() {
	var i int
	var hoge int
	for {
		hoge = 5 // ERROR "reassigned, but never used afterwards"
	}

	println(i)
	println(hoge)
	return
}

func noMugen() {
	var i int
	var hoge int
	for {
		hoge = 5
		break
	}

	println(i)
	println(hoge)
	return
}
