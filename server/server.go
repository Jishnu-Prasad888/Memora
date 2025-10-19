package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"Memora/commands"
	"Memora/store"
)

type Server struct {
	host           string
	port           string
	listener       net.Listener
	Store          *store.DataStore
	commandHandler *commands.CommandHandler
	protocol       *RESPProtocol
	clients        map[net.Conn]bool
	mu             sync.RWMutex
	shutdown       chan struct{}
}

func NewServer(host, port string) *Server {
	dataStore := store.NewDataStore()
	commandHandler := commands.NewCommandHandler(dataStore)

	return &Server{
		host:           host,
		port:           port,
		Store:          dataStore,
		commandHandler: commandHandler,
		protocol:       NewRESPProtocol(),
		clients:        make(map[net.Conn]bool),
		shutdown:       make(chan struct{}),
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = listener

	log.Printf("Redis server started on %s", addr)

	// Start background tasks
	go s.cleanupExpiredKeys()
	go s.handleSignals()

	// Main event loop
	for {
		select {
		case <-s.shutdown:
			return nil
		default:
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-s.shutdown:
					return nil
				default:
					log.Printf("Error accepting connection: %v", err)
					continue
				}
			}

			s.mu.Lock()
			s.clients[conn] = true
			s.mu.Unlock()

			go s.handleConnection(conn)
		}
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()
		conn.Close()
	}()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		// Set read timeout
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		command, err := s.protocol.ReadCommand(reader)
		if err != nil {
			if err == io.EOF {
				return
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			s.protocol.WriteError(writer, fmt.Sprintf("ERR %v", err))
			return
		}

		// Reset read deadline
		err = conn.SetReadDeadline(time.Time{})
		if err != nil {
			return
		}

		if len(command) == 0 {
			continue
		}

		result := s.commandHandler.HandleCommand(command)
		s.writeResponse(writer, result)
	}
}

func (s *Server) writeResponse(writer *bufio.Writer, result interface{}) {
	switch v := result.(type) {
	case string:
		s.protocol.WriteSimpleString(writer, v)
	case []byte:
		s.protocol.WriteBulkString(writer, v)
	case int:
		s.protocol.WriteInteger(writer, int64(v))
	case int64:
		s.protocol.WriteInteger(writer, v)
	case []interface{}:
		s.protocol.WriteArray(writer, v)
	case nil:
		s.protocol.WriteNull(writer)
	default:
		s.protocol.WriteBulkString(writer, []byte(fmt.Sprintf("%v", v)))
	}
}

func (s *Server) cleanupExpiredKeys() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.shutdown:
			return
		case <-ticker.C:
			removed := s.Store.RemoveExpired()
			if removed > 0 {
				log.Printf("Cleaned up %d expired keys", removed)
			}
		}
	}
}

func (s *Server) handleSignals() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	log.Println("Shutting down server...")

	close(s.shutdown)
	s.listener.Close()

	// Close all client connections
	s.mu.Lock()
	for conn := range s.clients {
		conn.Close()
	}
	s.mu.Unlock()
}

func (s *Server) Stop() {
	close(s.shutdown)
	if s.listener != nil {
		s.listener.Close()
	}
}
