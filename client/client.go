package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type Client struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
}

func NewClient(host, port string) (*Client, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) SendCommand(command []string) (interface{}, error) {
	if len(command) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	// If it's a single word and not a known command, treat it as GET
	if len(command) == 1 {
		cmd := strings.ToUpper(command[0])
		// Check if it's a known command
		knownCommands := map[string]bool{
			"PING": true, "ECHO": true, "FLUSHALL": true, "DBSIZE": true,
		}

		if !knownCommands[cmd] {
			// Treat as GET command
			command = []string{"GET", command[0]}
		}
	}

	// Build RESP array
	resp := fmt.Sprintf("*%d\r\n", len(command))
	for _, arg := range command {
		resp += fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg)
	}

	_, err := c.writer.WriteString(resp)
	if err != nil {
		return nil, err
	}
	err = c.writer.Flush()
	if err != nil {
		return nil, err
	}

	return c.readResponse()
}
func (c *Client) readResponse() (interface{}, error) {
	line, err := c.reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	if len(line) < 2 {
		return nil, fmt.Errorf("invalid response")
	}

	line = line[:len(line)-2] // Remove \r\n

	switch line[0] {
	case '+': // Simple string
		return line[1:], nil
	case '-': // Error
		return nil, fmt.Errorf(line[1:])
	case ':': // Integer
		return strconv.ParseInt(line[1:], 10, 64)
	case '$': // Bulk string
		length, err := strconv.Atoi(line[1:])
		if err != nil {
			return nil, err
		}
		if length == -1 {
			return nil, nil // Null bulk string
		}

		data := make([]byte, length+2) // +2 for \r\n
		_, err = c.reader.Read(data)
		if err != nil {
			return nil, err
		}
		return string(data[:length]), nil
	case '*': // Array
		length, err := strconv.Atoi(line[1:])
		if err != nil {
			return nil, err
		}
		if length == -1 {
			return nil, nil // Null array
		}

		array := make([]interface{}, length)
		for i := 0; i < length; i++ {
			item, err := c.readResponse()
			if err != nil {
				return nil, err
			}
			array[i] = item
		}
		return array, nil
	default:
		return nil, fmt.Errorf("unknown response type: %s", line)
	}
}

// Convenience methods
func (c *Client) Set(key, value string) (string, error) {
	result, err := c.SendCommand([]string{"SET", key, value})
	if err != nil {
		return "", err
	}
	return result.(string), nil
}

func (c *Client) Get(key string) (string, error) {
	result, err := c.SendCommand([]string{"GET", key})
	if err != nil {
		return "", err
	}
	if result == nil {
		return "", nil
	}
	return result.(string), nil
}

func (c *Client) Del(keys ...string) (int64, error) {
	command := append([]string{"DEL"}, keys...)
	result, err := c.SendCommand(command)
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}

func (c *Client) Exists(keys ...string) (int64, error) {
	command := append([]string{"EXISTS"}, keys...)
	result, err := c.SendCommand(command)
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}

func (c *Client) Keys(pattern string) ([]interface{}, error) {
	result, err := c.SendCommand([]string{"KEYS", pattern})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return []interface{}{}, nil
	}
	return result.([]interface{}), nil
}

func (c *Client) Ping() (string, error) {
	result, err := c.SendCommand([]string{"PING"})
	if err != nil {
		return "", err
	}
	return result.(string), nil
}

// Interactive CLI
func (c *Client) StartCLI() {
	fmt.Println("Connected to Redis server. Type commands or 'quit' to exit.")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "quit" || input == "exit" {
			break
		}
		if input == "" {
			continue
		}

		// Parse command
		parts := strings.Fields(input)
		result, err := c.SendCommand(parts)
		if err != nil {
			fmt.Printf("(error) %v\n", err)
			continue
		}

		// Pretty print result
		switch v := result.(type) {
		case string:
			fmt.Printf("\"%s\"\n", v)
		case int64:
			fmt.Printf("(integer) %d\n", v)
		case []interface{}:
			if len(v) == 0 {
				fmt.Println("(empty array)")
			} else {
				for i, item := range v {
					fmt.Printf("%d) %v\n", i+1, item)
				}
			}
		case nil:
			fmt.Println("(nil)")
		default:
			fmt.Printf("%v\n", v)
		}
	}
}
