package cockroach

import (
	"github.com/cortezaproject/corteza-server/store/rdbms"
)

// fieldToColumnTypeCaster handles special ComposeModule field query representations
// @todo Not as elegant as it should be but it'll do the trick until the #2 store iteration
//
// Return parameters:
//   * full cast: query column + datatype cast
//   * field cast tpl: fmt template to get query column
//   * type cast tpl: fmt template to cast the compared to value
func fieldToColumnTypeCaster(field rdbms.ModuleFieldTypeDetector, ident string) (string, string, string, error) {
	return "", "", "", nil
}
