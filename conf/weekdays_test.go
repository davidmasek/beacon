package conf

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestWeekdaysParse(t *testing.T) {
	monday := time.Date(2025, 4, 28, 12, 0, 0, 0, time.UTC)
	// sanity check for test setup
	require.Equal(t, "Monday", monday.Weekday().String())
	tuesday := monday.AddDate(0, 0, 1)
	wednesday := monday.AddDate(0, 0, 2)
	thursday := monday.AddDate(0, 0, 3)
	friday := monday.AddDate(0, 0, 4)
	saturday := monday.AddDate(0, 0, 5)
	sunday := monday.AddDate(0, 0, 6)

	src := `
Mon Tue Sun
`
	weekdays := WeekdaysSet{}
	err := yaml.Unmarshal([]byte(src), &weekdays)
	require.NoError(t, err)
	t.Logf("%#v\n", weekdays)

	require.True(t, weekdays.Contains(monday))
	require.True(t, weekdays.Contains(tuesday))
	require.False(t, weekdays.Contains(wednesday))
	require.False(t, weekdays.Contains(thursday))
	require.False(t, weekdays.Contains(friday))
	require.False(t, weekdays.Contains(saturday))
	require.True(t, weekdays.Contains(sunday))

	src = `
Wednesday Thursday Saturday
	`
	err = yaml.Unmarshal([]byte(src), &weekdays)
	require.NoError(t, err)
	t.Logf("%#v\n", weekdays)

	require.False(t, weekdays.Contains(monday))
	require.False(t, weekdays.Contains(tuesday))
	require.True(t, weekdays.Contains(wednesday))
	require.True(t, weekdays.Contains(thursday))
	require.False(t, weekdays.Contains(friday))
	require.True(t, weekdays.Contains(saturday))
	require.False(t, weekdays.Contains(sunday))
}

func TestWeekdaysMarshall(t *testing.T) {
	src := `
Mon Tue Sun
`
	weekdays := WeekdaysSet{}
	err := yaml.Unmarshal([]byte(src), &weekdays)
	require.NoError(t, err)
	t.Logf("%#v\n", weekdays)

	out, err := weekdays.MarshalYAML()
	require.NoError(t, err)
	require.Equal(t, "Sunday Monday Tuesday", out)
}
