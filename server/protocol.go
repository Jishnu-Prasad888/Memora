package server

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
)

var (
	ErrInvalidFormat = errors.New("invalid RESP format")
	ErrEmptyCommand  = errors.New("empty command")
)

type RESPType byte

const (
	SimpleString RESPType = '+'
	Error        RESPType = '-'
	Integer      RESPType = ':'
	BulkString   RESPType = '$'
	Array        RESPType = '*'
)

type RESPValue struct {
	Type    RESPType
	Simple  string
	Integer int64
	Bulk    []byte
	Array   []*RESPValue
}

func (v *RESPValue) String() string {
	switch v.Type {
	case SimpleString:
		return v.Simple
	case Error:
		return v.Simple
	case Integer:
		return strconv.FormatInt(v.Integer, 10)
	case BulkString:
		return string(v.Bulk)
	case Array:
		var buf bytes.Buffer
		buf.WriteString("[")
		for i, item := range v.Array {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(item.String())
		}
		buf.WriteString("]")
		return buf.String()
	default:
		return "unknown"
	}
}

type RESPProtocol struct{}

func NewRESPProtocol() *RESPProtocol {
	return &RESPProtocol{}
}

func (r *RESPProtocol) ReadCommand(reader *bufio.Reader) ([]string, error) {
	value, err := r.readValue(reader)
	if err != nil {
		return nil, err
	}

	if value.Type != Array {
		return nil, ErrInvalidFormat
	}

	if len(value.Array) == 0 {
		return nil, ErrEmptyCommand
	}

	command := make([]string, len(value.Array))
	for i, arg := range value.Array {
		if arg.Type != BulkString {
			return nil, ErrInvalidFormat
		}
		command[i] = string(arg.Bulk)
	}

	return command, nil
}

func (r *RESPProtocol) readValue(reader *bufio.Reader) (*RESPValue, error) {
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	if len(line) < 2 {
		return nil, ErrInvalidFormat
	}

	// Remove \r\n
	line = line[:len(line)-2]

	switch RESPType(line[0]) {
	case SimpleString:
		return &RESPValue{Type: SimpleString, Simple: string(line[1:])}, nil
	case Error:
		return &RESPValue{Type: Error, Simple: string(line[1:])}, nil
	case Integer:
		num, err := strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return nil, ErrInvalidFormat
		}
		return &RESPValue{Type: Integer, Integer: num}, nil
	case BulkString:
		length, err := strconv.Atoi(string(line[1:]))
		if err != nil {
			return nil, ErrInvalidFormat
		}

		if length == -1 {
			return &RESPValue{Type: BulkString, Bulk: nil}, nil // Null bulk string
		}

		data := make([]byte, length+2) // +2 for \r\n
		if _, err := io.ReadFull(reader, data); err != nil {
			return nil, err
		}

		return &RESPValue{Type: BulkString, Bulk: data[:length]}, nil
	case Array:
		length, err := strconv.Atoi(string(line[1:]))
		if err != nil {
			return nil, ErrInvalidFormat
		}

		if length == -1 {
			return &RESPValue{Type: Array, Array: nil}, nil // Null array
		}

		array := make([]*RESPValue, length)
		for i := 0; i < length; i++ {
			value, err := r.readValue(reader)
			if err != nil {
				return nil, err
			}
			array[i] = value
		}

		return &RESPValue{Type: Array, Array: array}, nil
	default:
		return nil, ErrInvalidFormat
	}
}

func (r *RESPProtocol) WriteSimpleString(writer *bufio.Writer, value string) error {
	_, err := writer.WriteString(fmt.Sprintf("+%s\r\n", value))
	if err != nil {
		return err
	}
	return writer.Flush()
}

func (r *RESPProtocol) WriteError(writer *bufio.Writer, value string) error {
	_, err := writer.WriteString(fmt.Sprintf("-%s\r\n", value))
	if err != nil {
		return err
	}
	return writer.Flush()
}

func (r *RESPProtocol) WriteInteger(writer *bufio.Writer, value int64) error {
	_, err := writer.WriteString(fmt.Sprintf(":%d\r\n", value))
	if err != nil {
		return err
	}
	return writer.Flush()
}

func (r *RESPProtocol) WriteBulkString(writer *bufio.Writer, value []byte) error {
	if value == nil {
		_, err := writer.WriteString("$-1\r\n")
		if err != nil {
			return err
		}
	} else {
		_, err := writer.WriteString(fmt.Sprintf("$%d\r\n", len(value)))
		if err != nil {
			return err
		}
		_, err = writer.Write(value)
		if err != nil {
			return err
		}
		_, err = writer.WriteString("\r\n")
		if err != nil {
			return err
		}
	}
	return writer.Flush()
}

func (r *RESPProtocol) WriteNull(writer *bufio.Writer) error {
	_, err := writer.WriteString("$-1\r\n")
	if err != nil {
		return err
	}
	return writer.Flush()
}

func (r *RESPProtocol) WriteArray(writer *bufio.Writer, values []interface{}) error {
	if values == nil {
		_, err := writer.WriteString("*-1\r\n")
		if err != nil {
			return err
		}
	} else {
		_, err := writer.WriteString(fmt.Sprintf("*%d\r\n", len(values)))
		if err != nil {
			return err
		}

		for _, value := range values {
			switch v := value.(type) {
			case string:
				err = r.WriteBulkString(writer, []byte(v))
			case []byte:
				err = r.WriteBulkString(writer, v)
			case int:
				err = r.WriteInteger(writer, int64(v))
			case int64:
				err = r.WriteInteger(writer, v)
			case nil:
				err = r.WriteNull(writer)
			default:
				err = r.WriteBulkString(writer, []byte(fmt.Sprintf("%v", v)))
			}

			if err != nil {
				return err
			}
		}
	}
	return writer.Flush()
}
