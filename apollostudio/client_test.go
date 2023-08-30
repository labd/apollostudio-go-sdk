package apollostudio

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateGraphRef(t *testing.T) {
	tests := []struct {
		ref   string
		valid bool
	}{
		{ref: "foo@bar", valid: true},
		{ref: "foobar", valid: false},
	}

	for _, test := range tests {
		valid := isValidGraphRef(test.ref)
		assert.Equal(t, test.valid, valid)
	}
}

func TestGraphRefGetGraphId(t *testing.T) {
	graphRef := GraphRef("foo@bar")

	assert.Equal(t, "foo", graphRef.getGraphId())
}

func TestGraphRefGetVariant(t *testing.T) {
	graphRef := GraphRef("foo@bar")

	assert.Equal(t, "bar", graphRef.getVariant())
}
