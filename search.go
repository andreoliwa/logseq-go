package logseq

import (
	"time"

	"github.com/aholstenson/logseq-go/internal/indexing"
)

// SearchResults is a result set from a search.
type SearchResults[R any] interface {
	// Size is the number of results available in this result set.
	Size() int

	// Count is the number of results that are available in total. For the
	// number of results available via Results, use Size.
	Count() int

	// Results is a slice of all the results in this result set.
	Results() []R
}

// SearchOption is an option for doing a search.
type SearchOption func(*searchOptions)

type searchOptions struct {
	query Query

	size int
	from int

	sortBy []indexing.SortField
}

// WithMaxHits sets the maximum number of hits to return. The default is 10.
func WithMaxHits(n int) SearchOption {
	return func(o *searchOptions) {
		o.size = n
	}
}

// FromHit sets the offset to start returning results from. This can be used
// for pagination.
func FromHit(n int) SearchOption {
	return func(o *searchOptions) {
		o.from = n
	}
}

// WithQuery sets the query to use for the search. If no query is set the
// default is to match everything. This option can be used multiple times in
// which case the queries are combined with a logical AND.
func WithQuery(q Query) SearchOption {
	return func(o *searchOptions) {
		if o.query == nil {
			o.query = q
		} else {
			o.query = And(o.query, q)
		}
	}
}

type searchResultsImpl[R any] struct {
	size    int
	count   int
	results []R
}

func (s *searchResultsImpl[R]) Size() int {
	return s.size
}

func (s *searchResultsImpl[R]) Count() int {
	return s.count
}

func (s *searchResultsImpl[R]) Results() []R {
	return s.results
}

func newSearchResults[I any, O any](r indexing.SearchResults[I], mapper func(I) O) SearchResults[O] {
	results := make([]O, len(r.Results()))
	for i, r := range r.Results() {
		results[i] = mapper(r)
	}
	return &searchResultsImpl[O]{
		size:    r.Size(),
		count:   r.Count(),
		results: results,
	}
}

type PageResult interface {
	// Type returns the type of the page.
	Type() PageType

	// Title returns the title of the page.
	Title() string

	// Date returns the date if this page is a journal.
	Date() time.Time

	// Open the page.
	Open() (Page, error)
}

type pageResultImpl struct {
	graph *Graph

	docType PageType
	title   string
	date    time.Time
	opener  func() (Page, error)
}

func (d *pageResultImpl) Type() PageType {
	return d.docType
}

func (d *pageResultImpl) Title() string {
	return d.title
}

func (d *pageResultImpl) Date() time.Time {
	return d.date
}

func (d *pageResultImpl) Open() (Page, error) {
	return d.opener()
}
