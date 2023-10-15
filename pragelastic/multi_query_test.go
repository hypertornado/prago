package pragelastic

/*

func TestMultiQuery(t *testing.T) {
	index := prepareTestIndex[TestStruct]()
	index.UpdateSingle(&TestStruct{
		ID:        "1",
		SomeCount: 1,
	})

	index.UpdateSingle(&TestStruct{
		ID:        "2",
		SomeCount: 2,
	})
	index.UpdateSingle(&TestStruct{
		ID:        "3",
		SomeCount: 3,
	})
	index.UpdateSingle(&TestStruct{
		ID:        "4",
		SomeCount: 4,
	})

	index.Flush()
	index.Refresh()

	mq := index.MultiQuery()
	mq.Add(
		index.Query().Range("SomeCount", 1, 1),
		index.Query().Range("SomeCount", 2, 3),
	)

	results, err := mq.Search()
	if err != nil {
		t.Fatal(err)
	}
	for i, v := range results {
		_, c, _ := index.SearchResultToList(v)
		if i == 0 {
			if c != 1 {
				t.Fatal(i, c)
			}
		}
		if i == 1 {
			if c != 2 {
				t.Fatal(i, c)
			}
		}
	}

}

*/
