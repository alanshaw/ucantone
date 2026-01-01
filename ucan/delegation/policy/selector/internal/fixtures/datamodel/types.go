package datamodel

import jsg "github.com/alanshaw/dag-json-gen"

type FixtureModel struct {
	Name     string        `dagjsongen:"name"`
	Selector string        `dagjsongen:"selector"`
	Input    jsg.Deferred  `dagjsongen:"input"`
	Output   *jsg.Deferred `dagjsongen:"output,omitempty"`
}

type FixturesModel struct {
	Success []FixtureModel `dagjsongen:"pass"`
	Null    []FixtureModel `dagjsongen:"null"`
	Error   []FixtureModel `dagjsongen:"fail"`
}
