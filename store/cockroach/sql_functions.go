package cockroach

import (
	"github.com/cortezaproject/corteza-server/pkg/ql"
)

func sqlFunctionHandler(f ql.Function) (ql.ASTNode, error) {
	// @todo ^^
	return f, nil
}
