package livesplit

import (
	"bufio"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	port    = 16834
	timeout = time.Duration(2) * time.Second
)

var (
	m sync.Mutex
)

type TimerPhase string

const (
	NotRunning TimerPhase = "NotRunning"
	Ended      TimerPhase = "Ended"
	Running    TimerPhase = "Running"
	Paused     TimerPhase = "Paused"
)

type Client struct {
	port int
	sock net.Conn
}

func NewClient() *Client {
	return NewClientWithPort(port)
}

func NewClientWithPort(port int) *Client {
	return &Client{port: port}
}

func (client *Client) establishConnection() error {
	if client.sock != nil {
		client.sock.Close()
	}

	sock, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), timeout)
	if err != nil {
		return err
	}
	client.sock = sock
	return nil
}

func (client *Client) Close() error {
	if client.sock != nil {
		return client.sock.Close()
	}
	return nil
}

func (client *Client) StartTimer() error {
	return client.send([]string{"starttimer"})
}

func (client *Client) StartOrSplit() error {
	return client.send([]string{"startorsplit"})
}

func (client *Client) Split() error {
	return client.send([]string{"split"})
}

func (client *Client) Unsplit() error {
	return client.send([]string{"unsplit"})
}

func (client *Client) SkipSplit() error {
	return client.send([]string{"skipsplit"})
}

func (client *Client) Pause() error {
	return client.send([]string{"pause"})
}

func (client *Client) Resume() error {
	return client.send([]string{"resume"})
}

func (client *Client) Reset() error {
	return client.send([]string{"reset"})
}

func (client *Client) InitGameTime() error {
	return client.send([]string{"initgametime"})
}

func (client *Client) SetGameTime(t time.Duration) error {
	return client.send([]string{"setgametime", DurationToString(t)})
}

func (client *Client) SetLoadingTimes(t time.Duration) error {
	return client.send([]string{"setloadingtimes", DurationToString(t)})
}

func (client *Client) PauseGameTime() error {
	return client.send([]string{"pausegametime"})
}

func (client *Client) UnpauseGameTime() error {
	return client.send([]string{"unpausegametime"})
}

func (client *Client) SetComparison(comparison string) error {
	return client.send([]string{"setcomparison", comparison})
}

func (client *Client) GetDelta(comparison string) (string, error) {
	if err := client.send([]string{"getdelta", comparison}); err != nil {
		return "", err
	}
	return client.recv()
}

func (client *Client) GetLastSplitTime() (time.Duration, error) {
	return client.sendAndRecvTime([]string{"getlastsplittime"})
}

func (client *Client) GetComparisonSplitTime() (time.Duration, error) {
	return client.sendAndRecvTime([]string{"getcomparisonsplittime"})
}

func (client *Client) GetCurrentTime() (time.Duration, error) {
	return client.sendAndRecvTime([]string{"getcurrenttime"})
}

func (client *Client) GetFinalTime(comparison string) (time.Duration, error) {
	return client.sendAndRecvTime([]string{"getfinaltime", comparison})
}

func (client *Client) GetPredictedTime(comparison string) (time.Duration, error) {
	return client.sendAndRecvTime([]string{"getpredicatedtime", comparison})
}

func (client *Client) GetBestPossibleTime() (time.Duration, error) {
	return client.sendAndRecvTime([]string{"getbestpossibletime"})
}

func (client *Client) GetSplitIndex() (int, error) {
	return client.sendAndRecvInt([]string{"getsplitindex"})
}

func (client *Client) GetCurrentSplitName() (string, error) {
	return client.sendAndRecv([]string{"getcurrentsplitname"})
}

func (client *Client) GetPreviousSplitName() (string, error) {
	return client.sendAndRecv([]string{"getprevioussplitname"})
}

func (client *Client) GetCurrentTimerPhase() (TimerPhase, error) {
	s, err := client.sendAndRecv([]string{"getcurrenttimerphase"})
	if err != nil {
		return "", err
	}
	return TimerPhase(s), nil
}

func DurationToString(t time.Duration) string {
	if math.Abs(t.Seconds()) < 60.0 {
		return strings.TrimSuffix(fmt.Sprintf("%.2f", t.Seconds()), ".00")
	}

	negative := ""
	if t.Seconds() < 0 {
		negative = "-"
		t = -t
	}

	total := int64(t.Seconds())
	total, seconds := total/60, float64(total%60)
	total, minutes := total/60, total%60
	total, hours := total/60, total%60

	seconds += t.Seconds() - float64(int64(t.Seconds()))

	ret := fmt.Sprintf("%02d:%02d:%05.2f", hours, minutes, seconds)
	ret = strings.TrimSuffix(ret, ".00")
	ret = strings.TrimPrefix(ret, "00:")

	return fmt.Sprintf("%s%s", negative, ret)
}

func StringToDuration(t string) (time.Duration, error) {
	factor := 1
	if strings.HasPrefix(t, "-") {
		factor = -1
		t = strings.TrimPrefix(t, "-")
	}

	nanos := float64(0)
	ts := strings.Split(t, ":")

	for i := 0; i < len(ts); i++ {
		part, err := strconv.ParseFloat(ts[i], 64)
		if err != nil {
			return 0, err
		}

		nanos += part * (math.Pow(60.0, float64(len(ts)-1-i))) * 1e9
	}

	return time.Duration(nanos) * time.Nanosecond * time.Duration(factor), nil
}

func (client *Client) send(cmd []string) error {
	m.Lock()
	defer m.Unlock()

	if client.sock == nil {
		if err := client.establishConnection(); err != nil {
			return err
		}
	}

	client.sock.SetDeadline(time.Now().Add(timeout))
	_, err := fmt.Fprintf(client.sock, "%s\r\n", strings.Join(cmd, " "))
	if err != nil {
		if reconnectErr := client.establishConnection(); reconnectErr != nil {
			return err
		}
		_, err := fmt.Fprintf(client.sock, "%s\r\n", strings.Join(cmd, " "))
		if err != nil {
			return err
		}
	}
	client.sock.SetDeadline(time.Time{})
	return err
}

func (client *Client) recv() (string, error) {
	if client.sock == nil {
		if err := client.establishConnection(); err != nil {
			return "", err
		}
	}

	var s string
	var err error

	m.Lock()
	defer m.Unlock()
	client.sock.SetDeadline(time.Now().Add(timeout))
	s, err = bufio.NewReader(client.sock).ReadString('\n')
	if err != nil {
		if reconnectErr := client.establishConnection(); reconnectErr != nil {
			return "", err
		}
		s, err = bufio.NewReader(client.sock).ReadString('\n')
		if err != nil {
			return "", err
		}
	}
	client.sock.SetDeadline(time.Time{})
	return strings.TrimSuffix(s, "\r\n"), nil
}

func (client *Client) sendAndRecv(cmd []string) (string, error) {
	if err := client.send(cmd); err != nil {
		return "", err
	}
	return client.recv()
}

func (client *Client) sendAndRecvTime(cmd []string) (time.Duration, error) {
	s, err := client.sendAndRecv(cmd)
	if err != nil {
		return 0, err
	}
	return StringToDuration(s)
}

func (client *Client) sendAndRecvInt(cmd []string) (int, error) {
	s, err := client.sendAndRecv(cmd)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(s)
}
