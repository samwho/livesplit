package livesplit

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	port    = 16834
	timeout = 20 * time.Millisecond
)

type socket struct {
	port int
	sock net.Conn
}

func newSocket(port int) *socket {
	return &socket{port: port}
}

func (s *socket) establishConnectionIfNecessary() error {
	if s.sock == nil {
		logger.Printf("establishing connection")
		if err := s.reestablishConnection(); err != nil {
			return err
		}
	}
	return nil
}

func (s *socket) reestablishConnection() error {
	if err := s.Close(); err != nil {
		logger.Printf("error closing socket, ignoring as we're about to create a new connection: %v", err)
	}

	logger.Printf("establishing new livesplit connection")
	sock, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), timeout)
	if err != nil {
		logger.Printf("error establishing connection: %v", err)
		return err
	}
	logger.Printf("connection established")
	s.sock = sock
	return nil
}

func (s *socket) Close() error {
	if s.sock != nil {
		if err := s.sock.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (s *socket) send(cmd []string) error {
	logger.Printf("send: %v", cmd)
	if err := s.establishConnectionIfNecessary(); err != nil {
		return err
	}

	revert := s.setDeadlines(timeout)
	defer revert()

	_, err := fmt.Fprintf(s.sock, "%s\r\n", strings.Join(cmd, " "))
	if err != nil {
		if reconnectErr := s.reestablishConnection(); reconnectErr != nil {
			return err
		}

		_, err = fmt.Fprintf(s.sock, "%s\r\n", strings.Join(cmd, " "))
	}
	return err
}

func (s *socket) recv() (string, error) {
	logger.Printf("recv")
	if err := s.establishConnectionIfNecessary(); err != nil {
		return "", err
	}

	revert := s.setDeadlines(timeout)
	defer revert()

	r, err := bufio.NewReader(s.sock).ReadString('\n')
	if err != nil {
		if reconnectErr := s.reestablishConnection(); reconnectErr != nil {
			return "", err
		}

		retryR, retryErr := bufio.NewReader(s.sock).ReadString('\n')
		if retryErr != nil {
			return "", retryErr
		}
		r = retryR
	}

	ret := strings.TrimSuffix(r, "\r\n")
	logger.Printf("recv'd: %v", ret)
	return ret, nil
}

func (s *socket) sendAndRecv(cmd []string) (string, error) {
	if err := s.send(cmd); err != nil {
		return "", err
	}
	return s.recv()
}

func (s *socket) sendAndRecvTime(cmd []string) (time.Duration, error) {
	r, err := s.sendAndRecv(cmd)
	if err != nil {
		return 0, err
	}
	return StringToDuration(r)
}

func (s *socket) sendAndRecvInt(cmd []string) (int, error) {
	r, err := s.sendAndRecv(cmd)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(r)
}

func (s *socket) setDeadlines(timeout time.Duration) func() {
	if err := s.sock.SetDeadline(time.Now().Add(timeout)); err != nil {
		logger.Printf("unable to set deadline: %v", err)
	}
	return func() {
		if err := s.sock.SetDeadline(time.Time{}); err != nil {
			logger.Printf("unable to set deadline: %v", err)
		}
	}
}
