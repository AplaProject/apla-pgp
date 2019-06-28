// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"time"

	"github.com/google/logger"
)

const (
	COMPRESS_NONE = iota
	COMPRESS_GZIP
)

func Compression(in []byte) []byte {
	var buf bytes.Buffer
	switch cfg.Settings.Compression {
	case COMPRESS_GZIP:
		zw := gzip.NewWriter(&buf)
		zw.Name = "block"
		zw.ModTime = time.Now()
		_, err := zw.Write(in)
		if err != nil {
			logger.Fatal(err)
		}
		if err := zw.Close(); err != nil {
			logger.Fatal(err)
		}
	default:
		logger.Fatal(`Unknown compression: `, cfg.Settings.Compression)
	}
	return buf.Bytes()
}

func Decompression(method int64, in []byte) []byte {
	var bufout bytes.Buffer
	buf := bytes.NewBuffer(in)
	switch method {
	case COMPRESS_GZIP:
		zr, err := gzip.NewReader(buf)
		if err != nil {
			logger.Fatal(err)
		}
		if _, err := io.Copy(&bufout, zr); err != nil {
			logger.Fatal(err)
		}
		if err := zr.Close(); err != nil {
			logger.Fatal(err)
		}
	default:
		logger.Fatal(`Unknown compression: `, method)
	}
	return bufout.Bytes()
}
