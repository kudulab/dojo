package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_checkIfInteractive(t *testing.T){
	i := checkIfInteractive()
	assert.False(t, i)
}