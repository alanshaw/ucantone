package datamodel_test

import (
	"bytes"
	"slices"
	"testing"

	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/testing/helpers"
	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	t.Run("get", func(t *testing.T) {
		obj := helpers.TestObject{Bytes: []byte{1, 2, 3}}
		initial, err := datamodel.NewMapFromCBORMarshaler(&obj)
		require.NoError(t, err)

		var buf bytes.Buffer
		err = initial.MarshalCBOR(&buf)
		require.NoError(t, err)

		var decoded datamodel.Map
		err = decoded.UnmarshalCBOR(&buf)
		require.NoError(t, err)

		value, ok := decoded.Value("bytes")
		require.True(t, ok)
		require.Equal(t, obj.Bytes, value)
	})

	t.Run("keys", func(t *testing.T) {
		obj := helpers.TestObject{Bytes: []byte{1, 2, 3}}
		initial, err := datamodel.NewMapFromCBORMarshaler(&obj)
		require.NoError(t, err)

		var buf bytes.Buffer
		err = initial.MarshalCBOR(&buf)
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
