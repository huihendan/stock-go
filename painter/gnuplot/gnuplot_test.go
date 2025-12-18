package gnuplot

import (
	"fmt"
	gnuplot "github.com/sbinet/go-gnuplot"
	"stock-go/utils"
	"testing"
	"time"
)

func Test_paint(t *testing.T) {
	fname := ""
	persist := false
	debug := true

	start1 := time.Now()
	defer utils.CostTime(start1)

	p, err := gnuplot.NewPlotter(fname, persist, debug)
	if err != nil {
		err_string := fmt.Sprintf("** err: %v\n", err)
		panic(err_string)
	}
	defer p.Close()

	p.PlotX([]float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, "some data")
	p.CheckedCmd("set terminal pngcairo")
	p.CheckedCmd("set output 'plot002.png'")
	p.CheckedCmd("replot")

	p.CheckedCmd("q")
	return
}
