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
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/google/logger"
	"github.com/vmihailenco/msgpack"
)

type BlockMsg struct {
	ID          int64
	Compression int64
	Signature   []byte
	Hash        []byte
	Body        []byte
}

func ProcessBlock(block BlockInfo) {
	var (
		sign, body []byte
		err        error
		idCompress int64
	)
	if cfg.Settings.Compression != COMPRESS_NONE && len(block.Data) > 1024 {
		body = Compression(block.Data)
		if len(body) >= len(block.Data) {
			body = block.Data
		} else {
			idCompress = cfg.Settings.Compression
		}
	} else {
		body = block.Data
	}
	body = PGPEncode(body)
	sign, err = SignECDSA(append([]byte(fmt.Sprintf("%d%x", block.ID, block.Hash)), body...))
	if err != nil {
		logger.Fatal(err)
	}
	msg := BlockMsg{
		ID:          block.ID,
		Compression: idCompress,
		Hash:        block.Hash,
		Signature:   sign,
		Body:        body,
	}
	out, err := msgpack.Marshal(msg)
	if err != nil {
		logger.Fatal(err, block.ID)
	}
	fname := filepath.Join(cfg.OutPath, fmt.Sprintf(`%d.block`, block.ID))
	if err = ioutil.WriteFile(fname, out, 0644); err != nil {
		logger.Fatal(err, block.ID)
	}
	StoreBlock(block)
	logger.Info(fmt.Sprintf(`Processed: %d %s`, block.ID, hex.EncodeToString(block.Hash)))
	if PGPPrivate != nil { // Verify
		var (
			verify   []byte
			checkMsg BlockMsg
		)
		if verify, err = ioutil.ReadFile(fname); err != nil {
			logger.Fatal(err)
		}
		err := msgpack.Unmarshal(verify, &checkMsg)
		if err != nil {
			logger.Fatal(err)
		}
		ok, err := CheckECDSA(append([]byte(fmt.Sprintf("%d%x", checkMsg.ID, checkMsg.Hash)),
			checkMsg.Body...), checkMsg.Signature)
		if err != nil || !ok {
			logger.Fatal(`Verifying signature: `, ok, err)
		}
		checkMsg.Body = PGPDecode(checkMsg.Body)
		if checkMsg.ID != block.ID || !bytes.Equal(checkMsg.Hash, block.Hash) {
			logger.Fatal(checkMsg.ID, checkMsg.Hash)
		}
		if checkMsg.Compression != COMPRESS_NONE {
			checkMsg.Body = Decompression(checkMsg.Compression, checkMsg.Body)
		}
		if !bytes.Equal(checkMsg.Body, block.Data) {
			logger.Fatal(errors.New(`Block data is different`))
		}
		logger.Info(`Verified: OK`)
	}
}
