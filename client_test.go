package livesplit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDurationAndTimeParsing(t *testing.T) {
	testCases := []struct {
		time     string
		duration time.Duration
	}{
		{time: "0", duration: time.Duration(0)},
		{time: "1", duration: time.Duration(1) * time.Second},
		{time: "2", duration: time.Duration(2) * time.Second},
		{time: "01:00", duration: time.Duration(1) * time.Minute},
		{time: "02:00", duration: time.Duration(2) * time.Minute},
		{time: "01:00:00", duration: time.Duration(1) * time.Hour},
		{time: "02:00:00", duration: time.Duration(2) * time.Hour},
		{time: "02:01:01", duration: time.Duration(2)*time.Hour + time.Duration(1)*time.Minute + time.Duration(1)*time.Second},
		{time: "0.001", duration: time.Duration(1) * time.Millisecond},

		{time: "-1", duration: time.Duration(-1) * time.Second},
		{time: "-2", duration: time.Duration(-2) * time.Second},
		{time: "-01:00", duration: time.Duration(-1) * time.Minute},
		{time: "-02:00", duration: time.Duration(-2) * time.Minute},
		{time: "-01:00:00", duration: time.Duration(-1) * time.Hour},
		{time: "-02:00:00", duration: time.Duration(-2) * time.Hour},
		{time: "-02:01:01", duration: time.Duration(-2)*time.Hour + time.Duration(-1)*time.Minute + time.Duration(-1)*time.Second},
		{time: "-0.001", duration: time.Duration(-1) * time.Millisecond},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.time, func(t *testing.T) {
			t.Parallel()

			actual, err := StringToDuration(tC.time)
			require.NoError(t, err)
			assert.Equal(t, tC.duration, actual)
		})
		t.Run(tC.time, func(t *testing.T) {
			t.Parallel()

			actual := DurationToString(tC.duration)
			assert.Equal(t, tC.time, actual)
		})
	}
}
