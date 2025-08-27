package gast

// CommentPosition contains resolved start/end coordinates for the comment (1-based).
type CommentPosition struct {
	StartLine int // 0-based
	StartCol  int // 0-based (column of first character)
	EndLine   int // 0-based
	EndCol    int // 0-based (column of last character + 1? we keep it as column of last char)
}

// CommentNode is a comment with attached position metadata.
// (This is what gast helpers should produce.)
type CommentNode struct {
	Text     string
	Position CommentPosition
	Index    int // index within the comment list (useful to detect ordering)
}

type CommentBlock struct {
	Comments  []CommentNode
	FileName  string
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
}
