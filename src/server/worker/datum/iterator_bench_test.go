//go:build k8s

package datum

// TODO: Update if useful.
//func benchmarkIterators(j int, b *testing.B) {
//	c, _ := minikubetestenv.AcquireCluster(b)
//
//	b.ResetTimer()
//	for n := 0; n < b.N; n++ {
//		dataRepo := tu.UniqueString("TestIteratorPFS_data")
//		require.NoError(b, c.CreateRepo(dataRepo))
//
//		// put files in structured in a way so that there are many ways to glob it
//		commit, err := c.StartCommit(dataRepo, "master")
//		require.NoError(b, err)
//		for i := 0; i < 100*j; i++ {
//			require.NoError(b, c.PutFile(commit, fmt.Sprintf("foo%v", i), strings.NewReader("bar")))
//		}
//		require.NoError(b, c.FinishCommit(dataRepo, commit.Branch.Name, commit.ID))
//
//		// make one with zero datums for testing edge cases
//		in0 := client.NewPFSInput(dataRepo, "!(**)")
//		in0.Pfs.Commit = commit.ID
//		pfs0, err := NewIterator(c, in0)
//		require.NoError(b, err)
//
//		in1 := client.NewPFSInput(dataRepo, "/foo?1*")
//		in1.Pfs.Commit = commit.ID
//		pfs1, err := NewIterator(c, in1)
//		require.NoError(b, err)
//
//		in2 := client.NewPFSInput(dataRepo, "/foo*2")
//		in2.Pfs.Commit = commit.ID
//		pfs2, err := NewIterator(c, in2)
//		require.NoError(b, err)
//
//		validateDI(b, pfs0)
//		validateDI(b, pfs1)
//		validateDI(b, pfs2)
//
//		b.Run("union", func(b *testing.B) {
//			in3 := client.NewUnionInput(in1, in2)
//			union1, err := NewIterator(c, in3)
//			require.NoError(b, err)
//			validateDI(b, union1)
//		})
//
//		b.Run("cross", func(b *testing.B) {
//			in4 := client.NewCrossInput(in1, in2)
//			cross1, err := NewIterator(c, in4)
//			require.NoError(b, err)
//			validateDI(b, cross1)
//		})
//
//		// TODO: Implement for V2.
//		//b.Run("join", func(b *testing.B) {
//		//	in8 := client.NewPFSInputOpts("", dataRepo, "", "/foo(?)(?)*", "$1$2", "", false, false)
//		//	in8.Pfs.Commit = commit.ID
//		//	in9 := client.NewPFSInputOpts("", dataRepo, "", "/foo(?)(?)*", "$2$1", "", false, false)
//		//	in9.Pfs.Commit = commit.ID
//		//	join1, err := newJoinIterator(c, []*pps.Input{in8, in9})
//		//	require.NoError(b, err)
//		//	validateDI(b, join1)
//		//})
//
//		//b.Run("group", func(b *testing.B) {
//		//	in10 := client.NewPFSInputOpts("", dataRepo, "", "/foo(?)(?)*", "", "$2", false, false)
//		//	in10.Pfs.Commit = commit.ID
//		//	group1, err := newGroupIterator(c, []*pps.Input{in10})
//		//	require.NoError(b, err)
//		//	validateDI(b, group1)
//		//})
//
//		b.Run("iterated", func(b *testing.B) {
//			in3 := client.NewUnionInput(in1, in2)
//			in4 := client.NewCrossInput(in1, in2)
//
//			in5 := client.NewCrossInput(in3, in4)
//			cross2, err := NewIterator(c, in5)
//			require.NoError(b, err)
//
//			// cross with a zero datum input should also be zero
//			in6 := client.NewCrossInput(in3, in0, in2, in4)
//			cross3, err := NewIterator(c, in6)
//			require.NoError(b, err)
//
//			// zero cross inside a cross should also be zero
//			in7 := client.NewCrossInput(in6, in1)
//			cross4, err := NewIterator(c, in7)
//			require.NoError(b, err)
//
//			validateDI(b, cross2)
//			validateDI(b, cross3)
//			validateDI(b, cross4)
//
//		})
//	}
//}
//
//func BenchmarkDI1(b *testing.B)  { benchmarkIterators(1, b) }
//func BenchmarkDI2(b *testing.B)  { benchmarkIterators(2, b) }
//func BenchmarkDI4(b *testing.B)  { benchmarkIterators(4, b) }
//func BenchmarkDI8(b *testing.B)  { benchmarkIterators(8, b) }
//func BenchmarkDI16(b *testing.B) { benchmarkIterators(16, b) }
//func BenchmarkDI32(b *testing.B) { benchmarkIterators(32, b) }
