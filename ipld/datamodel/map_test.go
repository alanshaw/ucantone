package datamodel_test

import (
	"bytes"
	"slices"
	"testing"

	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	t.Run("get", func(t *testing.T) {
		bytesValue := []byte{1, 2, 3}
		initial := datamodel.NewMap(datamodel.WithEntry("bytes", bytesValue))

		var buf bytes.Buffer
		err := initial.MarshalCBOR(&buf)
		require.NoError(t, err)

		var decoded datamodel.Map
		err = decoded.UnmarshalCBOR(&buf)
		require.NoError(t, err)

		value, ok := decoded.Get("bytes")
		require.True(t, ok)
		require.Equal(t, bytesValue, value)
	})

	t.Run("keys", func(t *testing.T) {
		initial := datamodel.NewMap(datamodel.WithEntry("bytes", []byte{1, 2, 3}))

		var buf bytes.Buffer
		err := initial.MarshalCBOR(&buf)
		require.NoError(t, err)

		var decoded datamodel.Map
		err = decoded.UnmarshalCBOR(&buf)
		require.NoError(t, err)
		require.Equal(t, []string{"bytes"}, slices.Collect(decoded.Keys()))
	})

	t.Run("empty", func(t *testing.T) {
		var initial datamodel.Map

		var buf bytes.Buffer
		err := initial.MarshalCBOR(&buf)
		require.NoError(t, err)

		var decoded datamodel.Map
		err = decoded.UnmarshalCBOR(&buf)
		require.NoError(t, err)
		require.Len(t, slices.Collect(decoded.Keys()), 0)
	})
}
