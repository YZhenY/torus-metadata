package common

import (
	"fmt"
	"math/big"
	"time"

	logging "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

func Stringify(i interface{}) string {
	bytArr, ok := i.([]byte)
	if ok {
		return string(bytArr)
	}
	str, ok := i.(string)
	if ok {
		return str
	}
	byt, err := bijson.Marshal(i)
	if err != nil {
		logging.WithError(err).Error("Could not bijsonmarshal")
	}
	return string(byt)
}

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s\n", name, elapsed)
}

func BigIntToPoint(x, y *big.Int) Point {
	return Point{X: *x, Y: *y}
}

func HexToBigInt(s string) *big.Int {
	r, ok := new(big.Int).SetString(s, 16)
	if !ok {
		panic("invalid hex in source file: " + s)
	}
	return r
}

type Point struct {
	X big.Int
	Y big.Int
}
