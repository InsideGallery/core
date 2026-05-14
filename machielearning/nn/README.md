# machielearning/nn

Import path: `github.com/InsideGallery/core/machielearning/nn`

`nn` wraps a small Gorgonia-based neural-network flow. It builds a sequence of
matrix weights, runs a forward pass, computes gradients from a simple negative
mean loss, and applies one RMSProp solver step.

## Main API

- `Layer` describes one network layer with `Activation`, `Size`, and `Dropout`.
- `NewNeuralNetwork(graph, layers)` creates matrix weights with Glorot
  initialization. Empty layer lists return `ErrEmptyLayers`.
- `ActivationReLu` and `ActivationSoftmax` are the recognized activation names.
  Any other activation string leaves the multiplied layer unchanged before
  dropout.
- `Forward(x)` multiplies through the weights, applies activation and dropout,
  and registers the output value for reading.
- `Learn(y)` computes gradients for the network weights.
- `Solve(network, x, y)` runs `Forward`, `Learn`, a Gorgonia tape machine, and
  one RMSProp update, then returns the last prediction value.
- `Learnables()` exposes the weight nodes for Gorgonia solvers.
- `GetGraph()` and `GetVal()` expose the underlying graph and last prediction.

## Usage

```go
package example

import (
	"github.com/InsideGallery/core/machielearning/nn"
	"gorgonia.org/gorgonia"
	"gorgonia.org/tensor"
)

func trainOnce() (gorgonia.Value, error) {
	graph := gorgonia.NewGraph()
	network, err := nn.NewNeuralNetwork(graph, []nn.Layer{
		{Activation: nn.ActivationReLu, Size: [2]int{784, 300}},
		{Activation: nn.ActivationSoftmax, Size: [2]int{300, 10}},
	})
	if err != nil {
		return nil, err
	}

	x := gorgonia.NewMatrix(network.GetGraph(), tensor.Float64,
		gorgonia.WithShape(10, 784),
		gorgonia.WithName("x"),
		gorgonia.WithInit(gorgonia.GlorotN(1.0)),
	)
	y := gorgonia.NewMatrix(network.GetGraph(), tensor.Float64,
		gorgonia.WithShape(10, 10),
		gorgonia.WithName("y"),
		gorgonia.WithInit(gorgonia.GlorotN(1.0)),
	)

	return nn.Solve(network, x, y)
}
```

## Compatibility Notes

This package depends directly on Gorgonia and tensor types. The repository's
`gorgonia_test.go` is guarded by the `local_test` build tag, so those tests are
not part of the default `go test ./...` package run.
