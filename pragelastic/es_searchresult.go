package pragelastic

import (
	"encoding/json"
	"net/http"
)

type ESSearchResult struct {
	Header          http.Header `json:"-"`
	TookInMillis    int64       `json:"took,omitempty"`             // search time in milliseconds
	TerminatedEarly bool        `json:"terminated_early,omitempty"` // request terminated early
	NumReducePhases int         `json:"num_reduce_phases,omitempty"`
	//Clusters        []*SearchResultCluster `json:"_clusters,omitempty"`    // 6.1.0+
	ScrollId     string          `json:"_scroll_id,omitempty"`   // only used with Scroll and Scan operations
	Hits         *ESSearchHits   `json:"hits,omitempty"`         // the actual search hits
	Suggest      ESSearchSuggest `json:"suggest,omitempty"`      // results from suggesters
	Aggregations ESAggregations  `json:"aggregations,omitempty"` // results from aggregations
	TimedOut     bool            `json:"timed_out,omitempty"`    // true if the search timed out
	Error        *ESErrorDetails `json:"error,omitempty"`        // only used in MultiGet
	//Profile      *SearchProfile `json:"profile,omitempty"`      // profiling results, if optional Profile API was active for this search
	//Shards       *ShardsInfo    `json:"_shards,omitempty"`      // shard information
	Status int `json:"status,omitempty"` // used in MultiSearch
}

type ESSearchHits struct {
	TotalHits *ESTotalHits `json:"total,omitempty"`     // total number of hits found
	MaxScore  *float64     `json:"max_score,omitempty"` // maximum score of all hits
	Hits      []*SearchHit `json:"hits,omitempty"`      // the actual hits returned
}

type ESTotalHits struct {
	Value    int64  `json:"value"`    // value of the total hit count
	Relation string `json:"relation"` // how the value should be interpreted: accurate ("eq") or a lower bound ("gte")
}

type SearchHit struct {
	Score       *float64      `json:"_score,omitempty"`   // computed score
	Index       string        `json:"_index,omitempty"`   // index name
	Type        string        `json:"_type,omitempty"`    // type meta field
	Id          string        `json:"_id,omitempty"`      // external or internal
	Uid         string        `json:"_uid,omitempty"`     // uid meta field (see MapperService.java for all meta fields)
	Routing     string        `json:"_routing,omitempty"` // routing meta field
	Parent      string        `json:"_parent,omitempty"`  // parent meta field
	Version     *int64        `json:"_version,omitempty"` // version number, when Version is set to true in SearchService
	SeqNo       *int64        `json:"_seq_no"`
	PrimaryTerm *int64        `json:"_primary_term"`
	Sort        []interface{} `json:"sort,omitempty"` // sort information
	//Highlight      SearchHitHighlight     `json:"highlight,omitempty"`       // highlighter information
	Source json.RawMessage        `json:"_source,omitempty"` // stored document source
	Fields map[string]interface{} `json:"fields,omitempty"`  // returned (stored) fields
	//Explanation    *SearchExplanation     `json:"_explanation,omitempty"`    // explains how the score was computed
	MatchedQueries []string `json:"matched_queries,omitempty"` // matched queries
	//InnerHits      map[string]*SearchHitInnerHits `json:"inner_hits,omitempty"`      // inner hits with ES >= 1.5.0
	//Nested         *NestedHit                     `json:"_nested,omitempty"`         // for nested inner hits
	Shard string `json:"_shard,omitempty"` // used e.g. in Search Explain
	Node  string `json:"_node,omitempty"`  // used e.g. in Search Explain

}

// Suggest

// SearchSuggest is a map of suggestions.
// See https://www.elastic.co/guide/en/elasticsearch/reference/7.0/search-suggesters.html.
type ESSearchSuggest map[string][]ESSearchSuggestion

// SearchSuggestion is a single search suggestion.
// See https://www.elastic.co/guide/en/elasticsearch/reference/7.0/search-suggesters.html.
type ESSearchSuggestion struct {
	Text    string                     `json:"text"`
	Offset  int                        `json:"offset"`
	Length  int                        `json:"length"`
	Options []ESSearchSuggestionOption `json:"options"`
}

// SearchSuggestionOption is an option of a SearchSuggestion.
// See https://www.elastic.co/guide/en/elasticsearch/reference/7.0/search-suggesters.html.
type ESSearchSuggestionOption struct {
	Text            string              `json:"text"`
	Index           string              `json:"_index"`
	Type            string              `json:"_type"`
	Id              string              `json:"_id"`
	Score           float64             `json:"score"`  // term and phrase suggesters uses "score" as of 6.2.4
	ScoreUnderscore float64             `json:"_score"` // completion and context suggesters uses "_score" as of 6.2.4
	Highlighted     string              `json:"highlighted"`
	CollateMatch    bool                `json:"collate_match"`
	Freq            int                 `json:"freq"` // from TermSuggestion.Option in Java API
	Source          json.RawMessage     `json:"_source"`
	Contexts        map[string][]string `json:"contexts,omitempty"`
}
