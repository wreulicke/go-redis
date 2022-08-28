package encoder

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wreulicke/go-redis/data"
)

func TestEncodeError(t *testing.T) {
	tests := []struct {
		expected string
		input    data.Data
	}{
		{"-unexpected\r\n", &data.Error{Message: "unexpected"}},
		{"-:-1231985214*2193820194812tring bbb\r\n", &data.Error{Message: ":-1231985214*2193820194812tring bbb"}},
	}

	for _, test := range tests {
		b := &bytes.Buffer{}
		e := New(b)

		err := e.Encode(test.input)
		assert.NoError(t, err)
		assert.Equal(t, test.expected, b.String())
	}
}

func TestEncodeInteger(t *testing.T) {
	tests := []struct {
		expected string
		input    data.Data
	}{
		{":1232414\r\n", &data.Integer{Value: 1232414}},
		{":-12020205891308653\r\n", &data.Integer{Value: -12020205891308653}},
	}

	for _, test := range tests {
		b := &bytes.Buffer{}
		e := New(b)

		err := e.Encode(test.input)
		assert.NoError(t, err)
		assert.Equal(t, test.expected, b.String())
	}

}

func TestEncodeString(t *testing.T) {
	tests := []struct {
		expected string
		input    data.Data
	}{
		{"$-1\r\n", data.NULL},
		{"$5\r\nhello\r\n", &data.String{Value: "hello"}},
		{"$0\r\n\r\n", &data.String{Value: ""}},
	}

	for _, test := range tests {
		b := &bytes.Buffer{}
		e := New(b)

		err := e.Encode(test.input)
		assert.NoError(t, err)
		assert.Equal(t, test.expected, b.String())
	}
}

func TestEncodeArray(t *testing.T) {
	tests := []struct {
		expected string
		input    data.Data
	}{
		{"*0\r\n", &data.Array{Values: []data.Data{}}},
		{"*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n", &data.Array{Values: []data.Data{
			&data.String{Value: "hello"},
			&data.String{Value: "world"},
		}}},
		{"*1\r\n*0\r\n", &data.Array{Values: []data.Data{
			&data.Array{},
		}}},
	}

	for _, test := range tests {
		b := &bytes.Buffer{}
		e := New(b)

		err := e.Encode(test.input)
		assert.NoError(t, err)
		assert.Equal(t, test.expected, b.String())
	}
}
