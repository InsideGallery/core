//go:build local_test
// +build local_test

package nn

import (
	"fmt"
	"log"
	"math"
	"testing"

	"github.com/InsideGallery/core/testutils"
	"github.com/spf13/cast"

	"gorgonia.org/tensor"

	"gorgonia.org/gorgonia"
)

func TestLn(t *testing.T) {
	g := gorgonia.NewGraph()
	values := tensor.New(tensor.WithShape(3), tensor.WithBacking([]float64{0.0001, 1, 200}))
	nv := gorgonia.NewVector(g, gorgonia.Float64, gorgonia.WithName("x"), gorgonia.WithValue(values))
	logx, err := gorgonia.Log(nv)
	testutils.Equal(t, err, nil)
	m := gorgonia.NewTapeMachine(g)
	err = m.RunAll()
	testutils.Equal(t, err, nil)
	fmt.Println(logx, logx.Value())
}

func TestSoftmax(t *testing.T) {
	g := gorgonia.NewGraph()
	values := tensor.New(tensor.WithShape(3), tensor.WithBacking([]float64{1, 2, 3}))
	nv := gorgonia.NewVector(g, gorgonia.Float64, gorgonia.WithName("x"), gorgonia.WithValue(values))
	sigma, err := gorgonia.SoftMax(nv)
	// num, err := gorgonia.Exp(nv)
	// testutils.Equal(t, err, nil)
	// den, err := gorgonia.Sum(num)
	// testutils.Equal(t, err, nil)
	// sigma, err := gorgonia.Div(num, den)
	// testutils.Equal(t, err, nil)

	m := gorgonia.NewTapeMachine(g)
	err = m.RunAll()
	testutils.Equal(t, err, nil)

	fmt.Println(sigma, sigma.Value())
}

func TestMulTwoVectorsDUDL(t *testing.T) {
	g := gorgonia.NewGraph()

	values := tensor.New(tensor.WithShape(4), tensor.WithBacking([]float64{1, 2, 3, 4}))
	values2 := tensor.New(tensor.WithShape(4), tensor.WithBacking([]float64{0, 1, 0, -1}))

	nv := gorgonia.NewVector(g, gorgonia.Float64, gorgonia.WithName("x"), gorgonia.WithValue(values))
	fmt.Println(nv.Value(), nv.IsVector(), nv.ID())

	nv2 := gorgonia.NewVector(g, gorgonia.Float64, gorgonia.WithName("y"), gorgonia.WithValue(values2))
	fmt.Println(nv2.Value(), nv2.IsVector(), nv2.ID())

	res, err := gorgonia.Mul(nv, nv2)
	testutils.Equal(t, err, nil)

	m := gorgonia.NewTapeMachine(g)
	err = m.RunAll()
	testutils.Equal(t, err, nil)

	testutils.Equal(t, res.Value().Data().(float64), float64(-2))
	fmt.Println(res, res.DataSize(), res.Value())
}

func TestMulTwoVectors(t *testing.T) {
	g := gorgonia.NewGraph()

	values := tensor.New(tensor.WithShape(5), tensor.WithBacking([]float64{1, 0, 2, 5, -2}))
	values2 := tensor.New(tensor.WithShape(5), tensor.WithBacking([]float64{2, 8, -6, 1, 0}))

	nv := gorgonia.NewVector(g, gorgonia.Float64, gorgonia.WithName("x"), gorgonia.WithValue(values))
	fmt.Println(nv.Value(), nv.IsVector(), nv.ID())

	nv2 := gorgonia.NewVector(g, gorgonia.Float64, gorgonia.WithName("y"), gorgonia.WithValue(values2))
	fmt.Println(nv2.Value(), nv2.IsVector(), nv2.ID())

	res, err := gorgonia.Mul(nv, nv2)
	testutils.Equal(t, err, nil)

	m := gorgonia.NewTapeMachine(g)
	err = m.RunAll()
	testutils.Equal(t, err, nil)

	testutils.Equal(t, res.Value().Data().(float64), float64(-5))
	fmt.Println(res, res.DataSize(), res.Value())
}

func TestMatrixDUDL(t *testing.T) {
	g := gorgonia.NewGraph()

	values := tensor.New(tensor.WithShape(1, 6), tensor.WithBacking([]float64{1, 2, 3, 4, 5, 6}))
	values2 := tensor.New(tensor.WithShape(1, 6), tensor.WithBacking([]float64{7, 8, 9, 10, 11, 12}))

	nv := gorgonia.NewMatrix(g, gorgonia.Float64, gorgonia.WithName("x"), gorgonia.WithValue(values))
	fmt.Println(nv.Value(), nv.IsMatrix(), nv.ID())

	nv2 := gorgonia.NewMatrix(g, gorgonia.Float64, gorgonia.WithName("y"), gorgonia.WithValue(values2))
	fmt.Println(nv2.Value(), nv2.IsMatrix(), nv2.ID())
	nv2T, err := gorgonia.Transpose(nv2)
	testutils.Equal(t, err, nil)

	res, err := gorgonia.Mul(nv, nv2T)
	testutils.Equal(t, err, nil)

	m := gorgonia.NewTapeMachine(g)
	err = m.RunAll()
	testutils.Equal(t, err, nil)

	testutils.Equal(t, res.Value().Data().([]float64), []float64{217})
	fmt.Println(res, res.DataSize(), res.Value())
}

func TestEntropy(t *testing.T) {
	p := []float64{1.0, 0.0}
	q := []float64{0.25, 0.75}
	var h float64
	for i := 0; i < len(q); i++ {
		h -= p[i] * math.Log(q[i])
	}
	fmt.Println("Correct entropy: ", h)

	g := gorgonia.NewGraph()

	values := tensor.New(tensor.WithShape(2), tensor.WithBacking(p))
	values2 := tensor.New(tensor.WithShape(2), tensor.WithBacking(q))
	nv := gorgonia.NewVector(g, gorgonia.Float64, gorgonia.WithName("x"), gorgonia.WithValue(values))
	nv2 := gorgonia.NewVector(g, gorgonia.Float64, gorgonia.WithName("y"), gorgonia.WithValue(values2))

	val, err := gorgonia.BinaryXent(nv2, nv)
	testutils.Equal(t, err, nil)

	m := gorgonia.NewTapeMachine(g)
	err = m.RunAll()
	testutils.Equal(t, err, nil)

	fmt.Println(val.Value())
}

func TestMean(t *testing.T) {
	x := []float64{1, 2, 4, 6, 5, 4, 0}
	n := float64(len(x))

	g := gorgonia.NewGraph()

	values := tensor.New(tensor.WithShape(len(x)), tensor.WithBacking(x))
	nv := gorgonia.NewVector(g, gorgonia.Float64, gorgonia.WithName("x"), gorgonia.WithValue(values))

	pw := gorgonia.NewScalar(g, gorgonia.Float64, gorgonia.WithName("p"))
	ln := gorgonia.NewScalar(g, gorgonia.Float64, gorgonia.WithName("l"))
	bs := gorgonia.NewScalar(g, gorgonia.Float64, gorgonia.WithName("b"))
	gorgonia.Let(pw, 2.0)
	gorgonia.Let(ln, float64(n))
	gorgonia.Let(bs, 1.0)

	meanG, err := gorgonia.Mean(nv)
	testutils.Equal(t, err, nil)

	op1, err := gorgonia.Sub(nv, meanG)
	testutils.Equal(t, err, nil)
	op2, err := gorgonia.Pow(op1, pw)
	testutils.Equal(t, err, nil)
	op3, err := gorgonia.Sum(op2)
	testutils.Equal(t, err, nil)

	m := gorgonia.NewTapeMachine(g)
	err = m.RunAll()
	testutils.Equal(t, err, nil)

	fmt.Println(cast.ToFloat64(meanG.Value().Data()))
	fmt.Println(1 / (float64(n) - 1) * cast.ToFloat64(op3.Value().Data()))
}

func TestGorgonia(t *testing.T) {
	g := gorgonia.NewGraph()
	network, err := NewNeuralNetwork(g, []Layer{
		{
			Activation: ActivationReLu,
			Size:       [2]int{784, 300},
		},
		{
			Activation: ActivationSoftmax,
			Size:       [2]int{300, 10},
		},
	})
	testutils.Equal(t, err, nil)

	x := gorgonia.NewMatrix(network.GetGraph(), tensor.Float64, gorgonia.WithShape(10, 784), gorgonia.WithName("x"), gorgonia.WithInit(gorgonia.GlorotN(1.0))) //nolint:mnd
	y := gorgonia.NewMatrix(network.GetGraph(), tensor.Float64, gorgonia.WithShape(10, 10), gorgonia.WithName("y"), gorgonia.WithInit(gorgonia.GlorotN(1.0)))  //nolint:mnd

	value, err := Solve(network, x, y)
	testutils.Equal(t, err, nil)

	fmt.Println(value)
}

// Basic example of representing mathematical equations as graphs.
//
// In this example, we want to represent the following equation
//
//	z = x + y
func Example_basic() {
	g := gorgonia.NewGraph()

	var x, y, z *gorgonia.Node
	var err error

	// define the expression
	x = gorgonia.NewScalar(g, gorgonia.Float64, gorgonia.WithName("x"))
	y = gorgonia.NewScalar(g, gorgonia.Float64, gorgonia.WithName("y"))
	if z, err = gorgonia.Add(x, y); err != nil {
		log.Fatal(err)
	}

	// create a VM to run the program on
	machine := gorgonia.NewTapeMachine(g)
	defer machine.Close()

	// set initial values then run
	gorgonia.Let(x, 2.0)
	gorgonia.Let(y, 2.5)
	if err = machine.RunAll(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v", z.Value())
	// Output: 4.5
}
