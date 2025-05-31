package nn

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"gorgonia.org/gorgonia"
	"gorgonia.org/tensor"
)

const (
	ActivationReLu    = "relu"
	ActivationSoftmax = "softmax"
)

type Layer struct {
	Activation string
	Size       [2]int
	Dropout    float64
}

type NeuralNetwork struct {
	predVal gorgonia.Value
	g       *gorgonia.ExprGraph

	out *gorgonia.Node

	weights     []*gorgonia.Node
	activations []string
	dropouts    []float64
}

func NewNeuralNetwork(g *gorgonia.ExprGraph, layers []Layer) (*NeuralNetwork, error) {
	if len(layers) == 0 {
		return nil, ErrEmptyLayers
	}

	var weights []*gorgonia.Node
	var activations []string
	var dropouts []float64

	for i, layer := range layers {
		name := strings.Join([]string{"w", strconv.Itoa(i)}, "")
		weights = append(weights,
			gorgonia.NewMatrix(
				g,
				tensor.Float64,
				gorgonia.WithShape(layer.Size[0], layer.Size[1]),
				gorgonia.WithName(name),
				gorgonia.WithInit(gorgonia.GlorotN(1.0)),
			), //nolint:mnd
		)
		activations = append(activations, layer.Activation)
		dropouts = append(dropouts, layer.Dropout)
	}

	return &NeuralNetwork{
		g:           g,
		activations: activations,
		weights:     weights,
		dropouts:    dropouts,
	}, nil
}

func (m *NeuralNetwork) GetVal() gorgonia.Value {
	return m.predVal
}

func (m *NeuralNetwork) GetGraph() *gorgonia.ExprGraph {
	return m.g
}

func (m *NeuralNetwork) Forward(x *gorgonia.Node) (err error) {
	var layer *gorgonia.Node

	layer = x
	for i, weight := range m.weights {
		if layer, err = gorgonia.Mul(layer, weight); err != nil {
			return errors.Wrapf(err, "Unable to multiply layer and weight: %v, %v, %d", layer, weight, i)
		}

		fn := m.activations[i]

		switch fn {
		case ActivationReLu:
			layer, err = gorgonia.Rectify(layer)
			if err != nil {
				return err
			}
		case ActivationSoftmax:
			layer, err = gorgonia.SoftMax(layer)
			if err != nil {
				return err
			}
		}
		d := m.dropouts[i]

		layer, err = gorgonia.Dropout(layer, d)
		if err != nil {
			return err
		}
	}
	m.out = layer
	gorgonia.Read(m.out, &m.predVal)

	return
}

func (m *NeuralNetwork) Learnables() gorgonia.Nodes {
	return m.weights
}

func (m *NeuralNetwork) Learn(y *gorgonia.Node) error {
	losses, err := gorgonia.HadamardProd(m.out, y)
	if err != nil {
		return err
	}
	cost := gorgonia.Must(gorgonia.Mean(losses))
	cost = gorgonia.Must(gorgonia.Neg(cost))

	// we wanna track costs
	var costVal gorgonia.Value
	gorgonia.Read(cost, &costVal)

	_, err = gorgonia.Grad(cost, m.Learnables()...)

	return err
}

func Solve(m *NeuralNetwork, x *gorgonia.Node, y *gorgonia.Node) (gorgonia.Value, error) {
	err := m.Forward(x)
	if err != nil {
		return m.GetVal(), err
	}

	err = m.Learn(y)
	if err != nil {
		return m.GetVal(), err
	}

	sh := y.Shape()
	vm := gorgonia.NewTapeMachine(m.GetGraph(), gorgonia.BindDualValues(m.Learnables()...))
	solver := gorgonia.NewRMSPropSolver(gorgonia.WithBatchSize(float64(sh[0])))

	if err = vm.RunAll(); err != nil {
		return m.GetVal(), err
	}

	err = solver.Step(gorgonia.NodesToValueGrads(m.Learnables()))
	if err != nil {
		return m.GetVal(), err
	}

	vm.Reset()

	return m.GetVal(), vm.Close()
}
