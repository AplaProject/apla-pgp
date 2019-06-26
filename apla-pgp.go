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
	"time"

	"github.com/BurntSushi/toml"
	"github.com/google/logger"
)

type DBConfig struct {
	Name     string
	Host     string // ipaddr, hostname, or "0.0.0.0"
	Port     int    // must be in range 1..65535
	User     string
	Password string
}

type PGPConfig struct {
	Path   string
	Phrase string // phrase for private key
}

type SettingsConfig struct {
	Timeout        int
	NodePrivateKey string
}

type Config struct {
	LogFile   string
	StoreFile string
	OutPath   string
	Settings  SettingsConfig
	PGP       PGPConfig
	DB        DBConfig
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
	logger.Info(`Read PGP keys: `, ReadPGPKeys())
	InitNodeKey()
	logger.Info(`Connect DB: `, DBOpen())
	defer DBClose()
	StoreOpen()
	defer StoreClose()

	logger.Info(`Previous state: `, LastID(), GetLastBlock())
	for LastID() < GetLastBlock() {
		blocks := LoadBlocks(LastID())
		for _, block := range blocks {
			ProcessBlock(block)
		}
	}
	chBlock := make(chan int)
	for {
		time.AfterFunc(time.Duration(cfg.Settings.Timeout)*time.Second, func() {
			blocks := GetBlocks()
			for _, block := range blocks {
				ProcessBlock(block)
			}
			chBlock <- 1
		})
		<-chBlock
	}
	logger.Info(`Finish`)
}
