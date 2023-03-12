package running

type ctxKey string

var CtxKey ctxKey = "rck"

type CtxParams struct {
	SkipNodes []string

	SkipOnCtxErr bool

	// MatchAllLabels the nodes with all specified labels will run.
	// Does not work for nodes without labels
	MatchAllLabels []string

	// MatchOneOfLabels the nodes with one of the specified labels will run.
	// Does not work for nodes without labels
	MatchOneOfLabels []string

	State State
}
