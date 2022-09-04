package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/wreulicke/go-redis/data"
	"github.com/wreulicke/go-redis/decoder"
	"github.com/wreulicke/go-redis/encoder"
)

const (
	B  int = 1
	KB int = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
)

type Server struct {
	cache *fastcache.Cache
}

func New() *Server {
	c := fastcache.New(2 * GB)
	return &Server{
		cache: c,
	}
}

func (s *Server) Serve() error {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		return err
	}

	fmt.Println("started localhost:6379")
	for {
		conn, err := l.Accept()
		if err != nil {
			// TODO use logger
			fmt.Println("Error accepting connection: ", err.Error())
		}

		go s.handleConnection(conn)
	}
}

func mainInternal() error {
	return New().Serve()
}

func expectArray(d data.Data, arr **data.Array) error {
	if v, ok := d.(*data.Array); ok {
		*arr = v
		return nil
	}
	return errors.New("expect array")
}

func expectString(d data.Data, str **data.String) error {
	if v, ok := d.(*data.String); ok {
		*str = v
		return nil
	}
	return errors.New("expect string")
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	decoder := decoder.New(r)
	encoder := encoder.New(conn)
	for {
		// TODO timeout
		d, err := decoder.Decode()
		if err != nil {
			encoder.Encode(&data.Error{
				Message: fmt.Sprintf("cannot decode command. %s", err.Error()),
			})
			continue
		}
		var arr *data.Array
		if err := expectArray(d, &arr); err != nil {
			encoder.Encode(&data.Error{
				Message: fmt.Sprintf("cannot decode command. %s", err.Error()),
			})
			continue
		}
		var command *data.String
		if err := expectString(arr.Values[0], &command); err != nil {
			_ = encoder.Encode(&data.Error{
				Message: fmt.Sprintf("cannot decode command. %s", err.Error()),
			})
			continue
		}
		switch strings.ToLower(command.Value) {
		case "ping":
			_ = encoder.Encode(&data.String{
				Value: "PONG",
			})
			continue
		case "get":
			var key *data.String
			if err := expectString(arr.Values[1], &key); err != nil {
				_ = encoder.Encode(&data.Error{
					Message: fmt.Sprintf("cannot decode key. %s", err.Error()),
				})
				continue
			}
			bs := s.cache.Get([]byte{}, []byte(key.Value))
			_ = encoder.Encode(&data.String{
				Value: string(bs),
			})
			continue
		case "set":
			var key *data.String
			if err := expectString(arr.Values[1], &key); err != nil {
				_ = encoder.Encode(&data.Error{
					Message: fmt.Sprintf("cannot decode key. %s", err.Error()),
				})
				continue
			}
			var value *data.String
			if err := expectString(arr.Values[2], &value); err != nil {
				_ = encoder.Encode(&data.Error{
					Message: fmt.Sprintf("cannot decode value. %s", err.Error()),
				})
				continue
			}
			s.cache.Set([]byte(key.Value), []byte(value.Value))
			_ = encoder.Encode(&data.String{
				Value: "OK",
			})
			continue
		}
	}
}

func main() {
	err := mainInternal()
	if err != nil {
		log.Fatal(err)
	}
}
