package livesplit

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

type TimerPhase string

const (
	NotRunning TimerPhase = "NotRunning"
	Ended      TimerPhase = "Ended"
	Running    TimerPhase = "Running"
	Paused     TimerPhase = "Paused"
)

const (
	close                  = "close"
	starttimer             = "starttimer"
	startorsplit           = "startorsplit"
	split                  = "split"
	unsplit                = "unsplit"
	skipsplit              = "skipsplit"
	pause                  = "pause"
	resume                 = "resume"
	reset                  = "reset"
	initgametime           = "initgametime"
	setgametime            = "setgametime"
	setloadingtimes        = "setloadingtimes"
	pausegametime          = "pausegametime"
	unpausegametime        = "unpausegametime"
	setcomparison          = "setcomparison"
	getdelta               = "getdelta"
	getlastsplittime       = "getlastsplittime"
	getcomparisonsplittime = "getcomparisonsplittime"
	getcurrenttime         = "getcurrenttime"
	getfinaltime           = "getfinaltime"
	getpredictedtime       = "getpredictedtime"
	getbestpossibletime    = "getbestpossibletime"
	getsplitindex          = "getsplitindex"
	getcurrentsplitname    = "getcurrentsplitname"
	getprevioussplitname   = "getprevioussplitname"
	getcurrenttimerphase   = "getcurrenttimerphase"
)

type Client struct {
	m         sync.Mutex
	sock      *socket
	callbacks map[string][]func(cmd []string) error
}

func NewClient() *Client {
	return NewClientWithPort(port)
}

func NewClientWithPort(port int) *Client {
	return &Client{sock: newSocket(port), callbacks: make(map[string][]func(cmd []string) error)}
}

func (client *Client) OnClose(callback func(cmd []string) error) {
	client.registerCallback(close, callback)
}

func (client *Client) Close() error {
	if client.sock != nil {
		if err := client.sock.Close(); err != nil {
			return err
		}
	}
	client.callCallbacks([]string{close})
	return nil
}

func (client *Client) OnStartTimer(callback func(cmd []string) error) {
	client.registerCallback(starttimer, callback)
}

func (client *Client) StartTimer() error {
	return client.cmd(starttimer)
}

func (client *Client) StartOrSplit() error {
	phase, err := client.GetCurrentTimerPhase()
	if err != nil {
		return err
	}

	if err := client.cmd(startorsplit); err != nil {
		return err
	}

	if phase == Running {
		client.callCallbacks([]string{starttimer})
	} else {
		client.callCallbacks([]string{split})
	}

	return nil
}

func (client *Client) OnSplit(callback func(cmd []string) error) {
	client.registerCallback(split, callback)
}

func (client *Client) Split() error {
	return client.cmd(split)
}

func (client *Client) OnUnsplit(callback func(cmd []string) error) {
	client.registerCallback(unsplit, callback)
}

func (client *Client) Unsplit() error {
	return client.cmd(unsplit)
}

func (client *Client) OnSkipSplit(callback func(cmd []string) error) {
	client.registerCallback(skipsplit, callback)
}

func (client *Client) SkipSplit() error {
	return client.cmd(skipsplit)
}

func (client *Client) OnPause(callback func(cmd []string) error) {
	client.registerCallback(pause, callback)
}

func (client *Client) Pause() error {
	return client.cmd(pause)
}

func (client *Client) OnResume(callback func(cmd []string) error) {
	client.registerCallback(resume, callback)
}

func (client *Client) Resume() error {
	return client.cmd(resume)
}

func (client *Client) OnReset(callback func(cmd []string) error) {
	client.registerCallback(reset, callback)
}

func (client *Client) Reset() error {
	return client.cmd(reset)
}

func (client *Client) OnInitGameTime(callback func(cmd []string) error) {
	client.registerCallback(initgametime, callback)
}

func (client *Client) InitGameTime() error {
	return client.cmd(initgametime)
}

func (client *Client) OnSetGameTime(callback func(cmd []string) error) {
	client.registerCallback(setgametime, callback)
}

func (client *Client) SetGameTime(t time.Duration) error {
	return client.cmd(setgametime, DurationToString(t))
}

func (client *Client) OnSetLoadingTimes(callback func(cmd []string) error) {
	client.registerCallback(setloadingtimes, callback)
}

func (client *Client) SetLoadingTimes(t time.Duration) error {
	return client.cmd(setloadingtimes, DurationToString(t))
}

func (client *Client) OnPauseGameTime(callback func(cmd []string) error) {
	client.registerCallback(pausegametime, callback)
}

func (client *Client) PauseGameTime() error {
	return client.cmd(pausegametime)
}

func (client *Client) OnUnpauseGameTime(callback func(cmd []string) error) {
	client.registerCallback(unpausegametime, callback)
}

func (client *Client) UnpauseGameTime() error {
	return client.cmd(unpausegametime)
}

func (client *Client) OnSetComparison(callback func(cmd []string) error) {
	client.registerCallback(setcomparison, callback)
}

func (client *Client) SetComparison(comparison string) error {
	return client.cmd(setcomparison, comparison)
}

func (client *Client) OnGetDelta(callback func(cmd []string) error) {
	client.registerCallback(getdelta, callback)
}

func (client *Client) GetDelta(comparison string) (string, error) {
	return client.cmdWithResult(getdelta, comparison)
}

func (client *Client) OnGetLastSplitTime(callback func(cmd []string) error) {
	client.registerCallback(getlastsplittime, callback)
}

func (client *Client) GetLastSplitTime() (time.Duration, error) {
	return client.cmdWithResultTime(getlastsplittime)
}

func (client *Client) OnGetComparisonSplitTime(callback func(cmd []string) error) {
	client.registerCallback(getcomparisonsplittime, callback)
}

func (client *Client) GetComparisonSplitTime() (time.Duration, error) {
	return client.cmdWithResultTime(getcomparisonsplittime)
}

func (client *Client) OnGetCurrentTime(callback func(cmd []string) error) {
	client.registerCallback(getcurrenttime, callback)
}

func (client *Client) GetCurrentTime() (time.Duration, error) {
	return client.cmdWithResultTime(getcurrenttime)
}

func (client *Client) OnGetFinalTime(callback func(cmd []string) error) {
	client.registerCallback(getfinaltime, callback)
}

func (client *Client) GetFinalTime(comparison string) (time.Duration, error) {
	return client.cmdWithResultTime(getfinaltime)
}

func (client *Client) OnGetPredicatedTime(callback func(cmd []string) error) {
	client.registerCallback(getpredictedtime, callback)
}

func (client *Client) GetPredictedTime(comparison string) (time.Duration, error) {
	return client.cmdWithResultTime(getpredictedtime)
}

func (client *Client) OnGetBestPossibleTime(callback func(cmd []string) error) {
	client.registerCallback(getbestpossibletime, callback)
}

func (client *Client) GetBestPossibleTime() (time.Duration, error) {
	return client.cmdWithResultTime(getbestpossibletime)
}

func (client *Client) OnGetSplitIndex(callback func(cmd []string) error) {
	client.registerCallback(getsplitindex, callback)
}

func (client *Client) GetSplitIndex() (int, error) {
	return client.cmdWithResultInt(getsplitindex)
}

func (client *Client) OnGetCurrentSplitName(callback func(cmd []string) error) {
	client.registerCallback(getcurrentsplitname, callback)
}

func (client *Client) GetCurrentSplitName() (string, error) {
	return client.cmdWithResult(getcurrentsplitname)
}

func (client *Client) OnGetPreviousSplitName(callback func(cmd []string) error) {
	client.registerCallback(getprevioussplitname, callback)
}

func (client *Client) GetPreviousSplitName() (string, error) {
	return client.cmdWithResult(getprevioussplitname)
}

func (client *Client) OnGetCurrentTimerPhase(callback func(cmd []string) error) {
	client.registerCallback(getcurrenttimerphase, callback)
}

func (client *Client) GetCurrentTimerPhase() (TimerPhase, error) {
	s, err := client.cmdWithResult(getcurrenttimerphase)
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

func (client *Client) cmd(cmd ...string) error {
	client.m.Lock()
	defer client.m.Unlock()
	if err := client.sock.send(cmd); err != nil {
		return err
	}
	client.callCallbacks(cmd)
	return nil
}

func (client *Client) cmdWithResult(cmd ...string) (string, error) {
	client.m.Lock()
	defer client.m.Unlock()
	s, err := client.sock.sendAndRecv(cmd)
	if err != nil {
		return "", err
	}
	client.callCallbacks(cmd)
	return s, nil
}

func (client *Client) cmdWithResultInt(cmd ...string) (int, error) {
	s, err := client.cmdWithResult(cmd...)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(s)
}

func (client *Client) cmdWithResultTime(cmd ...string) (time.Duration, error) {
	s, err := client.cmdWithResult(cmd...)
	if err != nil {
		return 0, err
	}
	return StringToDuration(s)
}

func (client *Client) registerCallback(name string, callback func(cmd []string) error) {
	client.callbacks[name] = append(client.callbacks[name], callback)
}

func (client *Client) callCallbacks(cmd []string) {
	for _, callback := range client.callbacks[cmd[0]] {
		if err := callback(cmd); err != nil {
			log.Printf("error in callback for %s: %v", cmd[0], err)
		}
	}
}
