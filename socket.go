package livesplit

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	port    = 16834
	timeout = 5 * time.Millisecond
)

type socket struct {
	port int
	sock net.Conn
}

func newSocket(port int) *socket {
	return &socket{port: port}
}

func (s *socket) establishConnection() error {
	if err := s.Close(); err != nil {
		log.Printf("error closing livesplit socket, ignoring as we're about to create a new connection: %v", err)
	}

	log.Printf("establishing new livesplit connection")
	sock, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), timeout)
	if err != nil {
		log.Printf("error establishing livesplit connection: %v", err)
		return err
	}
	log.Printf("livesplit connection established")
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
	if s.sock == nil {
		if err := s.establishConnection(); err != nil {
			return err
		}
	}

	if err := s.sock.SetDeadline(time.Now().Add(timeout)); err != nil {
		log.Printf("unable to set send deadline: %v", err)
	}
	defer func() {
		if err := s.sock.SetDeadline(time.Time{}); err != nil {
			log.Printf("unable to set send deadline: %v", err)
		}
	}()

	_, err := fmt.Fprintf(s.sock, "%s\r\n", strings.Join(cmd, " "))
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			log.Printf("livesplit write timeout: %v", err)
			return err
		}

		if reconnectErr := s.establishConnection(); reconnectErr != nil {
			return err
		}

		_, retryErr := fmt.Fprintf(s.sock, "%s\r\n", strings.Join(cmd, " "))
		err = retryErr
	}
	return err
}

func (s *socket) recv() (string, error) {
	if s.sock == nil {
		if err := s.establishConnection(); err != nil {
			return "", err
		}
	}

	if err := s.sock.SetDeadline(time.Now().Add(timeout)); err != nil {
		log.Printf("unable to set send deadline: %v", err)
	}
	defer func() {
		if err := s.sock.SetDeadline(time.Time{}); err != nil {
			log.Printf("unable to set send deadline: %v", err)
		}
	}()

	r, err := bufio.NewReader(s.sock).ReadString('\n')
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			log.Printf("livesplit read timeout: %v", err)
			return "", err
		}

		if reconnectErr := s.establishConnection(); reconnectErr != nil {
			return "", err
		}

		retryR, retryErr := bufio.NewReader(s.sock).ReadString('\n')
		if retryErr != nil {
			return "", retryErr
		}

		r = retryR
	}

	ret := strings.TrimSuffix(r, "\r\n")
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
