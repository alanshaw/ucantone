package datamodel

const Tag = "ctn-v1"

type ContainerModel struct {
	Ctn1 [][]byte `cborgen:"ctn-v1" dagjsongen:"ctn-v1"`
}
