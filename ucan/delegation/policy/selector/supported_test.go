package selector_test

import (
	"bytes"
	_ "embed"
	"os"
	"reflect"
	"testing"

	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/ucan/delegation/policy/selector"
	sdm "github.com/alanshaw/ucantone/ucan/delegation/policy/selector/internal/fixtures/datamodel"
	"github.com/stretchr/testify/require"
)

// TestSupportedForms runs tests against the Selector according to the
// proposed "Supported Forms" presented in this GitHub issue:
// https://github.com/ucan-wg/delegation/issues/5#issue-2154766496
func TestSupportedForms(t *testing.T) {
	fixturesFile, err := os.Open("./internal/fixtures/supported.json")
	require.NoError(t, err)

	var fixtures sdm.FixturesModel
	err = fixtures.UnmarshalDagJSON(fixturesFile)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		for _, testcase := range fixtures.Success {
			t.Run(testcase.Name, func(t *testing.T) {
				t.Logf("Input: %s\n", string(testcase.Input.Raw))

				sel, err := selector.Parse(testcase.Selector)
				require.NoError(t, err)
				t.Logf("Selector: %s\n", sel)

				in := datamodel.Any{}
				err = in.UnmarshalDagJSON(bytes.NewReader(testcase.Input.Raw))
				require.NoError(t, err)

				one, many, err := selector.Select(sel, in.Value)
				require.NoError(t, err)

				if testcase.Output != nil {
					t.Logf("Expected Output: %s\n", string(testcase.Output.Raw))
					expectOut := datamodel.Any{}
					err = expectOut.UnmarshalDagJSON(bytes.NewReader(testcase.Output.Raw))
					require.NoError(t, err)
					out := datamodel.Any{Value: one}
					expectOutType := reflect.TypeOf(expectOut.Value)
					if expectOutType != nil && expectOutType.Kind() == reflect.Slice {
						out.Value = many
					}

					switch testcase.Name {
					// special case for "Bytes Index" test - selecting a byte from []byte
					// will yield a uint8, which is not supported by the IPLD datamodel
					case "Bytes Index":
						out.Value = int(one.(uint8))
					// special case for "Bytes Slice" test - selecting a slice from []byte
					// will yield a []any, which needs to be converted back to []byte
					case "Bytes Slice":
						var bytes []byte
						for _, v := range many {
							bytes = append(bytes, v.(uint8))
						}
						out.Value = bytes
					}
					var buf bytes.Buffer
					err = out.MarshalDagJSON(&buf)
					require.NoError(t, err)
					t.Logf("Actual Output: %s\n", buf.String())
					require.Equal(t, string(testcase.Output.Raw), buf.String())
				}
			})
		}
	})

	t.Run("null", func(t *testing.T) {
		for _, testcase := range fixtures.Null {
			t.Run(testcase.Name, func(t *testing.T) {
				t.Logf("Input: %s\n", string(testcase.Input.Raw))

				sel, err := selector.Parse(testcase.Selector)
				require.NoError(t, err)
				t.Logf("Selector: %s\n", sel)

				in := datamodel.Any{}
				err = in.UnmarshalDagJSON(bytes.NewReader(testcase.Input.Raw))
				require.NoError(t, err)

				one, many, err := selector.Select(sel, in.Value)
				require.NoError(t, err)
				require.Nil(t, one)
				require.Empty(t, many)
			})
		}
	})

	t.Run("error", func(t *testing.T) {
		for _, testcase := range fixtures.Error {
			t.Run(testcase.Name, func(t *testing.T) {
				t.Logf("Input: %s\n", string(testcase.Input.Raw))

				sel, err := selector.Parse(testcase.Selector)
				require.NoError(t, err)
				t.Logf("Selector: %s\n", sel)

				in := datamodel.Any{}
				err = in.UnmarshalDagJSON(bytes.NewReader(testcase.Input.Raw))
				require.NoError(t, err)

				one, many, err := selector.Select(sel, in.Value)
				require.Error(t, err)
				t.Logf("Error: %v\n", err)
				require.Nil(t, one)
				require.Empty(t, many)
			})
		}
	})
}
