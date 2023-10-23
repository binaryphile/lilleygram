package opt

var (
	NoInt64 = Int64{}

	NoUint32 = Uint32{}
)

type (
	Int64 = Value[int64]

	Uint32 = Value[uint32]
)
