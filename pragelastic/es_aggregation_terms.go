package pragelastic

import "fmt"

// TermsAggregation is a multi-bucket value source based aggregation
// where buckets are dynamically built - one per unique value.
//
// See: http://www.elastic.co/guide/en/elasticsearch/reference/7.0/search-aggregations-bucket-terms-aggregation.html
type ESTermsAggregation struct {
	field string
	//script          *Script
	missing         interface{}
	subAggregations map[string]ESAggregation
	meta            map[string]interface{}

	size                  *int
	shardSize             *int
	requiredSize          *int
	minDocCount           *int
	shardMinDocCount      *int
	valueType             string
	includeExclude        *ESTermsAggregationIncludeExclude
	executionHint         string
	collectionMode        string
	showTermDocCountError *bool
	order                 []TermsOrder
}

func NewESTermsAggregation() *ESTermsAggregation {
	return &ESTermsAggregation{
		subAggregations: make(map[string]ESAggregation),
	}
}

func (a *ESTermsAggregation) Field(field string) *ESTermsAggregation {
	a.field = field
	return a
}

/*func (a *TermsAggregation) Script(script *Script) *TermsAggregation {
	a.script = script
	return a
}*/

// Missing configures the value to use when documents miss a value.
func (a *ESTermsAggregation) Missing(missing interface{}) *ESTermsAggregation {
	a.missing = missing
	return a
}

func (a *ESTermsAggregation) SubAggregation(name string, subAggregation ESAggregation) *ESTermsAggregation {
	a.subAggregations[name] = subAggregation
	return a
}

// Meta sets the meta data to be included in the aggregation response.
func (a *ESTermsAggregation) Meta(metaData map[string]interface{}) *ESTermsAggregation {
	a.meta = metaData
	return a
}

func (a *ESTermsAggregation) Size(size int) *ESTermsAggregation {
	a.size = &size
	return a
}

func (a *ESTermsAggregation) RequiredSize(requiredSize int) *ESTermsAggregation {
	a.requiredSize = &requiredSize
	return a
}

func (a *ESTermsAggregation) ShardSize(shardSize int) *ESTermsAggregation {
	a.shardSize = &shardSize
	return a
}

func (a *ESTermsAggregation) MinDocCount(minDocCount int) *ESTermsAggregation {
	a.minDocCount = &minDocCount
	return a
}

func (a *ESTermsAggregation) ShardMinDocCount(shardMinDocCount int) *ESTermsAggregation {
	a.shardMinDocCount = &shardMinDocCount
	return a
}

func (a *ESTermsAggregation) Include(regexp string) *ESTermsAggregation {
	if a.includeExclude == nil {
		a.includeExclude = &ESTermsAggregationIncludeExclude{}
	}
	a.includeExclude.Include = regexp
	return a
}

func (a *ESTermsAggregation) IncludeValues(values ...interface{}) *ESTermsAggregation {
	if a.includeExclude == nil {
		a.includeExclude = &ESTermsAggregationIncludeExclude{}
	}
	a.includeExclude.IncludeValues = append(a.includeExclude.IncludeValues, values...)
	return a
}

func (a *ESTermsAggregation) Exclude(regexp string) *ESTermsAggregation {
	if a.includeExclude == nil {
		a.includeExclude = &ESTermsAggregationIncludeExclude{}
	}
	a.includeExclude.Exclude = regexp
	return a
}

func (a *ESTermsAggregation) ExcludeValues(values ...interface{}) *ESTermsAggregation {
	if a.includeExclude == nil {
		a.includeExclude = &ESTermsAggregationIncludeExclude{}
	}
	a.includeExclude.ExcludeValues = append(a.includeExclude.ExcludeValues, values...)
	return a
}

func (a *ESTermsAggregation) Partition(p int) *ESTermsAggregation {
	if a.includeExclude == nil {
		a.includeExclude = &ESTermsAggregationIncludeExclude{}
	}
	a.includeExclude.Partition = p
	return a
}

func (a *ESTermsAggregation) NumPartitions(n int) *ESTermsAggregation {
	if a.includeExclude == nil {
		a.includeExclude = &ESTermsAggregationIncludeExclude{}
	}
	a.includeExclude.NumPartitions = n
	return a
}

func (a *ESTermsAggregation) IncludeExclude(includeExclude *ESTermsAggregationIncludeExclude) *ESTermsAggregation {
	a.includeExclude = includeExclude
	return a
}

// ValueType can be string, long, or double.
func (a *ESTermsAggregation) ValueType(valueType string) *ESTermsAggregation {
	a.valueType = valueType
	return a
}

func (a *ESTermsAggregation) Order(order string, asc bool) *ESTermsAggregation {
	a.order = append(a.order, TermsOrder{Field: order, Ascending: asc})
	return a
}

func (a *ESTermsAggregation) OrderByCount(asc bool) *ESTermsAggregation {
	// "order" : { "_count" : "asc" }
	a.order = append(a.order, TermsOrder{Field: "_count", Ascending: asc})
	return a
}

func (a *ESTermsAggregation) OrderByCountAsc() *ESTermsAggregation {
	return a.OrderByCount(true)
}

func (a *ESTermsAggregation) OrderByCountDesc() *ESTermsAggregation {
	return a.OrderByCount(false)
}

// Deprecated: Use OrderByKey instead.
func (a *ESTermsAggregation) OrderByTerm(asc bool) *ESTermsAggregation {
	// "order" : { "_term" : "asc" }
	a.order = append(a.order, TermsOrder{Field: "_term", Ascending: asc})
	return a
}

// Deprecated: Use OrderByKeyAsc instead.
func (a *ESTermsAggregation) OrderByTermAsc() *ESTermsAggregation {
	return a.OrderByTerm(true)
}

// Deprecated: Use OrderByKeyDesc instead.
func (a *ESTermsAggregation) OrderByTermDesc() *ESTermsAggregation {
	return a.OrderByTerm(false)
}

func (a *ESTermsAggregation) OrderByKey(asc bool) *ESTermsAggregation {
	// "order" : { "_term" : "asc" }
	a.order = append(a.order, TermsOrder{Field: "_key", Ascending: asc})
	return a
}

func (a *ESTermsAggregation) OrderByKeyAsc() *ESTermsAggregation {
	return a.OrderByKey(true)
}

func (a *ESTermsAggregation) OrderByKeyDesc() *ESTermsAggregation {
	return a.OrderByKey(false)
}

// OrderByAggregation creates a bucket ordering strategy which sorts buckets
// based on a single-valued calc get.
func (a *ESTermsAggregation) OrderByAggregation(aggName string, asc bool) *ESTermsAggregation {
	// {
	//     "aggs" : {
	//         "genders" : {
	//             "terms" : {
	//                 "field" : "gender",
	//                 "order" : { "avg_height" : "desc" }
	//             },
	//             "aggs" : {
	//                 "avg_height" : { "avg" : { "field" : "height" } }
	//             }
	//         }
	//     }
	// }
	a.order = append(a.order, TermsOrder{Field: aggName, Ascending: asc})
	return a
}

// OrderByAggregationAndMetric creates a bucket ordering strategy which
// sorts buckets based on a multi-valued calc get.
func (a *ESTermsAggregation) OrderByAggregationAndMetric(aggName, metric string, asc bool) *ESTermsAggregation {
	// {
	//     "aggs" : {
	//         "genders" : {
	//             "terms" : {
	//                 "field" : "gender",
	//                 "order" : { "height_stats.avg" : "desc" }
	//             },
	//             "aggs" : {
	//                 "height_stats" : { "stats" : { "field" : "height" } }
	//             }
	//         }
	//     }
	// }
	a.order = append(a.order, TermsOrder{Field: aggName + "." + metric, Ascending: asc})
	return a
}

func (a *ESTermsAggregation) ExecutionHint(hint string) *ESTermsAggregation {
	a.executionHint = hint
	return a
}

// Collection mode can be depth_first or breadth_first as of 1.4.0.
func (a *ESTermsAggregation) CollectionMode(collectionMode string) *ESTermsAggregation {
	a.collectionMode = collectionMode
	return a
}

func (a *ESTermsAggregation) ShowTermDocCountError(showTermDocCountError bool) *ESTermsAggregation {
	a.showTermDocCountError = &showTermDocCountError
	return a
}

func (a *ESTermsAggregation) Source() (interface{}, error) {
	// Example:
	//	{
	//    "aggs" : {
	//      "genders" : {
	//        "terms" : { "field" : "gender" }
	//      }
	//    }
	//	}
	// This method returns only the { "terms" : { "field" : "gender" } } part.

	source := make(map[string]interface{})
	opts := make(map[string]interface{})
	source["terms"] = opts

	// ValuesSourceAggregationBuilder
	if a.field != "" {
		opts["field"] = a.field
	}
	/*if a.script != nil {
		src, err := a.script.Source()
		if err != nil {
			return nil, err
		}
		opts["script"] = src
	}*/
	if a.missing != nil {
		opts["missing"] = a.missing
	}

	// TermsBuilder
	if a.size != nil && *a.size >= 0 {
		opts["size"] = *a.size
	}
	if a.shardSize != nil && *a.shardSize >= 0 {
		opts["shard_size"] = *a.shardSize
	}
	if a.requiredSize != nil && *a.requiredSize >= 0 {
		opts["required_size"] = *a.requiredSize
	}
	if a.minDocCount != nil && *a.minDocCount >= 0 {
		opts["min_doc_count"] = *a.minDocCount
	}
	if a.shardMinDocCount != nil && *a.shardMinDocCount >= 0 {
		opts["shard_min_doc_count"] = *a.shardMinDocCount
	}
	if a.showTermDocCountError != nil {
		opts["show_term_doc_count_error"] = *a.showTermDocCountError
	}
	if a.collectionMode != "" {
		opts["collect_mode"] = a.collectionMode
	}
	if a.valueType != "" {
		opts["value_type"] = a.valueType
	}
	if len(a.order) > 0 {
		var orderSlice []interface{}
		for _, order := range a.order {
			src, err := order.Source()
			if err != nil {
				return nil, err
			}
			orderSlice = append(orderSlice, src)
		}
		opts["order"] = orderSlice
	}

	// Include/Exclude
	if ie := a.includeExclude; ie != nil {
		if err := ie.MergeInto(opts); err != nil {
			return nil, err
		}
	}

	if a.executionHint != "" {
		opts["execution_hint"] = a.executionHint
	}

	// AggregationBuilder (SubAggregations)
	if len(a.subAggregations) > 0 {
		aggsMap := make(map[string]interface{})
		source["aggregations"] = aggsMap
		for name, aggregate := range a.subAggregations {
			src, err := aggregate.Source()
			if err != nil {
				return nil, err
			}
			aggsMap[name] = src
		}
	}

	// Add Meta data if available
	if len(a.meta) > 0 {
		source["meta"] = a.meta
	}

	return source, nil
}

// TermsAggregationIncludeExclude allows for include/exclude in a TermsAggregation.
type ESTermsAggregationIncludeExclude struct {
	Include       string
	Exclude       string
	IncludeValues []interface{}
	ExcludeValues []interface{}
	Partition     int
	NumPartitions int
}

// Source returns a JSON serializable struct.
func (ie *ESTermsAggregationIncludeExclude) Source() (interface{}, error) {
	source := make(map[string]interface{})

	// Include
	if ie.Include != "" {
		source["include"] = ie.Include
	} else if len(ie.IncludeValues) > 0 {
		source["include"] = ie.IncludeValues
	} else if ie.NumPartitions > 0 {
		inc := make(map[string]interface{})
		inc["partition"] = ie.Partition
		inc["num_partitions"] = ie.NumPartitions
		source["include"] = inc
	}

	// Exclude
	if ie.Exclude != "" {
		source["exclude"] = ie.Exclude
	} else if len(ie.ExcludeValues) > 0 {
		source["exclude"] = ie.ExcludeValues
	}

	return source, nil
}

// MergeInto merges the values of the include/exclude options into source.
func (ie *ESTermsAggregationIncludeExclude) MergeInto(source map[string]interface{}) error {
	values, err := ie.Source()
	if err != nil {
		return err
	}
	mv, ok := values.(map[string]interface{})
	if !ok {
		return fmt.Errorf("IncludeExclude: expected a map[string]interface{}, got %T", values)
	}
	for k, v := range mv {
		source[k] = v
	}
	return nil
}

// TermsOrder specifies a single order field for a terms aggregation.
type TermsOrder struct {
	Field     string
	Ascending bool
}

// Source returns serializable JSON of the TermsOrder.
func (order *TermsOrder) Source() (interface{}, error) {
	source := make(map[string]string)
	if order.Ascending {
		source[order.Field] = "asc"
	} else {
		source[order.Field] = "desc"
	}
	return source, nil
}
