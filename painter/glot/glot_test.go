package glot

import (
	"fmt"
	"github.com/Arafatk/glot"
	"testing"
)

func Test_line(t *testing.T) {
	dimensions := 3
	persist := false
	debug := false
	plot, _ := glot.NewPlot(dimensions, persist, debug)
	plot.AddPointGroup("Sample 1", "lines", []float64{2, 3, 4, 1})
	plot.SetTitle("Test Results")
	plot.SetZrange(-2, 2)
	err := plot.SavePlot("D:\\Test_line.jpeg")
	if err != nil {
		fmt.Printf("err %v", err)
	}
}

func PaintStockLine(code string) {
	//plot, _ := glot.NewPlot(dimensions, persist, debug)
}
