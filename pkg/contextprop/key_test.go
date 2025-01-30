package contextprop

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetAndGetValue(t *testing.T) {
	ctx := SetValue(context.Background(), CorrelationIDKey, "123")
	assert.Equal(t, "123", GetValue(ctx, CorrelationIDKey))
}
