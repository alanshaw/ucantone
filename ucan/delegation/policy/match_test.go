package policy

import (
	"testing"

	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/testing/helpers"
	"github.com/alanshaw/ucantone/ucan/delegation/policy/selector"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"
)

func TestMatch(t *testing.T) {
	t.Run("equality", func(t *testing.T) {
		t.Run("string", func(t *testing.T) {
			val := "test"

			pol := Policy{Equal(selector.MustParse("."), "test")}
			ok := Match(pol, val)
			require.True(t, ok)

			pol = Policy{Equal(selector.MustParse("."), "test2")}
			ok = Match(pol, val)
			require.False(t, ok)

			pol = Policy{Equal(selector.MustParse("."), 138)}
			ok = Match(pol, val)
			require.False(t, ok)
		})

		t.Run("int", func(t *testing.T) {
			val := 138

			pol := Policy{Equal(selector.MustParse("."), 138)}
			ok := Match(pol, val)
			require.True(t, ok)

			pol = Policy{Equal(selector.MustParse("."), 1138)}
			ok = Match(pol, val)
			require.False(t, ok)

			pol = Policy{Equal(selector.MustParse("."), "138")}
			ok = Match(pol, val)
			require.False(t, ok)
		})

		// 	t.Run("float", func(t *testing.T) {
		// 		np := basicnode.Prototype.Float
		// 		nb := np.NewBuilder()
		// 		nb.AssignFloat(1.138)
		// 		nd := nb.Build()

		// 		pol := Policy{Equal(selector.MustParse("."), literal.Float(1.138))}
		// 		ok := Match(pol, nd)
		// 		require.True(t, ok)

		// 		pol = Policy{Equal(selector.MustParse("."), literal.Float(11.38))}
		// 		ok = Match(pol, nd)
		// 		require.False(t, ok)

		// 		pol = Policy{Equal(selector.MustParse("."), literal.String("138"))}
		// 		ok = Match(pol, nd)
		// 		require.False(t, ok)
		// 	})

		t.Run("CID", func(t *testing.T) {
			l0 := cid.MustParse("bafybeif4owy5gno5lwnixqm52rwqfodklf76hsetxdhffuxnplvijskzqq")
			l1 := cid.MustParse("bafkreifau35r7vi37tvbvfy3hdwvgb4tlflqf7zcdzeujqcjk3rsphiwte")
			val := l0

			pol := Policy{Equal(selector.MustParse("."), l0)}
			ok := Match(pol, val)
			require.True(t, ok)

			pol = Policy{Equal(selector.MustParse("."), l1)}
			ok = Match(pol, val)
			require.False(t, ok)

			pol = Policy{Equal(selector.MustParse("."), "bafybeif4owy5gno5lwnixqm52rwqfodklf76hsetxdhffuxnplvijskzqq")}
			ok = Match(pol, val)
			require.False(t, ok)
		})

		t.Run("string in map", func(t *testing.T) {
			val := helpers.Must(datamodel.NewMap(
				datamodel.WithValues(map[string]any{"foo": "bar"}),
			))(t)

			pol := Policy{Equal(selector.MustParse(".foo"), "bar")}
			ok := Match(pol, val)
			require.True(t, ok)

			pol = Policy{Equal(selector.MustParse(`.["foo"]`), "bar")}
			ok = Match(pol, val)
			require.True(t, ok)

			pol = Policy{Equal(selector.MustParse(".foo"), "baz")}
			ok = Match(pol, val)
			require.False(t, ok)

			pol = Policy{Equal(selector.MustParse(".foobar"), "bar")}
			ok = Match(pol, val)
			require.False(t, ok)
		})

		t.Run("string in list", func(t *testing.T) {
			val := []string{"foo"}

			pol := Policy{Equal(selector.MustParse(".[0]"), "foo")}
			ok := Match(pol, val)
			require.True(t, ok)

			pol = Policy{Equal(selector.MustParse(".[1]"), "foo")}
			ok = Match(pol, val)
			require.False(t, ok)
		})
	})

	t.Run("inequality", func(t *testing.T) {
		t.Run("gt int", func(t *testing.T) {
			val := 138
			pol := Policy{GreaterThan(selector.MustParse("."), 1)}
			ok := Match(pol, val)
			require.True(t, ok)
		})

		t.Run("gte int", func(t *testing.T) {
			val := 138

			pol := Policy{GreaterThanOrEqual(selector.MustParse("."), 1)}
			ok := Match(pol, val)
			require.True(t, ok)

			pol = Policy{GreaterThanOrEqual(selector.MustParse("."), 138)}
			ok = Match(pol, val)
			require.True(t, ok)
		})

		// t.Run("gt float", func(t *testing.T) {
		// 	np := basicnode.Prototype.Float
		// 	nb := np.NewBuilder()
		// 	nb.AssignFloat(1.38)
		// 	nd := nb.Build()

		// 	pol := Policy{GreaterThan(selector.MustParse("."), literal.Float(1))}
		// 	ok := Match(pol, nd)
		// 	require.True(t, ok)
		// })

		// t.Run("gte float", func(t *testing.T) {
		// 	np := basicnode.Prototype.Float
		// 	nb := np.NewBuilder()
		// 	nb.AssignFloat(1.38)
		// 	nd := nb.Build()

		// 	pol := Policy{GreaterThanOrEqual(selector.MustParse("."), literal.Float(1))}
		// 	ok := Match(pol, nd)
		// 	require.True(t, ok)

		// 	pol = Policy{GreaterThanOrEqual(selector.MustParse("."), literal.Float(1.38))}
		// 	ok = Match(pol, nd)
		// 	require.True(t, ok)
		// })

		t.Run("lt int", func(t *testing.T) {
			val := 138
			pol := Policy{LessThan(selector.MustParse("."), 1138)}
			ok := Match(pol, val)
			require.True(t, ok)
		})

		t.Run("lte int", func(t *testing.T) {
			val := 138

			pol := Policy{LessThanOrEqual(selector.MustParse("."), 1138)}
			ok := Match(pol, val)
			require.True(t, ok)

			pol = Policy{LessThanOrEqual(selector.MustParse("."), 138)}
			ok = Match(pol, val)
			require.True(t, ok)
		})
	})

	// t.Run("negation", func(t *testing.T) {
	// 	np := basicnode.Prototype.Bool
	// 	nb := np.NewBuilder()
	// 	nb.AssignBool(false)
	// 	nd := nb.Build()

	// 	pol := Policy{Not(Equal(selector.MustParse("."), literal.Bool(true)))}
	// 	ok := Match(pol, nd)
	// 	require.True(t, ok)

	// 	pol = Policy{Not(Equal(selector.MustParse("."), literal.Bool(false)))}
	// 	ok = Match(pol, nd)
	// 	require.False(t, ok)
	// })

	// t.Run("conjunction", func(t *testing.T) {
	// 	np := basicnode.Prototype.Int
	// 	nb := np.NewBuilder()
	// 	nb.AssignInt(138)
	// 	nd := nb.Build()

	// 	pol := Policy{
	// 		And(
	// 			GreaterThan(selector.MustParse("."), literal.Int(1)),
	// 			LessThan(selector.MustParse("."), literal.Int(1138)),
	// 		),
	// 	}
	// 	ok := Match(pol, nd)
	// 	require.True(t, ok)

	// 	pol = Policy{
	// 		And(
	// 			GreaterThan(selector.MustParse("."), literal.Int(1)),
	// 			Equal(selector.MustParse("."), literal.Int(1138)),
	// 		),
	// 	}
	// 	ok = Match(pol, nd)
	// 	require.False(t, ok)

	// 	pol = Policy{And()}
	// 	ok = Match(pol, nd)
	// 	require.True(t, ok)
	// })

	// t.Run("disjunction", func(t *testing.T) {
	// 	np := basicnode.Prototype.Int
	// 	nb := np.NewBuilder()
	// 	nb.AssignInt(138)
	// 	nd := nb.Build()

	// 	pol := Policy{
	// 		Or(
	// 			GreaterThan(selector.MustParse("."), literal.Int(138)),
	// 			LessThan(selector.MustParse("."), literal.Int(1138)),
	// 		),
	// 	}
	// 	ok := Match(pol, nd)
	// 	require.True(t, ok)

	// 	pol = Policy{
	// 		Or(
	// 			GreaterThan(selector.MustParse("."), literal.Int(138)),
	// 			Equal(selector.MustParse("."), literal.Int(1138)),
	// 		),
	// 	}
	// 	ok = Match(pol, nd)
	// 	require.False(t, ok)

	// 	pol = Policy{Or()}
	// 	ok = Match(pol, nd)
	// 	require.True(t, ok)
	// })

	// t.Run("wildcard", func(t *testing.T) {
	// 	glb, err := glob.Compile(`Alice\*, Bob*, Carol.`)
	// 	require.NoError(t, err)

	// 	for _, s := range []string{
	// 		"Alice*, Bob, Carol.",
	// 		"Alice*, Bob, Dan, Erin, Carol.",
	// 		"Alice*, Bob  , Carol.",
	// 		"Alice*, Bob*, Carol.",
	// 	} {
	// 		func(s string) {
	// 			t.Run(fmt.Sprintf("pass %s", s), func(t *testing.T) {
	// 				np := basicnode.Prototype.String
	// 				nb := np.NewBuilder()
	// 				nb.AssignString(s)
	// 				nd := nb.Build()

	// 				pol := Policy{Like(selector.MustParse("."), glb)}
	// 				ok := Match(pol, nd)
	// 				require.True(t, ok)
	// 			})
	// 		}(s)
	// 	}

	// 	for _, s := range []string{
	// 		"Alice*, Bob, Carol",
	// 		"Alice*, Bob*, Carol!",
	// 		"Alice, Bob, Carol.",
	// 		" Alice*, Bob, Carol. ",
	// 	} {
	// 		func(s string) {
	// 			t.Run(fmt.Sprintf("fail %s", s), func(t *testing.T) {
	// 				np := basicnode.Prototype.String
	// 				nb := np.NewBuilder()
	// 				nb.AssignString(s)
	// 				nd := nb.Build()

	// 				pol := Policy{Like(selector.MustParse("."), glb)}
	// 				ok := Match(pol, nd)
	// 				require.False(t, ok)
	// 			})
	// 		}(s)
	// 	}
	// })

	// t.Run("quantification", func(t *testing.T) {
	// 	buildValueNode := func(v int64) ipld.Node {
	// 		np := basicnode.Prototype.Map
	// 		nb := np.NewBuilder()
	// 		ma, _ := nb.BeginMap(1)
	// 		ma.AssembleKey().AssignString("value")
	// 		ma.AssembleValue().AssignInt(v)
	// 		ma.Finish()
	// 		return nb.Build()
	// 	}

	// 	t.Run("all", func(t *testing.T) {
	// 		np := basicnode.Prototype.List
	// 		nb := np.NewBuilder()
	// 		la, _ := nb.BeginList(5)
	// 		la.AssembleValue().AssignNode(buildValueNode(5))
	// 		la.AssembleValue().AssignNode(buildValueNode(10))
	// 		la.AssembleValue().AssignNode(buildValueNode(20))
	// 		la.AssembleValue().AssignNode(buildValueNode(50))
	// 		la.AssembleValue().AssignNode(buildValueNode(100))
	// 		la.Finish()
	// 		nd := nb.Build()

	// 		pol := Policy{
	// 			All(
	// 				selector.MustParse(".[]"),
	// 				GreaterThan(selector.MustParse(".value"), literal.Int(2)),
	// 			),
	// 		}
	// 		ok := Match(pol, nd)
	// 		require.True(t, ok)

	// 		pol = Policy{
	// 			All(
	// 				selector.MustParse(".[]"),
	// 				GreaterThan(selector.MustParse(".value"), literal.Int(20)),
	// 			),
	// 		}
	// 		ok = Match(pol, nd)
	// 		require.False(t, ok)
	// 	})

	// 	t.Run("any", func(t *testing.T) {
	// 		np := basicnode.Prototype.List
	// 		nb := np.NewBuilder()
	// 		la, _ := nb.BeginList(5)
	// 		la.AssembleValue().AssignNode(buildValueNode(5))
	// 		la.AssembleValue().AssignNode(buildValueNode(10))
	// 		la.AssembleValue().AssignNode(buildValueNode(20))
	// 		la.AssembleValue().AssignNode(buildValueNode(50))
	// 		la.AssembleValue().AssignNode(buildValueNode(100))
	// 		la.Finish()
	// 		nd := nb.Build()

	// 		pol := Policy{
	// 			Any(
	// 				selector.MustParse(".[]"),
	// 				GreaterThan(selector.MustParse(".value"), literal.Int(10)),
	// 				LessThan(selector.MustParse(".value"), literal.Int(50)),
	// 			),
	// 		}
	// 		ok := Match(pol, nd)
	// 		require.True(t, ok)

	// 		pol = Policy{
	// 			Any(
	// 				selector.MustParse(".[]"),
	// 				GreaterThan(selector.MustParse(".value"), literal.Int(100)),
	// 			),
	// 		}
	// 		ok = Match(pol, nd)
	// 		require.False(t, ok)
	// 	})
	// })
}
