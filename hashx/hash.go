package hashx

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"hash/crc32"
	"hash/crc64"
	"hash/fnv"
	"io"
	"reflect"
	"strings"

	"github.com/cespare/xxhash/v2"

	"github.com/minio/highwayhash"
	"github.com/zeebo/blake3"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/blake2s"
	"golang.org/x/crypto/md4"
	"golang.org/x/crypto/sha3"
)

var hashMap =  make(map[string]reflect.Type)

type HashAlgorithm string
type Hash32Algorithm string
type Hash64Algorithm string
type KeyedHashAlgorithm string

var hashAlgorithmType = reflect.TypeOf(HashAlgorithm(""))
var hash32AlgorithmType = reflect.TypeOf(Hash32Algorithm(""))
var hash64AlgorithmType = reflect.TypeOf(Hash64Algorithm(""))
var keyedHashAlgorithmType = reflect.TypeOf(KeyedHashAlgorithm(""))

func init() {
	// HashAlgorithm
	hashMap[string(SHA1)] = hashAlgorithmType
	hashMap[string(SHA224)] = hashAlgorithmType
	hashMap[string(SHA256)] = hashAlgorithmType
	hashMap[string(SHA384)] = hashAlgorithmType
	hashMap[string(SHA512)] = hashAlgorithmType
	hashMap[string(SHA3_224)] = hashAlgorithmType
	hashMap[string(SHA3_256)] = hashAlgorithmType
	hashMap[string(SHA3_384)] = hashAlgorithmType
	hashMap[string(SHA3_512)] = hashAlgorithmType
	hashMap[string(MD4)] = hashAlgorithmType
	hashMap[string(MD5)] = hashAlgorithmType
	hashMap[string(XXHash)] = hashAlgorithmType

	// KeyedHashAlgorithm
	hashMap[string(Blake2s)] = keyedHashAlgorithmType
	hashMap[string(Blake2b)] = keyedHashAlgorithmType
	hashMap[string(Blake3)] = keyedHashAlgorithmType
	hashMap[string(Highway)] = keyedHashAlgorithmType
	hashMap[string(HMAC_SHA256)] = keyedHashAlgorithmType
	hashMap[string(HMAC_SHA512)] = keyedHashAlgorithmType

	// Hash32Algorithm
	hashMap[string(CRC32)] = hash32AlgorithmType
	hashMap[string(FNV32)] = hash32AlgorithmType

	// Hash64Algorithm
	hashMap[string(CRC64)] = hash64AlgorithmType
	hashMap[string(FNV64)] = hash64AlgorithmType
}


const (
	// SHA Family
	SHA1      HashAlgorithm = "sha1"
	SHA224    HashAlgorithm = "sha224"
	SHA256    HashAlgorithm = "sha256"
	SHA384    HashAlgorithm = "sha384"
	SHA512    HashAlgorithm = "sha512"
	SHA3_224  HashAlgorithm = "sha3-224"
	SHA3_256  HashAlgorithm = "sha3-256"
	SHA3_384  HashAlgorithm = "sha3-384"
	SHA3_512  HashAlgorithm = "sha3-512"

	// MD Family
	MD4  HashAlgorithm = "md4"
	MD5  HashAlgorithm = "md5"

	// Blake Family, keyed
	Blake2s KeyedHashAlgorithm = "blake2s"
	// Blake Family, keyed
	Blake2b KeyedHashAlgorithm = "blake2b"
	// Blake Family, keyed
	Blake3  KeyedHashAlgorithm = "blake3"

	// Others
	CRC32       Hash32Algorithm = "crc32"
	CRC64       Hash64Algorithm = "crc64"
	FNV32       Hash32Algorithm = "fnv32"
	FNV64       Hash64Algorithm = "fnv64"
	XXHash      HashAlgorithm = "xxhash"
	
	// Keyed
	HMAC_SHA256 KeyedHashAlgorithm = "hmac-sha256"
	// Keyed
	HMAC_SHA512 KeyedHashAlgorithm = "hmac-sha512"
	// Keyed
	Highway     KeyedHashAlgorithm = "highway"
)

type HashSum struct {
	data []byte
	algo string
}

var ErrUnsupported = errors.New("hashx: unsupported hashing algorithm")
var ErrKeyRequired = errors.New("hashx: key is required")
var ErrWrongKeyLength = func(desiredLen, len int) error {
	return fmt.Errorf("hashx: key wrong length, expected: %v but go %v", desiredLen, len)
}

func HashBytes[T AnyHashAlgorithm](algo T, data []byte, key ...[]byte) (*HashSum, error) {
	fn, err := GetHash(algo, key...)
	if err != nil || fn == nil {
		return nil, err
	}
	h := fn()
	h.Write(data)
	return &HashSum{
		data: h.Sum(nil),
		algo: string(algo),
	}, nil
}

func HashString[T AnyHashAlgorithm](algo T, str string, key ...[]byte) (*HashSum, error) {
	return HashBytes(algo, []byte(str), key...)
}

func HashReader[T AnyHashAlgorithm](algo T, r io.Reader, key ...[]byte) (*HashSum, error) {
	fn, err := GetHash(algo, key...)
	if err != nil || fn == nil {
		return nil, err
	}

	h := fn()
	_, err = io.Copy(h, r)
	if err != nil {
		return nil, err
	}

	return &HashSum{
		data: h.Sum(nil),
		algo: string(algo),
	}, nil
}

func (h HashSum) Encode() string {
	return fmt.Sprintf("%s:%s", h.algo, hex.EncodeToString(h.data))
}

func (algo32 Hash32Algorithm) HashBytes(data []byte) uint32 {
	fn, _ := getHash32Func(algo32)


	h := fn()

	h.Write(data)

	if h32, ok := h.(hash.Hash32); ok {
		return h32.Sum32()
	}

	return 0
}

func (algo32 Hash32Algorithm) HashString(str string) uint32 {
	return algo32.HashBytes([]byte(str))
}

func (algo32 Hash32Algorithm) HashReader(r io.Reader) (uint32, error) {
	fn, err := getHash32Func(algo32)

	if err != nil {
		return 0, err
	}

	h := fn()

	_, err = io.Copy(h, r)
	if err != nil {
		return 0, err
	}

	if h32, ok := h.(hash.Hash32); ok {
		return h32.Sum32(), nil
	}

	return 0, ErrUnsupported
}

func (algo64 Hash64Algorithm) HashBytes(data []byte) uint64 {
	fn, _ := getHash64Func(algo64)

	h := fn()

	h.Write(data)

	if h32, ok := h.(hash.Hash64); ok {
		return h32.Sum64()
	}

	return 0
}

func (algo64 Hash64Algorithm) HashString(str string) uint64 {
	return algo64.HashBytes([]byte(str))
}

func (algo64 Hash64Algorithm) HashReader(r io.Reader) (uint64, error) {
	fn, _ := getHash64Func(algo64)

	h := fn()

	_, err := io.Copy(h, r)
	if err != nil {
		return 0, err
	}

	if h32, ok := h.(hash.Hash64); ok {
		return h32.Sum64(), nil
	}

	return 0, ErrUnsupported
}


func GetHash[T AnyHashAlgorithm](algo T, key ...[]byte) (func() hash.Hash, error) {
	algoStr := string(algo)
	alg, ok := hashMap[algoStr]
	if !ok {
		return nil, ErrUnsupported
	}

	switch alg {
	case hashAlgorithmType:
		return getHashFunc(HashAlgorithm(algoStr))
	case keyedHashAlgorithmType:
		if len(key) == 0 || len(key[0]) == 0 {
			return nil, ErrKeyRequired
		}
		return getKeyedHashFunc(KeyedHashAlgorithm(algoStr), key[0])
	case hash32AlgorithmType:
		return getHash32Func(Hash32Algorithm(algoStr))
	case hash64AlgorithmType:
		return getHash64Func(Hash64Algorithm(algoStr))
	default:
		return nil, ErrUnsupported
	}
}


func getHashFunc(algo HashAlgorithm) (func() hash.Hash, error) {
	switch normalizeAlgName[HashAlgorithm](string(algo)){
	case SHA1:
		return sha1.New, nil
	case SHA224:
		return sha256.New224, nil
	case SHA256:
		return sha256.New, nil
	case SHA384:
		return sha512.New384, nil
	case SHA512:
		return sha512.New, nil
	case SHA3_224:
		return sha3.New224, nil
	case SHA3_384: 
		return sha3.New384, nil
	case SHA3_512:
		return sha3.New512, nil
	case MD4:
		return md4.New, nil
	case MD5:
		return md5.New, nil
	case XXHash:
		return func() hash.Hash {
			return xxhash.New()
		}, nil
	default:
		return nil, ErrUnsupported
	}
}

func getKeyedHashFunc(algo KeyedHashAlgorithm, key []byte) (func() hash.Hash, error) {
	switch normalizeAlgName[KeyedHashAlgorithm](string(algo)) {
	case HMAC_SHA256:
		return func() hash.Hash {
			return hmac.New(sha256.New, key)
		}, nil
	case HMAC_SHA512:
		return func() hash.Hash {
			return hmac.New(sha512.New, key)
		}, nil
	case Blake2b:
		if len(key) > 64 {
			return nil, ErrWrongKeyLength(64, len(key))
		}
		return func() hash.Hash {
			h, _ := blake2b.New512(key)
			return h
		}, nil
	case Blake2s:
		if len(key) > 64 {
			return nil, ErrWrongKeyLength(64, len(key))
		}
		return func() hash.Hash {
			h, _ := blake2s.New256(key)
			return h
		}, nil
	case Blake3:
			if len(key) != 32 {
			return nil, ErrWrongKeyLength(32, len(key))
		}
		return func() hash.Hash {
			h, _ := blake3.NewKeyed(key)
			return h
		}, nil
	case Highway:
			if len(key) != 32 {
			return nil, ErrWrongKeyLength(32, len(key))
		}
		return func() hash.Hash {
			h, _ := highwayhash.New(key)
			return h
		}, nil
	default:
		return nil, ErrUnsupported
	}
}

func getHash32Func(algo Hash32Algorithm) (func() hash.Hash, error) {
	switch normalizeAlgName[Hash32Algorithm](string(algo)) {
	case FNV32:
		return func() hash.Hash {
			return hash.Hash(fnv.New32a())
		}, nil
	case CRC32:
		return func() hash.Hash {
			return hash.Hash(crc32.NewIEEE())
		}, nil
	default:
		return nil, ErrUnsupported
	}
}

func getHash64Func(algo Hash64Algorithm) (func() hash.Hash, error) {
	switch normalizeAlgName[Hash64Algorithm](string(algo)) {
	case FNV64:
		return func() hash.Hash {
			return fnv.New64a()
		}, nil
	case CRC64:
		return func() hash.Hash {
			return crc64.New(crc64.MakeTable(crc64.ISO))
		}, nil
	default:
			return nil, ErrUnsupported
	}
}

type AnyHashAlgorithm interface {
	~string
}

func normalizeAlgName[T AnyHashAlgorithm](a string) T {
	return T(strings.ToLower(strings.ReplaceAll(a, "_", "-")))
}

func IsKeyed[T AnyHashAlgorithm](algo T) bool {
	alg, ok := hashMap[string(algo)]
	if !ok {
		return false
	}

	switch alg {
	case keyedHashAlgorithmType:
		return true
	default:
		return false
	}
}