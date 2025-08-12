package multi

import (
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"sync"

	"github.com/atlastore/belt/hashx"
)

var readerBufPool = sync.Pool{
	New: func() any { b := make([]byte, 32*1024); return b },
}

func HashString[T hashx.AnyHashAlgorithm](str string, algs ...T) ([]string, error) {
	return Hash([]byte(str), algs...)
}

func Hash[T hashx.AnyHashAlgorithm](data []byte, algs ...T) ([]string, error) {
	hs, err := NewHashers(algs...)
	if err != nil {
		return nil, err
	}

	mw := io.MultiWriter(toWriters(hs.hashers)...)
	if _, err := mw.Write(data); err != nil {
		return nil, err
	}

	return finalizeHashes(hs), nil
}

func HashReader[T hashx.AnyHashAlgorithm](r io.Reader, algs ...T) ([]string, error) {
	hs, err := NewHashers(algs...)
	if err != nil {
		return nil, err
	}

	mw := io.MultiWriter(toWriters(hs.hashers)...)

	buf := readerBufPool.Get().([]byte)
	defer readerBufPool.Put(buf)

	for {
		nr, er := r.Read(buf)
		if nr > 0 {
			if _, ew := mw.Write(buf[:nr]); ew != nil {
				return nil, ew
			}
		}
		if er != nil {
			if er == io.EOF {
				break
			}
			return nil, er
		}
	}


	return finalizeHashes(hs), nil
}

type algoHasher[T hashx.AnyHashAlgorithm] struct {
	algos   []T
	hashers []hash.Hash
}

func (ah *algoHasher[T]) Len() int {
	 if len(ah.algos) != len(ah.hashers) {
		return 0
	 }
	 return len(ah.algos)
}

func NewHashers[T hashx.AnyHashAlgorithm](algs ...T) (algoHasher[T], error) {
	hs := algoHasher[T]{
		algos: make([]T, len(algs)),
		hashers: make([]hash.Hash, len(algs)),
	}
	for i, algo := range algs {
		if hashx.IsKeyed(algo) {
			return algoHasher[T]{}, hashx.ErrKeyRequired
		}
		fn, err := hashx.GetHash(algo)
		if err != nil {
			return algoHasher[T]{}, err
		}
		hs.algos[i] = algo
		hs.hashers[i] = fn()
	}
	return hs, nil
}

func finalizeHashes[T hashx.AnyHashAlgorithm](hs algoHasher[T]) []string {
	results := make([]string, hs.Len())
	for i, h := range hs.hashers {
		sum := h.Sum(nil)
		dst := make([]byte, hex.EncodedLen(len(sum)))
		hex.Encode(dst, sum)
		results[i] = fmt.Sprintf("%s:%s", string(hs.algos[i]), string(dst))
	}
	return results
}

func toWriters(hs []hash.Hash) []io.Writer {
	ws := make([]io.Writer, len(hs))
	for i, h := range hs {
		ws[i] = h
	}
	return ws
}