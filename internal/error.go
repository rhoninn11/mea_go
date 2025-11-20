package internal

import (
	"fmt"
	"log"
)

type errorPanicker struct {
	treshold int
	count    int
}

func Panicker(treshold int) errorPanicker {
	return errorPanicker{
		treshold: treshold,
		count:    0,
	}
}

func (ep *errorPanicker) HasError(err error) bool {
	occured := err != nil
	if occured {
		log.Default().Fatalf("ERROR: %s\n", err.Error())
		ep.count += 1
	}

	if ep.count >= ep.treshold {
		panicMsg := fmt.Sprintf("PANIC: reached error treshold (%d)", ep.treshold)
		panic(panicMsg)
	}
	return occured
}
