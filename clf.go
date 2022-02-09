package main

import (
  "fmt"
  "io/ioutil"
  "github.com/pa-m/sklearn/preprocessing"
  "github.com/go-gota/gota/dataframe"
  "gonum.org/v1/gonum/mat"
  "gorgonia.org/tensor"
  "github.com/owulveryck/onnx-go"
  "github.com/owulveryck/onnx-go/backend/x/gorgonnx"
)

type matrix struct {
	dataframe.DataFrame
}

func (m matrix) At(i, j int) float64 {
	return m.Elem(i, j).Float()
}

func (m matrix) T() mat.Matrix {
	return mat.Transpose{m}
}

type Classifier struct {
  Scaler *preprocessing.MinMaxScaler
  Model *onnx.Model
  Backend *gorgonnx.Graph
}

func InitClassifier() Classifier {
    backend := gorgonnx.NewGraph()
  	model := onnx.NewModel(backend)

  	b, _ := ioutil.ReadFile("../models/onnx/pnd_v2s_mtl_2lh24.onnx")
  	err := model.UnmarshalBinary(b)
  	if err != nil {
  		panic(err)
  	}

  return Classifier {
    Scaler: preprocessing.NewMinMaxScaler([]float64{0, 1}),
    Model: model,
    Backend: backend,
  }
}

func (clf *Classifier) Predict(trades [][]Trade) {
  all := []*tensor.Dense{}

  for i := 0; i < len(trades); i++ {
    sts := trades[i]

    df := dataframe.LoadStructs(sts)
    scaled, _ := clf.Scaler.FitTransform(matrix{df.Select([]string{"Timestamp", "Price", "Amount"})}, nil)
    scaledDf := dataframe.LoadMatrix(scaled)
    scaledDf.SetNames("Timestamp", "Price", "Amount")

    df = df.Mutate(
  		scaledDf.Col("Timestamp"),
  	).Mutate(
  		scaledDf.Col("Price"),
  	).Mutate(
  		scaledDf.Col("Amount"),
  	)

    all = append(all, tensor.FromMat64(mat.DenseCopyOf(&matrix{df})))
  }

  concat, err := all[0].Stack(0, all[1:]...)
  fmt.Println(concat.Shape())

  if err != nil {
    panic(err)
  }

  clf.Model.SetInput(0, concat)
  err = clf.Backend.Run()
  if err != nil {
    panic(err)
  }

  output, _ := clf.Model.GetOutputTensors()

  fmt.Println(output[0])
}
