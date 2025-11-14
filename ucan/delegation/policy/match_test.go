package policy_test

import (
	"testing"

	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/testutil"
	"github.com/alanshaw/ucantone/ucan/delegation/policy"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"
)

func TestMatch(t *testing.T) {
	t.Run("equality", func(t *testing.T) {
		t.Run("string", func(t *testing.T) {
			pol := testutil.Must(policy.Build(policy.Equal(".", "test")))(t)
			ok, err := policy.Match(pol, "test")
			require.True(t, ok)
			require.NoError(t, err)

			pol = testutil.Must(policy.Build(policy.Equal(".", "test2")))(t)
			ok, err = policy.Match(pol, "test")
			require.False(t, ok)
			require.Error(t, err)
			t.Log(err)

			pol = testutil.Must(policy.Build(policy.Equal(".", 138)))(t)
			ok, err = policy.Match(pol, "test")
			require.False(t, ok)
			require.Error(t, err)
			t.Log(err)
		})

		t.Run("int", func(t *testing.T) {
			pol := testutil.Must(policy.Build(policy.Equal(".", 138)))(t)
			ok, err := policy.Match(pol, 138)
			require.True(t, ok)
			require.NoError(t, err)

			pol = testutil.Must(policy.Build(policy.Equal(".", 1138)))(t)
			ok, err = policy.Match(pol, 138)
			require.False(t, ok)
			require.Error(t, err)
			t.Log(err)

			pol = testutil.Must(policy.Build(policy.Equal(".", "1138")))(t)
			ok, err = policy.Match(pol, 138)
			require.False(t, ok)
			require.Error(t, err)
			t.Log(err)
		})

		// 	t.Run("float", func(t *testing.T) {
		// 		np := basicnode.Prototype.Float
		// 		nb := np.NewBuilder()
		// 		nb.AssignFloat(1.138)
		// 		nd := nb.Build()

		// 		pol := policy.Policy{policy.Equal(mustParse(t, "."), literal.Float(1.138))}
		// 		ok, err := policy.Match(pol, nd)
		// 		require.True(t, ok)

		// 		pol = policy.Policy{policy.Equal(mustParse(t, "."), literal.Float(11.38))}
		// 		ok, err = policy.Match(pol, nd)
		// 		require.False(t, ok)

		// 		pol = policy.Policy{policy.Equal(mustParse(t, "."), literal.String("138"))}
		// 		ok, err = policy.Match(pol, nd)
		// 		require.False(t, ok)
		// 	})

		t.Run("CID", func(t *testing.T) {
			l0 := cid.MustParse("bafybeif4owy5gno5lwnixqm52rwqfodklf76hsetxdhffuxnplvijskzqq")
			l1 := cid.MustParse("bafkreifau35r7vi37tvbvfy3hdwvgb4tlflqf7zcdzeujqcjk3rsphiwte")

			pol := testutil.Must(policy.Build(policy.Equal(".", l0)))(t)
			ok, err := policy.Match(pol, l0)
			require.True(t, ok)
			require.NoError(t, err)

			pol = testutil.Must(policy.Build(policy.Equal(".", l1)))(t)
			ok, err = policy.Match(pol, l0)
			require.False(t, ok)
			require.Error(t, err)
			t.Log(err)

			pol = testutil.Must(policy.Build(policy.Equal(".", "bafybeif4owy5gno5lwnixqm52rwqfodklf76hsetxdhffuxnplvijskzqq")))(t)
			ok, err = policy.Match(pol, l0)
			require.False(t, ok)
			require.Error(t, err)
			t.Log(err)
		})

		t.Run("string in map", func(t *testing.T) {
			val := datamodel.NewMap(datamodel.WithEntry("foo", "bar"))

			pol := testutil.Must(policy.Build(policy.Equal(".foo", "bar")))(t)
			ok, err := policy.Match(pol, val)
			require.True(t, ok)
			require.NoError(t, err)

			pol = testutil.Must(policy.Build(policy.Equal(`.["foo"]`, "bar")))(t)
			ok, err = policy.Match(pol, val)
			require.True(t, ok)
			require.NoError(t, err)

			pol = testutil.Must(policy.Build(policy.Equal(".foo", "baz")))(t)
			ok, err = policy.Match(pol, val)
			require.False(t, ok)
			require.Error(t, err)
			t.Log(err)

			pol = testutil.Must(policy.Build(policy.Equal(".foobar", "bar")))(t)
			ok, err = policy.Match(pol, val)
			require.False(t, ok)
			require.Error(t, err)
			t.Log(err)
		})

		t.Run("string in list", func(t *testing.T) {
			val := []string{"foo"}

			pol := testutil.Must(policy.Build(policy.Equal(".[0]", "foo")))(t)
			ok, err := policy.Match(pol, val)
			require.True(t, ok)
			require.NoError(t, err)

			pol = testutil.Must(policy.Build(policy.Equal(".[1]", "foo")))(t)
			ok, err = policy.Match(pol, val)
			require.False(t, ok)
			require.Error(t, err)
			t.Log(err)
		})
	})

	t.Run("inequality", func(t *testing.T) {
		t.Run("gt int", func(t *testing.T) {
			val := 138
			pol := testutil.Must(policy.Build(policy.GreaterThan(".", 1)))(t)
			ok, err := policy.Match(pol, val)
			require.True(t, ok)
			require.NoError(t, err)
		})

		t.Run("gte int", func(t *testing.T) {
			val := 138

			pol := testutil.Must(policy.Build(policy.GreaterThanOrEqual(".", 1)))(t)
			ok, err := policy.Match(pol, val)
			require.True(t, ok)
			require.NoError(t, err)

			pol = testutil.Must(policy.Build(policy.GreaterThanOrEqual(".", 138)))(t)
			ok, err = policy.Match(pol, val)
			require.True(t, ok)
			require.NoError(t, err)
		})

		// t.Run("gt float", func(t *testing.T) {
		// 	np := basicnode.Prototype.Float
		// 	nb := np.NewBuilder()
		// 	nb.AssignFloat(1.38)
		// 	nd := nb.Build()

		// 	pol := policy.Policy{GreaterThan(mustParse(t, "."), literal.Float(1))}
		// 	ok, err := policy.Match(pol, nd)
		// 	require.True(t, ok)
		// })

		// t.Run("gte float", func(t *testing.T) {
		// 	np := basicnode.Prototype.Float
		// 	nb := np.NewBuilder()
		// 	nb.AssignFloat(1.38)
		// 	nd := nb.Build()

		// 	pol := policy.Policy{GreaterThanOrpolicy.Equal(mustParse(t, "."), literal.Float(1))}
		// 	ok, err := policy.Match(pol, nd)
		// 	require.True(t, ok)

		// 	pol = policy.Policy{GreaterThanOrpolicy.Equal(mustParse(t, "."), literal.Float(1.38))}
		// 	ok, err = policy.Match(pol, nd)
		// 	require.True(t, ok)
		// })

		t.Run("lt int", func(t *testing.T) {
			val := 138
			pol := testutil.Must(policy.Build(policy.LessThan(".", 1138)))(t)
			ok, err := policy.Match(pol, val)
			require.True(t, ok)
			require.NoError(t, err)
		})

		t.Run("lte int", func(t *testing.T) {
			val := 138

			pol := testutil.Must(policy.Build(policy.LessThanOrEqual(".", 1138)))(t)
			ok, err := policy.Match(pol, val)
			require.True(t, ok)
			require.NoError(t, err)

			pol = testutil.Must(policy.Build(policy.LessThanOrEqual(".", 138)))(t)
			ok, err = policy.Match(pol, val)
			require.True(t, ok)
			require.NoError(t, err)
		})
	})

	t.Run("negation", func(t *testing.T) {
		val := false

		pol := testutil.Must(policy.Build(policy.Not(policy.Equal(".", true))))(t)
		ok, err := policy.Match(pol, val)
		require.True(t, ok)
		require.NoError(t, err)

		pol = testutil.Must(policy.Build(policy.Not(policy.Equal(".", false))))(t)
		ok, err = policy.Match(pol, val)
		require.False(t, ok)
		require.Error(t, err)
		t.Log(err)
	})

	t.Run("conjunction", func(t *testing.T) {
		val := 138

		pol := testutil.Must(
			policy.Build(
				policy.And(
					policy.GreaterThan(".", 1),
					policy.LessThan(".", 1138),
				),
			),
		)(t)
		ok, err := policy.Match(pol, val)
		require.True(t, ok)
		require.NoError(t, err)

		pol = testutil.Must(
			policy.Build(
				policy.And(
					policy.GreaterThan(".", 1),
					policy.Equal(".", 1138),
				),
			),
		)(t)
		ok, err = policy.Match(pol, val)
		require.False(t, ok)
		require.Error(t, err)
		t.Log(err)

		pol = testutil.Must(
			policy.Build(
				policy.And(),
			),
		)(t)
		ok, err = policy.Match(pol, val)
		require.True(t, ok)
		require.NoError(t, err)
	})

	// t.Run("disjunction", func(t *testing.T) {
	// 	np := basicnode.Prototype.Int
	// 	nb := np.NewBuilder()
	// 	nb.AssignInt(138)
	// 	nd := nb.Build()

	// 	pol := policy.Policy{
	// 		Or(
	// 			GreaterThan(mustParse(t, "."), literal.Int(138)),
	// 			LessThan(mustParse(t, "."), literal.Int(1138)),
	// 		),
	// 	}
	// 	ok, err := policy.Match(pol, nd)
	// 	require.True(t, ok)

	// 	pol = policy.Policy{
	// 		Or(
	// 			GreaterThan(mustParse(t, "."), literal.Int(138)),
	// 			policy.Equal(mustParse(t, "."), literal.Int(1138)),
	// 		),
	// 	}
	// 	ok, err = policy.Match(pol, nd)
	// 	require.False(t, ok)

	// 	pol = policy.Policy{Or()}
	// 	ok, err = policy.Match(pol, nd)
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

	// 				pol := policy.Policy{Like(mustParse(t, "."), glb)}
	// 				ok, err := policy.Match(pol, nd)
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

	// 				pol := policy.Policy{Like(mustParse(t, "."), glb)}
	// 				ok, err := policy.Match(pol, nd)
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

	// 		pol := policy.Policy{
	// 			All(
	// 				mustParse(t, ".[]"),
	// 				GreaterThan(mustParse(t, ".value"), literal.Int(2)),
	// 			),
	// 		}
	// 		ok, err := policy.Match(pol, nd)
	// 		require.True(t, ok)

	// 		pol = policy.Policy{
	// 			All(
	// 				mustParse(t, ".[]"),
	// 				GreaterThan(mustParse(t, ".value"), literal.Int(20)),
	// 			),
	// 		}
	// 		ok, err = policy.Match(pol, nd)
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

	// 		pol := policy.Policy{
	// 			Any(
	// 				mustParse(t, ".[]"),
	// 				GreaterThan(mustParse(t, ".value"), literal.Int(10)),
	// 				LessThan(mustParse(t, ".value"), literal.Int(50)),
	// 			),
	// 		}
	// 		ok, err := policy.Match(pol, nd)
	// 		require.True(t, ok)

	// 		pol = policy.Policy{
	// 			Any(
	// 				mustParse(t, ".[]"),
	// 				GreaterThan(mustParse(t, ".value"), literal.Int(100)),
	// 			),
	// 		}
	// 		ok, err = policy.Match(pol, nd)
	// 		require.False(t, ok)
	// 	})
	// })
}
