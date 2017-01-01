package inetdata

import (
	"bufio"
	"fmt"
	"github.com/edmonds/golang-mtbl"
	"io"
	"os"
	"regexp"
	"strings"
)

var MTBLCompressionTypes = map[string]int{
	"none":   mtbl.COMPRESSION_NONE,
	"snappy": mtbl.COMPRESSION_SNAPPY,
	"zlib":   mtbl.COMPRESSION_ZLIB,
	"lz4":    mtbl.COMPRESSION_LZ4,
	"lz4hc":  mtbl.COMPRESSION_LZ4HC,
}

var Split_WS = regexp.MustCompile(`\s+`)

func PrintVersion() {
	fmt.Fprintf(os.Stderr, "v%s\n", Version)
}

func ReverseKey(s string) string {
	b := make([]byte, len(s))
	var j int = len(s) - 1
	for i := 0; i <= j; i++ {
		b[j-i] = s[i]
	}
	return string(b)
}

func ReadLines(input *os.File, out chan<- string) error {

	var (
		backbufferSize  = 200000
		frontbufferSize = 50000
		r               = bufio.NewReaderSize(os.Stdin, frontbufferSize)
		buf             []byte
		pred            []byte
		err             error
	)

	if backbufferSize <= frontbufferSize {
		backbufferSize = (frontbufferSize / 3) * 4
	}

	for {
		buf, err = r.ReadSlice('\n')

		if err == bufio.ErrBufferFull {
			if len(buf) == 0 {
				continue
			}

			if pred == nil {
				pred = make([]byte, len(buf), backbufferSize)
				copy(pred, buf)
			} else {
				pred = append(pred, buf...)
			}
			continue
		} else if err == io.EOF && len(buf) == 0 && len(pred) == 0 {
			break
		}

		if len(pred) > 0 {
			buf, pred = append(pred, buf...), pred[:0]
		}

		if len(buf) > 0 && buf[len(buf)-1] == '\n' {
			buf = buf[:len(buf)-1]
		}

		if len(buf) == 0 {
			continue
		}

		out <- string(buf)
	}

	close(out)

	if err != nil && err != io.EOF {
		return err
	}

	return nil
}

func PublicSuffixMap() map[string]bool {
	suffixes := make(map[string]bool, len(Public_Suffixes))
	for i := range Public_Suffixes {
		suffixes[Public_Suffixes[i]] = true
	}
	return suffixes
}

func PublicSuffixFind(name string) (string, error) {
	if len(name) == 0 {
		return "", fmt.Errorf("empty hostname")
	}

	// Exit early if obviously invalid input is received
	if !strings.Contains(name, ".") ||
		strings.Contains(name, " ") ||
		strings.Contains(name, ":") {
		return "", fmt.Errorf("invalid hostname")
	}

	for i := range Public_Suffixes {
		suffix := "." + Public_Suffixes[i]

		if len(suffix) > len(name) {
			continue
		}

		sidx := len(name) - len(suffix)
		if strings.Compare(name[sidx:], suffix) == 0 {
			return Public_Suffixes[i], nil
		}
	}
	return "", fmt.Errorf("domain not found")
}
