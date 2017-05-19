package main

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/suite"
)

type TypelistSuite struct {
	suite.Suite
}

func TestTypelistSuite(t *testing.T) {
	suite.Run(t, new(TypelistSuite))
}

func (t *TypelistSuite) SetupTest() {}
func (t *TypelistSuite) TearDownTest() {
	expected, err := parse("./tests/test.go")
	t.Require().NoError(err)
	actual, err := ioutil.ReadFile("./tests/test.json")
	t.Require().NoError(err)
	t.Equal(expected, string(actual), "output from parse should equal actual json")
}
