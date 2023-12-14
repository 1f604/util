package util_test

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"math/rand"
	"regexp"
	"strconv"
	"testing"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/sha3"
)

func BenchmarkChecksums(b *testing.B) {
	sizes := []string{"16B", "32B", "64B", "140B"}
	fodderBySize := map[string][]byte{}
	for _, size := range sizes {
		fodderBySize[size] = generateRandomBytes(b, size)
	}

	b.Run("cryptographic", func(b *testing.B) {
		for _, size := range sizes {
			fodder := fodderBySize[size]
			for _, hi := range cryptographic {
				b.Run(fmt.Sprintf("%s-%s", hi.hashName, size), func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						h := hi.hashNewFunc()
						h.Write(fodder)
						h.Sum(nil)
					}
				})
			}
		}
	})

}

var cryptographic = []struct {
	hashName    string
	hashNewFunc func() hash.Hash
}{{
	hashName:    "md5",
	hashNewFunc: md5.New,
}, {
	hashName:    "sha1",
	hashNewFunc: sha1.New,
}, {
	hashName:    "sha256",
	hashNewFunc: sha256.New,
}, {
	hashName:    "sha512",
	hashNewFunc: sha512.New,
}, {
	hashName:    "sha3-256",
	hashNewFunc: sha3.New256,
}, {
	hashName:    "sha3-512",
	hashNewFunc: sha3.New512,
}, {
	hashName:    "sha512-224",
	hashNewFunc: sha512.New512_224,
}, {
	hashName:    "sha512-256",
	hashNewFunc: sha512.New512_256,
}, {
	hashName: "blake2b-256",
	hashNewFunc: func() hash.Hash {
		h, _ := blake2b.New256(nil)
		return h
	},
}, {
	hashName: "blake2b-512",
	hashNewFunc: func() hash.Hash {
		h, _ := blake2b.New512(nil)
		return h
	},
}}

var stringToSize = regexp.MustCompile(`(\d)([KMG]?)B`)

func generateRandomBytes(tb testing.TB, sizeString string) []byte {
	ss := stringToSize.FindStringSubmatch(sizeString)
	if len(ss) != 3 {
		tb.Fatalf("invalid size string %q", ss)
	}
	size, err := strconv.Atoi(ss[1])
	if err != nil {
		tb.Fatalf("failed to parse integer %q", ss[1])
	}
	switch ss[2] {
	case "G":
		size *= 1024
		fallthrough
	case "M":
		size *= 1024
		fallthrough
	case "K":
		size *= 1024
	}
	s := make([]byte, size)
	_, _ = rand.Read(s)
	return s
}
