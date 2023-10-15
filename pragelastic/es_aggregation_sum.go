package pragelastic

type ESSumAggregation struct {
	field string
	//script          *Script
	format          string
	missing         interface{}
	subAggregations map[string]ESAggregation
	meta            map[string]interface{}
}

func NewESSumAggregation() *ESSumAggregation {
	return &ESSumAggregation{
		subAggregations: make(map[string]ESAggregation),
	}
}

func (a *ESSumAggregation) Field(field string) *ESSumAggregation {
	a.field = field
	return a
}

/*func (a *SumAggregation) Script(script *Script) *SumAggregation {
	a.script = script
	return a
}*/

func (a *ESSumAggregation) Format(format string) *ESSumAggregation {
	a.format = format
	return a
}

func (a *ESSumAggregation) Missing(missing interface{}) *ESSumAggregation {
	a.missing = missing
	return a
}

func (a *ESSumAggregation) SubAggregation(name string, subAggregation ESAggregation) *ESSumAggregation {
	a.subAggregations[name] = subAggregation
	return a
}

// Meta sets the meta data to be included in the aggregation response.
func (a *ESSumAggregation) Meta(metaData map[string]interface{}) *ESSumAggregation {
	a.meta = metaData
	return a
}

func (a *ESSumAggregation) Source() (interface{}, error) {
	// Example:
	//	{
	//    "aggs" : {
	//      "intraday_return" : { "sum" : { "field" : "change" } }
	//    }
	//	}
	// This method returns only the { "sum" : { "field" : "change" } } part.

	source := make(map[string]interface{})
	opts := make(map[string]interface{})
	source["sum"] = opts

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
	if a.format != "" {
		opts["format"] = a.format
	}
	if a.missing != nil {
		opts["missing"] = a.missing
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
