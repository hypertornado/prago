package pragelastic

import (
	"bytes"
	"encoding/json"
)

type ESAggregation interface {
	Source() (interface{}, error)
}

func (q *Query[T]) Aggregation(name string, aggregation ESAggregation) *Query[T] {
	q.aggregations[name] = aggregation
	return q
}

type ESAggregations map[string]json.RawMessage

type ESAggregationValueMetric struct {
	ESAggregations

	Value *float64               //`json:"value"`
	Meta  map[string]interface{} // `json:"meta,omitempty"`
}

func (a ESAggregations) Sum(name string) (*ESAggregationValueMetric, bool) {
	if raw, found := a[name]; found {
		agg := new(ESAggregationValueMetric)
		if raw == nil {
			return agg, true
		}
		if err := json.Unmarshal(raw, agg); err == nil {
			return agg, true
		}
	}
	return nil, false
}

func (a ESAggregations) Histogram(name string) (*ESAggregationBucketHistogramItems, bool) {
	if raw, found := a[name]; found {
		agg := new(ESAggregationBucketHistogramItems)
		if raw == nil {
			return agg, true
		}
		if err := json.Unmarshal(raw, agg); err == nil {
			return agg, true
		}
	}
	return nil, false
}

type ESAggregationBucketHistogramItems struct {
	ESAggregations

	Buckets  []*ESAggregationBucketHistogramItem //`json:"buckets"`
	Interval interface{}                         // `json:"interval"` // can be numeric or a string
	Meta     map[string]interface{}              // `json:"meta,omitempty"`
}

// UnmarshalJSON decodes JSON data and initializes an AggregationBucketHistogramItems structure.
func (a *ESAggregationBucketHistogramItems) UnmarshalJSON(data []byte) error {
	var aggs map[string]json.RawMessage
	if err := json.Unmarshal(data, &aggs); err != nil {
		return err
	}
	if v, ok := aggs["buckets"]; ok && v != nil {
		json.Unmarshal(v, &a.Buckets)
	}
	if v, ok := aggs["interval"]; ok && v != nil {
		json.Unmarshal(v, &a.Interval)
	}
	if v, ok := aggs["meta"]; ok && v != nil {
		json.Unmarshal(v, &a.Meta)
	}
	a.ESAggregations = aggs
	return nil
}

type ESAggregationBucketHistogramItem struct {
	ESAggregations

	Key         float64 //`json:"key"`
	KeyAsString *string //`json:"key_as_string"`
	DocCount    int64   //`json:"doc_count"`
}

// UnmarshalJSON decodes JSON data and initializes an AggregationBucketHistogramItem structure.
func (a *ESAggregationBucketHistogramItem) UnmarshalJSON(data []byte) error {
	var aggs map[string]json.RawMessage
	if err := json.Unmarshal(data, &aggs); err != nil {
		return err
	}
	if v, ok := aggs["key"]; ok && v != nil {
		json.Unmarshal(v, &a.Key)
	}
	if v, ok := aggs["key_as_string"]; ok && v != nil {
		json.Unmarshal(v, &a.KeyAsString)
	}
	if v, ok := aggs["doc_count"]; ok && v != nil {
		json.Unmarshal(v, &a.DocCount)
	}
	a.ESAggregations = aggs
	return nil
}

func (a ESAggregations) Terms(name string) (*ESAggregationBucketKeyItems, bool) {
	if raw, found := a[name]; found {
		agg := new(ESAggregationBucketKeyItems)
		if raw == nil {
			return agg, true
		}
		if err := json.Unmarshal(raw, agg); err == nil {
			return agg, true
		}
	}
	return nil, false
}

type ESAggregationBucketKeyItems struct {
	ESAggregations

	DocCountErrorUpperBound int64                         //`json:"doc_count_error_upper_bound"`
	SumOfOtherDocCount      int64                         //`json:"sum_other_doc_count"`
	Buckets                 []*ESAggregationBucketKeyItem //`json:"buckets"`
	Meta                    map[string]interface{}        // `json:"meta,omitempty"`
}

// UnmarshalJSON decodes JSON data and initializes an AggregationBucketKeyItems structure.
func (a *ESAggregationBucketKeyItems) UnmarshalJSON(data []byte) error {
	var aggs map[string]json.RawMessage
	if err := json.Unmarshal(data, &aggs); err != nil {
		return err
	}
	if v, ok := aggs["doc_count_error_upper_bound"]; ok && v != nil {
		json.Unmarshal(v, &a.DocCountErrorUpperBound)
	}
	if v, ok := aggs["sum_other_doc_count"]; ok && v != nil {
		json.Unmarshal(v, &a.SumOfOtherDocCount)
	}
	if v, ok := aggs["buckets"]; ok && v != nil {
		json.Unmarshal(v, &a.Buckets)
	}
	if v, ok := aggs["meta"]; ok && v != nil {
		json.Unmarshal(v, &a.Meta)
	}
	a.ESAggregations = aggs
	return nil
}

type ESAggregationBucketKeyItem struct {
	ESAggregations

	Key         interface{} //`json:"key"`
	KeyAsString *string     //`json:"key_as_string"`
	KeyNumber   json.Number
	DocCount    int64 //`json:"doc_count"`
}

// UnmarshalJSON decodes JSON data and initializes an AggregationBucketKeyItem structure.
func (a *ESAggregationBucketKeyItem) UnmarshalJSON(data []byte) error {
	var aggs map[string]json.RawMessage
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&aggs); err != nil {
		return err
	}
	if v, ok := aggs["key"]; ok && v != nil {
		json.Unmarshal(v, &a.Key)
		json.Unmarshal(v, &a.KeyNumber)
	}
	if v, ok := aggs["key_as_string"]; ok && v != nil {
		json.Unmarshal(v, &a.KeyAsString)
	}
	if v, ok := aggs["doc_count"]; ok && v != nil {
		json.Unmarshal(v, &a.DocCount)
	}
	a.ESAggregations = aggs
	return nil
}
