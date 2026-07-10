package sqlcompleteservice

// CandidateKind 补全候选类型。
type CandidateKind string

const (
	KindKeyword   CandidateKind = "keyword"
	KindTable     CandidateKind = "table"
	KindView      CandidateKind = "view"
	KindColumn    CandidateKind = "column"
	KindFunction  CandidateKind = "function"
	KindSchema    CandidateKind = "schema"
	KindDatabase  CandidateKind = "database"
)

// CompleteCandidate 单个补全候选。
type CompleteCandidate struct {
	Label      string
	Kind       CandidateKind
	Detail     string
	InsertText string
	SortText   string
}

// CompleteResult 补全结果。
type CompleteResult struct {
	Candidates []CompleteCandidate
}
