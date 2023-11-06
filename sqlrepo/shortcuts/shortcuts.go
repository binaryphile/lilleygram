package shortcuts

import (
	"github.com/doug-martin/goqu/v9"
)

var (
	COUNT = goqu.COUNT

	COALESCE = goqu.COALESCE

	I = goqu.I

	On = goqu.On

	Or = goqu.Or

	T = goqu.T

	V = goqu.V
)

type (
	Database = goqu.Database

	Ex = goqu.Ex

	ExOr = goqu.ExOr

	Op = goqu.Op

	Record = goqu.Record
)
