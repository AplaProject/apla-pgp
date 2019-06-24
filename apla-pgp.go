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
	"io/ioutil"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/google/logger"
)

type Config struct {
	PGPPath string
	LogFile string
}

const (
	confFile = `apla-pgp.conf`
)

var (
	cfg Config
)

func main() {
	var (
		confData []byte
		err      error
	)
	if confData, err = ioutil.ReadFile(confFile); err != nil {
		logger.Fatalf("Failed to read config file: %v", err)
	}
	if _, err := toml.Decode(string(confData), &cfg); err != nil {
		logger.Fatalf("Failed to parse config: %v", err)
	}

	lf, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}
	defer lf.Close()
	defer logger.Init("apla-pgp", false, false, lf).Close()
	logger.SetFlags(log.LstdFlags)
	logger.Info(`Start`)
	/*	encStr, err := encTest(mySecretString)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Decrypted Secret:", len(encStr))
		decStr, err := decTest(encStr)
		if err != nil {
			log.Fatal(err)
		}
		// should be done
		log.Println("Decrypted Secret:", len(decStr))*/
	logger.Info(`Finish`)
}
