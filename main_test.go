package main

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
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
}

func (t *TypelistSuite) TestTypelist() {
	expected, err := parse("./tests/test.go")
	t.Require().NoError(err)

	var exp interface{}
	var act interface{}

	err = json.Unmarshal([]byte(expected), &exp)
	t.Require().NoError(err)
	actual, err := ioutil.ReadFile("./tests/test.json")
	t.Require().NoError(err)

	err = json.Unmarshal(actual, &act)
	t.Require().NoError(err)

	t.True(reflect.DeepEqual(exp, act), "output from parse should equal actual json")
}
