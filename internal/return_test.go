package internal

import (
	"testing"
)

func TestReturnObject_nilReturn(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(emptyReturn).Equals(Return{})
}
