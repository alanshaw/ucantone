package datamodel_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/testing/helpers"
	"github.com/stretchr/testify/require"
)

func TestAny(t *testing.T) {
	values := []any{
		int64(138),
		true,
		false,
		nil,
		helpers.RandomCID(t),
		"test",
		[]byte{1, 2, 3},
		[]string{"one", "two", "three"},
	}

	for _, v := range values {
		t.Run(fmt.Sprintf("%T", v), func(t *testing.T) {
			initial := datamodel.New(v)

			var buf bytes.Buffer
			err := initial.MarshalCBOR(&buf)
			require.NoError(t, err)

			var decoded datamodel.Any
			err = decoded.UnmarshalCBOR(&buf)
			require.NoError(t, err)
			require.Equal(t, v, decoded.Value)
		})
	}
}
