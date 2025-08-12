package multi

import (
	"io"
)

func (ah *algoHasher[T]) Writers() []io.Writer {
	return toWriters(ah.hashers)
}

func (ah *algoHasher[T]) Sum() []string {
	return finalizeHashes(*ah)
}