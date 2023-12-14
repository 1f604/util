package util_test

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"hash"
	"testing"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/blake2s"
	"golang.org/x/crypto/sha3"
)

func benchmarkRun(h hash.Hash, i int, b *testing.B) {
	bs := make([]byte, i)
	_, err := rand.Read(bs)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sha256.Sum256(bs)
		//h.Reset()
		//h.Write(bs)
		//h.Sum(nil)
	}

}

func BenchmarkMD5_16b(b *testing.B) {
	benchmarkRun(md5.New(), 16, b)
}

func BenchmarkMD5_32b(b *testing.B) {
	benchmarkRun(md5.New(), 32, b)
}

func BenchmarkMD5_64b(b *testing.B) {
	benchmarkRun(md5.New(), 64, b)
}

func BenchmarkMD5_140b(b *testing.B) {
	benchmarkRun(md5.New(), 140, b)
}

func BenchmarkSHA1_16b(b *testing.B) {
	benchmarkRun(sha1.New(), 16, b)
}

func BenchmarkSha1_32b(b *testing.B) {
	benchmarkRun(sha1.New(), 32, b)
}

func BenchmarkSha1_64b(b *testing.B) {
	benchmarkRun(sha1.New(), 64, b)
}

func BenchmarkSha1_140b(b *testing.B) {
	benchmarkRun(sha1.New(), 140, b)
}

func BenchmarkSha256_16b(b *testing.B) {
	benchmarkRun(sha256.New(), 16, b)
}

func BenchmarkSha256_32b(b *testing.B) {
	benchmarkRun(sha256.New(), 32, b)
}

func BenchmarkSha256_64b(b *testing.B) {
	benchmarkRun(sha256.New(), 64, b)
}

func BenchmarkSha256_140b(b *testing.B) {
	benchmarkRun(sha256.New(), 140, b)
}

func BenchmarkSHA3_16b(b *testing.B) {
	benchmarkRun(sha3.New224(), 16, b)
}

func BenchmarkSHA3_32b(b *testing.B) {
	benchmarkRun(sha3.New224(), 32, b)
}

func BenchmarkSHA3_64b(b *testing.B) {
	benchmarkRun(sha3.New224(), 64, b)
}

func BenchmarkSHA3_140b(b *testing.B) {
	benchmarkRun(sha3.New224(), 140, b)
}

func BenchmarkBLAKE2b_16b(b *testing.B) {
	h, _ := blake2b.New256(nil)
	benchmarkRun(h, 16, b)
}

func BenchmarkBLAKE2b_32b(b *testing.B) {
	h, _ := blake2b.New256(nil)
	benchmarkRun(h, 32, b)
}

func BenchmarkBLAKE2b_64b(b *testing.B) {
	h, _ := blake2b.New256(nil)
	benchmarkRun(h, 64, b)
}

func BenchmarkBLAKE2b_140b(b *testing.B) {
	h, _ := blake2b.New256(nil)
	benchmarkRun(h, 140, b)
}

func BenchmarkBLAKE2s_16b(b *testing.B) {
	h, _ := blake2s.New256(nil)
	benchmarkRun(h, 16, b)
}

func BenchmarkBLAKE2s_32b(b *testing.B) {
	h, _ := blake2s.New256(nil)
	benchmarkRun(h, 32, b)
}

func BenchmarkBLAKE2s_64b(b *testing.B) {
	h, _ := blake2s.New256(nil)
	benchmarkRun(h, 64, b)
}

func BenchmarkBLAKE2s_140b(b *testing.B) {
	h, _ := blake2s.New256(nil)
	benchmarkRun(h, 140, b)
}
