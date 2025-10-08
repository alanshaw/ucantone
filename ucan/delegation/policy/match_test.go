package policy_test

import (
	"testing"

	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/testing/helpers"
	"github.com/alanshaw/ucantone/ucan/delegation/policy"
	"github.com/alanshaw/ucantone/ucan/delegation/policy/selector"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"
)

func TestMatch(t *testing.T) {
	t.Run("equality", func(t *testing.T) {
		t.Run("string", func(t *testing.T) {
			sel := helpers.Must(selector.Parse("."))(t)
			val := "test"

			pol := policy.Policy{Statements: []policy.Statement{policy.Equal(sel, "test")}}
			ok, err := policy.Match(pol, val)
			require.True(t, ok)
			require.Nil(t, err)

			pol = policy.Policy{Statements: []policy.Statement{policy.Equal(sel, "test2")}}
			ok, err = policy.Match(pol, val)
			require.False(t, ok)
			require.NotNil(t, err)
			t.Log(err)

			pol = policy.Policy{Statements: []policy.Statement{policy.Equal(sel, 138)}}
			ok, err = policy.Match(pol, val)
			require.False(t, ok)
			require.NotNil(t, err)
			t.Log(err)
		})

		t.Run("int", func(t *testing.T) {
			sel := helpers.Must(selector.Parse("."))(t)
			val := 138

			pol := policy.Policy{Statements: []policy.Statement{policy.Equal(sel, 138)}}
			ok, err := policy.Match(pol, val)
			require.True(t, ok)
			require.Nil(t, err)

			pol = policy.Policy{Statements: []policy.Statement{policy.Equal(sel, 1138)}}
			ok, err = policy.Match(pol, val)
			require.False(t, ok)
			require.NotNil(t, err)
			t.Log(err)

			pol = policy.Policy{Statements: []policy.Statement{policy.Equal(sel, "138")}}
			ok, err = policy.Match(pol, val)
			require.False(t, ok)
			require.NotNil(t, err)
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
			sel := helpers.Must(selector.Parse("."))(t)
			l0 := cid.MustParse("bafybeif4owy5gno5lwnixqm52rwqfodklf76hsetxdhffuxnplvijskzqq")
			l1 := cid.MustParse("bafkreifau35r7vi37tvbvfy3hdwvgb4tlflqf7zcdzeujqcjk3rsphiwte")
			val := l0

			pol := policy.Policy{Statements: []policy.Statement{policy.Equal(sel, l0)}}
			ok, err := policy.Match(pol, val)
			require.True(t, ok)
			require.Nil(t, err)

			pol = policy.Policy{Statements: []policy.Statement{policy.Equal(sel, l1)}}
			ok, err = policy.Match(pol, val)
			require.False(t, ok)
			require.NotNil(t, err)
			t.Log(err)

			pol = policy.Policy{Statements: []policy.Statement{policy.Equal(sel, "bafybeif4owy5gno5lwnixqm52rwqfodklf76hsetxdhffuxnplvijskzqq")}}
			ok, err = policy.Match(pol, val)
			require.False(t, ok)
			require.NotNil(t, err)
			t.Log(err)
		})

		t.Run("string in map", func(t *testing.T) {
			val := datamodel.NewMap(datamodel.WithEntry("foo", "bar"))

			sel := helpers.Must(selector.Parse(".foo"))(t)
			pol := policy.Policy{Statements: []policy.Statement{policy.Equal(sel, "bar")}}
			ok, err := policy.Match(pol, val)
			require.True(t, ok)
			require.Nil(t, err)

			sel = helpers.Must(selector.Parse(`.["foo"]`))(t)
			pol = policy.Policy{Statements: []policy.Statement{policy.Equal(sel, "bar")}}
			ok, err = policy.Match(pol, val)
			require.True(t, ok)
			require.Nil(t, err)

			sel = helpers.Must(selector.Parse(".foo"))(t)
			pol = policy.Policy{Statements: []policy.Statement{policy.Equal(sel, "baz")}}
			ok, err = policy.Match(pol, val)
			require.False(t, ok)
			require.NotNil(t, err)
			t.Log(err)

			sel = helpers.Must(selector.Parse(".foobar"))(t)
			pol = policy.Policy{Statements: []policy.Statement{policy.Equal(sel, "bar")}}
			ok, err = policy.Match(pol, val)
			require.False(t, ok)
			require.NotNil(t, err)
			t.Log(err)
		})

		t.Run("string in list", func(t *testing.T) {
			val := []string{"foo"}

			sel := helpers.Must(selector.Parse(".[0]"))(t)
			pol := policy.Policy{Statements: []policy.Statement{policy.Equal(sel, "foo")}}
			ok, err := policy.Match(pol, val)
			require.True(t, ok)
			require.Nil(t, err)

			sel = helpers.Must(selector.Parse(".[1]"))(t)
			pol = policy.Policy{Statements: []policy.Statement{policy.Equal(sel, "foo")}}
			ok, err = policy.Match(pol, val)
			require.False(t, ok)
			require.NotNil(t, err)
			t.Log(err)
		})
	})

	t.Run("inequality", func(t *testing.T) {
		t.Run("gt int", func(t *testing.T) {
			sel := helpers.Must(selector.Parse("."))(t)
			val := 138
			pol := policy.Policy{Statements: []policy.Statement{policy.GreaterThan(sel, 1)}}
			ok, err := policy.Match(pol, val)
			require.True(t, ok)
			require.Nil(t, err)
		})

		t.Run("gte int", func(t *testing.T) {
			sel := helpers.Must(selector.Parse("."))(t)
			val := 138

			pol := policy.Policy{Statements: []policy.Statement{policy.GreaterThanOrEqual(sel, 1)}}
			ok, err := policy.Match(pol, val)
			require.True(t, ok)
			require.Nil(t, err)

			pol = policy.Policy{Statements: []policy.Statement{policy.GreaterThanOrEqual(sel, 138)}}
			ok, err = policy.Match(pol, val)
			require.True(t, ok)
			require.Nil(t, err)
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
			sel := helpers.Must(selector.Parse("."))(t)
			val := 138
			pol := policy.Policy{Statements: []policy.Statement{policy.LessThan(sel, 1138)}}
			ok, err := policy.Match(pol, val)
			require.True(t, ok)
			require.Nil(t, err)
		})

		t.Run("lte int", func(t *testing.T) {
			sel := helpers.Must(selector.Parse("."))(t)
			val := 138

			pol := policy.Policy{Statements: []policy.Statement{policy.LessThanOrEqual(sel, 1138)}}
			ok, err := policy.Match(pol, val)
			require.True(t, ok)
			require.Nil(t, err)

			pol = policy.Policy{Statements: []policy.Statement{policy.LessThanOrEqual(sel, 138)}}
			ok, err = policy.Match(pol, val)
			require.True(t, ok)
			require.Nil(t, err)
		})
	})

	// t.Run("negation", func(t *testing.T) {
	// 	np := basicnode.Prototype.Bool
	// 	nb := np.NewBuilder()
	// 	nb.AssignBool(false)
	// 	nd := nb.Build()

	// 	pol := policy.Policy{Not(policy.Equal(mustParse(t, "."), literal.Bool(true)))}
	// 	ok, err := policy.Match(pol, nd)
	// 	require.True(t, ok)

	// 	pol = policy.Policy{Not(policy.Equal(mustParse(t, "."), literal.Bool(false)))}
	// 	ok, err = policy.Match(pol, nd)
	// 	require.False(t, ok)
	// })

	// t.Run("conjunction", func(t *testing.T) {
	// 	np := basicnode.Prototype.Int
	// 	nb := np.NewBuilder()
	// 	nb.AssignInt(138)
	// 	nd := nb.Build()

	// 	pol := policy.Policy{
	// 		And(
	// 			GreaterThan(mustParse(t, "."), literal.Int(1)),
	// 			LessThan(mustParse(t, "."), literal.Int(1138)),
	// 		),
	// 	}
	// 	ok, err := policy.Match(pol, nd)
	// 	require.True(t, ok)

	// 	pol = policy.Policy{
	// 		And(
	// 			GreaterThan(mustParse(t, "."), literal.Int(1)),
	// 			policy.Equal(mustParse(t, "."), literal.Int(1138)),
	// 		),
	// 	}
	// 	ok, err = policy.Match(pol, nd)
	// 	require.False(t, ok)

	// 	pol = policy.Policy{And()}
	// 	ok, err = policy.Match(pol, nd)
	// 	require.True(t, ok)
	// })

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
