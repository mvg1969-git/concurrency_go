package network

import (
	"bufio"
	"context"
	"fmt"
	"mv_db/internal/database"
	"net"

	"go.uber.org/zap"
)

type TCPServer struct {
	address string
	db      *database.Database
	logger  *zap.Logger
}

func NewTCPServer(address string, db *database.Database, logger *zap.Logger) *TCPServer {
	return &TCPServer{
		address: address,
		db:      db,
		logger:  logger,
	}
}

func (s *TCPServer) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer listener.Close()

	s.logger.Info("TCP Server started", zap.String("address", s.address))

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				s.logger.Error("failed to accept connection", zap.Error(err))
				continue
			}
		}
		go s.handleConnection(ctx, conn)
	}
}

func (s *TCPServer) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		// Читаем до символа переноса строки (простейший протокол)
		queryStr, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		// Вызываем бизнес-логику БД
		response := s.db.HandleQuery(ctx, queryStr)

		// Отправляем ответ клиенту
		_, err = conn.Write([]byte(response + "\n"))
		if err != nil {
			s.logger.Error("failed to write response", zap.Error(err))
			return
		}
	}
}
