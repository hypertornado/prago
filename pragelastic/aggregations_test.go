package pragelastic

import (
	"testing"

	"github.com/olivere/elastic/v7"
)

//https://github.com/olivere/elastic/blob/release-branch.v7/search_aggs_test.go

func TestSumAggregations(t *testing.T) {
	index := prepareTestIndex[TestStruct]()
	index.UpdateSingle(&TestStruct{
		ID:        "1",
		Name:      "A",
		SomeCount: 5,
	})
	index.UpdateSingle(&TestStruct{
		ID:        "2",
		Name:      "A",
		SomeCount: 7,
	})
	index.UpdateSingle(&TestStruct{
		ID:        "3",
		Name:      "B",
		SomeCount: 7,
	})

	index.Flush()
	index.Refresh()

	sumAgg := elastic.NewSumAggregation().Field("SomeCount")
	res, err := index.Query().Aggregation("sum", sumAgg).SearchResult()
	if err != nil {
		t.Fatal(err)
	}

	sumRes, ok := res.Aggregations.Sum("sum")
	if !ok {
		t.Fatal("wrong")
	}
	if *sumRes.Value != 19 {
		t.Fatal(*sumRes.Value)
	}
}

func TestBucketHistogramAggregations(t *testing.T) {
	index := prepareTestIndex[TestStruct]()
	index.UpdateSingle(&TestStruct{
		ID:        "1",
		SomeCount: 5,
	})
	index.UpdateSingle(&TestStruct{
		ID:        "2",
		SomeCount: 7,
	})
	index.UpdateSingle(&TestStruct{
		ID:        "3",
		Name:      "B",
		SomeCount: 7,
	})

	index.Flush()
	index.Refresh()

	agg := elastic.NewHistogramAggregation().Field("SomeCount").Interval(1) //.Offset(3)

	res, err := index.Query().Aggregation("agg", agg).SearchResult()
	if err != nil {
		t.Fatal(err)
	}

	aggRes, ok := res.Aggregations.Histogram("agg")
	if !ok {
		t.Fatal("wrong")
	}

	if len(aggRes.Buckets) != 3 {
		t.Fatal(len(aggRes.Buckets))
	}

	if aggRes.Buckets[0].DocCount != 1 {
		t.Fatal(aggRes.Buckets[0].DocCount)
	}

	if aggRes.Buckets[1].DocCount != 0 {
		t.Fatal(aggRes.Buckets[1].DocCount)
	}

	if aggRes.Buckets[2].DocCount != 2 {
		t.Fatal(aggRes.Buckets[2].DocCount)
	}
}
