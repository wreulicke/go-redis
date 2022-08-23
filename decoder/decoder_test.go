package decoder

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wreulicke/go-redis/data"
)

func TestDecodeSimpleString(t *testing.T) {
	b := bytes.NewBufferString("+aaaa\r\n")
	d := New(b)

	r, e := d.Decode()
	assert.NoError(t, e)
	result, ok := r.(*data.String)
	assert.True(t, ok)
	assert.Equal(t, "aaaa", result.Value)
}

func TestDecodeError(t *testing.T) {
	b := bytes.NewBufferString("-unexpected\r\n")
	d := New(b)

	r, e := d.Decode()
	assert.NoError(t, e)
	result, ok := r.(*data.Error)
	assert.True(t, ok)
	assert.Equal(t, "unexpected", result.Message)
}

func TestDecodeInteger(t *testing.T) {
	b := bytes.NewBufferString(":-12020205891308653\r\n")
	d := New(b)

	r, e := d.Decode()
	assert.NoError(t, e)
	result, ok := r.(*data.Integer)
	assert.True(t, ok)
	assert.Equal(t, int64(-12020205891308653), result.Value)
}

func TestDecodeBulkString(t *testing.T) {
	tests := []struct {
		input    string
		expected data.Data
	}{
		{"$-1\r\n", data.NULL},
		{"$5\r\nhello\r\n", &data.String{Value: "hello"}},
		{"$0\r\n\r\n", &data.String{Value: ""}},
	}

	for _, test := range tests {
		b := bytes.NewBufferString(test.input)
		d := New(b)

		actual, err := d.Decode()
		assert.NoError(t, err)
		assertData(t, test.expected, actual)
	}
}

func TestDecodeArray(t *testing.T) {
	tests := []struct {
		input    string
		expected data.Data
	}{
		{"*0\r\n", &data.Array{Values: []data.Data{}}},
		{"*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n", &data.Array{Values: []data.Data{
			&data.String{Value: "hello"},
			&data.String{Value: "world"},
		}}},
	}

	for _, test := range tests {
		b := bytes.NewBufferString(test.input)
		d := New(b)

		actual, err := d.Decode()
		assert.NoError(t, err)
		assertData(t, test.expected, actual)
	}
}

func assertData(t *testing.T, expected, actual data.Data) {
	t.Helper()
	assert.Equal(t, expected, actual)
}

func assertArray(t *testing.T, expected *data.Array, actual data.Data) {
	t.Helper()
	r, ok := actual.(*data.Array)
	if !ok {
		t.Error("actual is not array")
		return
	}
	assert.Equal(t, expected.Values, r.Values)
}
