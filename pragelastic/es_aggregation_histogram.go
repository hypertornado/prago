package pragelastic

type ESHistogramAggregation struct {
	field string
	//script          *Script
	missing         interface{}
	subAggregations map[string]ESAggregation
	meta            map[string]interface{}

	interval    float64
	order       string
	orderAsc    bool
	minDocCount *int64
	minBounds   *float64
	maxBounds   *float64
	offset      *float64
}

func NewESHistogramAggregation() *ESHistogramAggregation {
	return &ESHistogramAggregation{
		subAggregations: make(map[string]ESAggregation),
	}
}

func (a *ESHistogramAggregation) Field(field string) *ESHistogramAggregation {
	a.field = field
	return a
}

/*func (a *HistogramAggregation) Script(script *Script) *HistogramAggregation {
	a.script = script
	return a
}*/

// Missing configures the value to use when documents miss a value.
func (a *ESHistogramAggregation) Missing(missing interface{}) *ESHistogramAggregation {
	a.missing = missing
	return a
}

func (a *ESHistogramAggregation) SubAggregation(name string, subAggregation ESAggregation) *ESHistogramAggregation {
	a.subAggregations[name] = subAggregation
	return a
}

// Meta sets the meta data to be included in the aggregation response.
func (a *ESHistogramAggregation) Meta(metaData map[string]interface{}) *ESHistogramAggregation {
	a.meta = metaData
	return a
}

// Interval for this builder, must be greater than 0.
func (a *ESHistogramAggregation) Interval(interval float64) *ESHistogramAggregation {
	a.interval = interval
	return a
}

// Order specifies the sort order. Valid values for order are:
// "_key", "_count", a sub-aggregation name, or a sub-aggregation name
// with a metric.
func (a *ESHistogramAggregation) Order(order string, asc bool) *ESHistogramAggregation {
	a.order = order
	a.orderAsc = asc
	return a
}

func (a *ESHistogramAggregation) OrderByCount(asc bool) *ESHistogramAggregation {
	// "order" : { "_count" : "asc" }
	a.order = "_count"
	a.orderAsc = asc
	return a
}

func (a *ESHistogramAggregation) OrderByCountAsc() *ESHistogramAggregation {
	return a.OrderByCount(true)
}

func (a *ESHistogramAggregation) OrderByCountDesc() *ESHistogramAggregation {
	return a.OrderByCount(false)
}

func (a *ESHistogramAggregation) OrderByKey(asc bool) *ESHistogramAggregation {
	// "order" : { "_key" : "asc" }
	a.order = "_key"
	a.orderAsc = asc
	return a
}

func (a *ESHistogramAggregation) OrderByKeyAsc() *ESHistogramAggregation {
	return a.OrderByKey(true)
}

func (a *ESHistogramAggregation) OrderByKeyDesc() *ESHistogramAggregation {
	return a.OrderByKey(false)
}

// OrderByAggregation creates a bucket ordering strategy which sorts buckets
// based on a single-valued calc get.
func (a *ESHistogramAggregation) OrderByAggregation(aggName string, asc bool) *ESHistogramAggregation {
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
	a.order = aggName
	a.orderAsc = asc
	return a
}

// OrderByAggregationAndMetric creates a bucket ordering strategy which
// sorts buckets based on a multi-valued calc get.
func (a *ESHistogramAggregation) OrderByAggregationAndMetric(aggName, metric string, asc bool) *ESHistogramAggregation {
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
	a.order = aggName + "." + metric
	a.orderAsc = asc
	return a
}

func (a *ESHistogramAggregation) MinDocCount(minDocCount int64) *ESHistogramAggregation {
	a.minDocCount = &minDocCount
	return a
}

func (a *ESHistogramAggregation) ExtendedBounds(min, max float64) *ESHistogramAggregation {
	a.minBounds = &min
	a.maxBounds = &max
	return a
}

func (a *ESHistogramAggregation) ExtendedBoundsMin(min float64) *ESHistogramAggregation {
	a.minBounds = &min
	return a
}

func (a *ESHistogramAggregation) MinBounds(min float64) *ESHistogramAggregation {
	a.minBounds = &min
	return a
}

func (a *ESHistogramAggregation) ExtendedBoundsMax(max float64) *ESHistogramAggregation {
	a.maxBounds = &max
	return a
}

func (a *ESHistogramAggregation) MaxBounds(max float64) *ESHistogramAggregation {
	a.maxBounds = &max
	return a
}

// Offset into the histogram
func (a *ESHistogramAggregation) Offset(offset float64) *ESHistogramAggregation {
	a.offset = &offset
	return a
}

func (a *ESHistogramAggregation) Source() (interface{}, error) {
	// Example:
	// {
	//     "aggs" : {
	//         "prices" : {
	//             "histogram" : {
	//                 "field" : "price",
	//                 "interval" : 50
	//             }
	//         }
	//     }
	// }
	//
	// This method returns only the { "histogram" : { ... } } part.

	source := make(map[string]interface{})
	opts := make(map[string]interface{})
	source["histogram"] = opts

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
	}
	if a.missing != nil {
		opts["missing"] = a.missing
	}*/

	opts["interval"] = a.interval
	if a.order != "" {
		o := make(map[string]interface{})
		if a.orderAsc {
			o[a.order] = "asc"
		} else {
			o[a.order] = "desc"
		}
		opts["order"] = o
	}
	if a.offset != nil {
		opts["offset"] = *a.offset
	}
	if a.minDocCount != nil {
		opts["min_doc_count"] = *a.minDocCount
	}
	if a.minBounds != nil || a.maxBounds != nil {
		bounds := make(map[string]interface{})
		if a.minBounds != nil {
			bounds["min"] = a.minBounds
		}
		if a.maxBounds != nil {
			bounds["max"] = a.maxBounds
		}
		opts["extended_bounds"] = bounds
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
