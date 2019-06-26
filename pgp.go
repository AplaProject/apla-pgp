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
	"io/ioutil"
	"os"

	"github.com/google/logger"
	"golang.org/x/crypto/openpgp"
)

// create gpg keys with
// $ gpg --gen-key
// ensure you correct paths and passphrase

var (
	PGPPrivate openpgp.EntityList
	PGPPublic  openpgp.EntityList
)

func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func ReadPGPKeys() (out string) {
	keyringPubFile, err := os.Open(cfg.PGP.Path + `/pubring.gpg`)
	if err != nil {
		logger.Fatal(err)
	}
	defer keyringPubFile.Close()
	PGPPublic, err = openpgp.ReadKeyRing(keyringPubFile)
	if err != nil {
		logger.Fatal(err)
	}
	out = `Public`
	privKey := cfg.PGP.Path + `/secring.gpg`
	if FileExists(privKey) {
		var entity *openpgp.Entity

		keyringPrivFile, err := os.Open(privKey)
		if err != nil {
			logger.Fatal(err)
		}
		defer keyringPrivFile.Close()
		PGPPrivate, err = openpgp.ReadKeyRing(keyringPrivFile)
		if err != nil {
			logger.Fatal(err)
		}
		entity = PGPPrivate[0]

		passphraseByte := []byte(cfg.PGP.Phrase)
		entity.PrivateKey.Decrypt(passphraseByte)
		for _, subkey := range entity.Subkeys {
			subkey.PrivateKey.Decrypt(passphraseByte)
		}
		out += ` & Private`
	}
	return out
}

func PGPEncode(in []byte) []byte {
	buf := new(bytes.Buffer)
	w, err := openpgp.Encrypt(buf, PGPPublic, nil, nil, nil)
	if err != nil {
		logger.Fatal(err)
	}
	_, err = w.Write(in)
	if err != nil {
		logger.Fatal(err)
	}
	if err = w.Close(); err != nil {
		logger.Fatal(err)
	}
	return buf.Bytes()
}

func PGPDecode(in []byte) []byte {
	md, err := openpgp.ReadMessage(bytes.NewBuffer(in), PGPPrivate, nil, nil)
	if err != nil {
		logger.Fatal(err)
	}
	out, err := ioutil.ReadAll(md.UnverifiedBody)
	if err != nil {
		logger.Fatal(err)
	}
	return out
}
