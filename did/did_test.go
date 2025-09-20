package did

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Run("did:key", func(t *testing.T) {
		str := "did:key:z6Mkod5Jr3yd5SC7UDueqK4dAAw5xYJYjksy722tA9Boxc4z"
		d, err := Parse(str)
		require.NoError(t, err)
		require.Equal(t, str, d.String())
	})

	t.Run("did:web", func(t *testing.T) {
		str := "did:web:up.storacha.network"
		d, err := Parse(str)
		require.NoError(t, err)
		require.Equal(t, str, d.String())
	})
}

func TestEncodeDecode(t *testing.T) {
	t.Run("did:key", func(t *testing.T) {
		str := "did:key:z6Mkod5Jr3yd5SC7UDueqK4dAAw5xYJYjksy722tA9Boxc4z"
		d0, err := Parse(str)
		require.NoError(t, err)
		bytes, err := Encode(d0)
		require.NoError(t, err)
		d1, err := Decode(bytes)
		require.NoError(t, err)
		require.Equal(t, str, d1.String())
	})

	t.Run("did:web", func(t *testing.T) {
		str := "did:web:up.storacha.network"
		d0, err := Parse(str)
		require.NoError(t, err)
		bytes, err := Encode(d0)
		require.NoError(t, err)
		d1, err := Decode(bytes)
		require.NoError(t, err)
		require.Equal(t, str, d1.String())
	})
}

func TestEquivalence(t *testing.T) {
	d0, err := Parse("did:key:z6Mkod5Jr3yd5SC7UDueqK4dAAw5xYJYjksy722tA9Boxc4z")
	require.NoError(t, err)

	d1, err := Parse("did:key:z6Mkod5Jr3yd5SC7UDueqK4dAAw5xYJYjksy722tA9Boxc4z")
	require.NoError(t, err)

	if d0 != d1 {
		require.Fail(t, "DIDs were not equal")
	}

	require.Equal(t, d0, d1)
}

func TestMapKey(t *testing.T) {
	d0, err := Parse("did:key:z6Mkod5Jr3yd5SC7UDueqK4dAAw5xYJYjksy722tA9Boxc4z")
	require.NoError(t, err)

	d1, err := Parse("did:key:z6Mkod5Jr3yd5SC7UDueqK4dAAw5xYJYjksy722tA9Boxc4z")
	require.NoError(t, err)

	m := map[DID]string{}
	m[d0] = "test"
	require.Equal(t, "test", m[d1])
}

func TestRoundtripJSON(t *testing.T) {
	id, err := Parse("did:key:z6Mkod5Jr3yd5SC7UDueqK4dAAw5xYJYjksy722tA9Boxc4z")
	require.NoError(t, err)

	type Object struct {
		ID                DID  `json:"id"`
		UndefID           DID  `json:"undef_id"`
		OptionalPresentID *DID `json:"optional_present_id"`
		OptionalAbsentID  *DID `json:"optional_absent_id"`
	}

	var undef DID
	obj := Object{
		ID:                id,
		UndefID:           undef,
		OptionalPresentID: &id,
		OptionalAbsentID:  nil,
	}

	data, err := json.Marshal(obj)
	require.NoError(t, err)

	t.Log(string(data))

	var out Object
	err = json.Unmarshal(data, &out)
	require.NoError(t, err)

	require.Equal(t, obj.ID, out.ID)
	require.Equal(t, obj.UndefID, out.UndefID)
	require.Equal(t, obj.OptionalPresentID.String(), out.OptionalPresentID.String())
	require.Nil(t, out.OptionalAbsentID)
}
