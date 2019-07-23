package livesplit

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Microsoft/go-winio"
)

const (
	timeout = time.Duration(20) * time.Millisecond
)

type socket struct {
	sock net.Conn
}

func newSocket() *socket {
	return &socket{}
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
	sock, err := winio.DialPipe("\\\\.\\pipe\\LiveSplit", nil)
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
		logger.Printf("closing socket")
		if err := s.sock.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (s *socket) send(cmd []string) error {
	logger.Printf("send: %v", cmd)
	if err := s.establishConnectionIfNecessary(); err != nil {
		logger.Printf("error establishing connection: %v", err)
		return err
	}

	revert := s.setDeadlines(timeout)
	defer revert()

	_, err := fmt.Fprintf(s.sock, "%s\r\n", strings.Join(cmd, " "))
	if err != nil {
		logger.Printf("error in send: %v", err)

		if reconnectErr := s.reestablishConnection(); reconnectErr != nil {
			logger.Printf("error reestablishing connection after error: %v", reconnectErr)
			return err
		}

		logger.Printf("retrying send: %v", cmd)
		_, err = fmt.Fprintf(s.sock, "%s\r\n", strings.Join(cmd, " "))
		if err != nil {
			logger.Printf("error retrying send: %v", err)
		}
	}
	return err
}

func (s *socket) recv() (string, error) {
	logger.Printf("recv")
	if err := s.establishConnectionIfNecessary(); err != nil {
		logger.Printf("error establishing connection: %v", err)
		return "", err
	}

	revert := s.setDeadlines(timeout)
	defer revert()

	r, err := bufio.NewReader(s.sock).ReadString('\n')
	if err != nil {
		logger.Printf("error in recv: %v", err)
		if reconnectErr := s.reestablishConnection(); reconnectErr != nil {
			logger.Printf("error reestablishing connection after error: %v", reconnectErr)
		}
		return "", err
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
