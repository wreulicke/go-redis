package decoder

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strconv"
	"unicode/utf8"

	"github.com/wreulicke/go-redis/data"
)

type Decoder struct {
	reader *bufio.Reader
	buffer bytes.Buffer
}

type Result struct {
	Value data.Data
}

func New(r io.Reader) *Decoder {
	return &Decoder{
		reader: bufio.NewReader(r),
	}
}

func (d *Decoder) Decode() (data.Data, error) {
	data, err := d.decode()
	if err := d.expect('\r'); err != nil {
		return nil, err
	}
	if err := d.expect('\n'); err != nil {
		return nil, err
	}
	return data, err
}

func (d *Decoder) decode() (data.Data, error) {
	r, err := d.next()
	if err != nil {
		// TODO add information to error
		return nil, err
	}
	defer d.buffer.Reset()
	switch r {
	case '+':
		decoded, err := d.decodeString()
		return &data.String{
			Value: decoded,
		}, err
	case '-':
		decoded, err := d.decodeString()
		return &data.Error{
			Message: decoded,
		}, err
	case ':':
		decoded, err := d.decodeInteger()
		return &data.Integer{
			Value: decoded,
		}, err
	case '$':
		return d.decodeBulkString()
	case '*':
		return d.decodeArray()
	default:
		return nil, errors.New("unknown data type")
	}
}

func (d *Decoder) decodeInteger() (int64, error) {
	raw, err := d.decodeString()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(raw, 10, 64)
}

func (d *Decoder) decodeString() (string, error) {
	for {
		r, err := d.next()
		if err != nil {
			return "", err
		}
		_, err = d.buffer.WriteRune(r)
		if err != nil {
			return "", err
		}
		bs, err := d.reader.Peek(2)
		// TODO consider io.EOF
		if err != nil {
			return "", err
		}
		if string(bs) == "\r\n" {
			return d.buffer.String(), err
		}
	}
}

func (d *Decoder) decodeBulkString() (data.Data, error) {
	n, err := d.decodeInteger()
	if err != nil {
		return nil, err
	}
	if n == -1 {
		return data.NULL, nil
	}
	if err := d.expect('\r'); err != nil {
		return nil, err
	}
	if err := d.expect('\n'); err != nil {
		return nil, err
	}
	b := bytes.Buffer{}
	size := int64(2048)
	if n < size {
		size = n
	}
	bs := make([]byte, size)
	for i := int64(0); i < n; {
		read, err := d.reader.Read(bs)
		// TODO catch eof with read > 0
		if err != nil {
			return nil, err
		}
		_, err = b.Write(bs[:read])
		if err != nil {
			return nil, err
		}
		i = i + int64(read)
		if n-i < 2048 {
			bs = bs[:n-i]
		}
	}
	return &data.String{
		Value: b.String(),
	}, nil
}

func (d *Decoder) decodeArray() (data.Data, error) {
	n, err := d.decodeInteger()
	if err != nil {
		return nil, err
	}
	d.buffer.Reset()
	r := []data.Data{}
	for i := int64(0); i < n; i++ {
		if err := d.expect('\r'); err != nil {
			return nil, err
		}
		if err := d.expect('\n'); err != nil {
			return nil, err
		}
		decoded, err := d.decode()
		if err != nil {
			return nil, err
		}
		r = append(r, decoded)
	}
	return &data.Array{
		Values: r,
	}, nil
}

const eof = -1

func (d *Decoder) next() (rune, error) {
	r, _, err := d.reader.ReadRune()
	if err == io.EOF {
		return eof, err
	}
	return r, nil
}

func (d *Decoder) peek() rune {
	lead, err := d.reader.Peek(1)
	if err == io.EOF {
		return eof
	} else if err != nil {
		return 0
	}

	p, err := d.reader.Peek(runeLen(lead[0]))

	if err == io.EOF {
		return eof
	} else if err != nil {
		return 0
	}

	ruNe, _ := utf8.DecodeRune(p)
	return ruNe
}

func (d *Decoder) expect(ruNe rune) error {
	r, err := d.next()
	if err != nil {
		return err
	}
	if r != ruNe {
		return errors.New("unexpected rune")
	}
	return nil
}

func runeLen(lead byte) int {
	if lead < 0xC0 {
		return 1
	} else if lead < 0xE0 {
		return 2
	} else if lead < 0xF0 {
		return 3
	}
	return 4
}
