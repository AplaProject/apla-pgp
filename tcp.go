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
	"context"
	"crypto/sha256"

	"github.com/AplaProject/go-apla/packages/network/tcpclient"
	"github.com/google/logger"
)

func GetLastBlock() (id int64) {
	blockID, err := tcpclient.GetMaxBlockID(cfg.TCP.Host)
	if err != nil {
		logger.Fatal(err)
	}
	return blockID
}

func BinToDec(bin []byte) int64 {
	var a uint64
	l := len(bin)
	for i, b := range bin {
		shift := uint64((l - i - 1) * 8)
		a |= uint64(b) << shift
	}
	return int64(a)
}

func TCPLoadBlocks() {
	var skip bool
	ctx := context.Background()
	ctxDone, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
	}()
	from := LastID() + 1
	if from > 1 {
		from--
	}
	rawBlocksChan, err := tcpclient.GetBlocksBodies(ctxDone, cfg.TCP.Host, from, false)
	if err != nil {
		logger.Error(err)
		return
	}
	for rawBlock := range rawBlocksChan {
		if skip {
			continue
		}
		buf := bytes.NewBuffer(rawBlock)
		buf.Next(2)
		blockID := BinToDec(buf.Next(4))
		hash := sha256.Sum256(rawBlock)
		if blockID == from && from > 1 {
			if !bytes.Equal(hash[:], GetHash(from)) {
				if from -= 10; from < 0 {
					from = 0
				}
				store.Set(lastId, from)
				skip = true
			} else {
				continue
			}
		}
		ProcessBlock(BlockInfo{ID: blockID, Hash: hash[:], Data: rawBlock})
	}
}
