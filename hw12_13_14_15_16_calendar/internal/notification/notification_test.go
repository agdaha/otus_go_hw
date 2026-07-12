package notification

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMarshalUnmarshalRoundTrip(t *testing.T) {
	n := Notification{
		EventID: "1",
		Title:   "t",
		EventAt: time.Now().Truncate(time.Second).UTC(),
		UserID:  "u1",
	}

	data, err := n.Marshal()
	require.NoError(t, err)

	back, err := Unmarshal(data)
	require.NoError(t, err)
	require.Equal(t, n, back)
}

func TestUnmarshalInvalid(t *testing.T) {
	_, err := Unmarshal([]byte("not-json"))
	require.Error(t, err)
}
