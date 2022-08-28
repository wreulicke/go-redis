package decoder

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wreulicke/go-redis/data"
)

func TestDecodeSimpleString(t *testing.T) {
	tests := []struct {
		input    string
		expected data.Data
	}{
		{"+aaaa\r\n", &data.String{Value: "aaaa"}},
		{"+string bbb\r\n", &data.String{Value: "string bbb"}},
		{"+-12321421aaaa\r\n", &data.String{Value: "-12321421aaaa"}},
	}

	for _, test := range tests {
		b := bytes.NewBufferString(test.input)
		d := New(b)

		actual, err := d.Decode()
		assert.NoError(t, err)
		assertData(t, test.expected, actual)
	}
}

func TestDecodeError(t *testing.T) {
	tests := []struct {
		input    string
		expected data.Data
	}{
		{"-unexpected\r\n", &data.Error{Message: "unexpected"}},
		{"-:-1231985214*2193820194812tring bbb\r\n", &data.Error{Message: ":-1231985214*2193820194812tring bbb"}},
	}

	for _, test := range tests {
		b := bytes.NewBufferString(test.input)
		d := New(b)

		actual, err := d.Decode()
		assert.NoError(t, err)
		assertData(t, test.expected, actual)
	}
}

func TestDecodeInteger(t *testing.T) {

	tests := []struct {
		input    string
		expected data.Data
	}{
		{":1232414\r\n", &data.Integer{Value: 1232414}},
		{":-12020205891308653\r\n", &data.Integer{Value: -12020205891308653}},
	}

	for _, test := range tests {
		b := bytes.NewBufferString(test.input)
		d := New(b)

		actual, err := d.Decode()
		assert.NoError(t, err)
		assertData(t, test.expected, actual)
	}

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
