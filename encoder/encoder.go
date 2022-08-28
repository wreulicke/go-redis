package encoder

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"github.com/wreulicke/go-redis/data"
)

type Encoder struct {
	writer *bufio.Writer
}

func New(w io.Writer) *Encoder {
	return &Encoder{
		writer: bufio.NewWriter(w),
	}
}

func (e *Encoder) Encode(d data.Data) error {
	defer e.writer.Flush()
	e.encode(d)
	_, err := e.write("\r\n")
	return err
}

func (e *Encoder) encode(d data.Data) error {
	switch v := d.(type) {
	case *data.Error:
		if _, err := e.write("-"); err != nil {
			return err
		}
		if _, err := e.write(v.Message); err != nil {
			return err
		}
		return nil
	case *data.Integer:
		if _, err := e.write(":"); err != nil {
			return err
		}
		_, err := e.write(fmt.Sprint(v.Value))
		return err
	case *data.Null:
		_, err := e.write("$-1")
		return err
	case *data.String:
		if _, err := e.write("$" + fmt.Sprint(len(v.Value)) + "\r\n"); err != nil {
			return err
		}
		if _, err := e.write(v.Value); err != nil {
			return err
		}
		return nil
	case *data.Array:
		if _, err := e.write("*" + fmt.Sprint(len(v.Values))); err != nil {
			return err
		}
		for _, v := range v.Values {
			if _, err := e.write("\r\n"); err != nil {
				return err
			}
			if err := e.encode(v); err != nil {
				return err
			}
		}
		return nil
	default:
		return errors.New("unsupported data type")
	}
}

func (e *Encoder) write(p string) (int, error) {
	return e.writer.WriteString(p)
}
