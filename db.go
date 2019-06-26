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
	"database/sql"
	"fmt"

	"github.com/google/logger"
	_ "github.com/lib/pq"
)

type BlockInfo struct {
	ID   int64
	Hash []byte
	Data []byte
}

var (
	db *sql.DB
)

func DBOpen() string {
	var err error

	postInit := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable password=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Name, cfg.DB.Password)
	db, err = sql.Open("postgres", postInit)
	if err != nil {
		logger.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		logger.Fatal(err)
	}
	return `OK`
}

func DBClose() {
	db.Close()
}

func GetLastBlock() (id int64) {
	rows, err := db.Query(`select id from "block_chain" order by id desc limit 1`)
	if err != nil {
		logger.Error(err)
	}
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&id)
	}
	return
}

func LoadBlocks(offset int64) []BlockInfo {
	ret := make([]BlockInfo, 0, 25)
	rows, err := db.Query(`SELECT id, hash, data FROM "block_chain" WHERE id > $1 order by id limit 25`,
		offset)
	if err != nil {
		logger.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var block BlockInfo
		rows.Scan(&block.ID, &block.Hash, &block.Data)
		ret = append(ret, block)
	}
	return ret
}

func GetBlocks() []BlockInfo {
	ret := make([]BlockInfo, 0, 25)
	for {
		offset := LastID()
		rows, err := db.Query(`SELECT id, hash, data FROM "block_chain" WHERE id >= $1 order by id`,
			offset)
		if err != nil {
			logger.Fatal(err)
		}
		for rows.Next() {
			var (
				block BlockInfo
			)
			rows.Scan(&block.ID, &block.Hash, &block.Data)
			if block.ID == offset {
				if !bytes.Equal(block.Hash, GetHash(offset)) { // Rollback has been detected
					fmt.Println(`Rollback has been detected`)
					rows.Close()
					offset -= 10
					if offset < 0 {
						offset = 1
					}
					prev, err := db.Query(`SELECT id, hash FROM "block_chain" WHERE id >= $1 order by id     limit 10`,
						offset)
					if err != nil {
						logger.Fatal(err)
					}
					for prev.Next() {
						var (
							id   int64
							hash []byte
						)
						prev.Scan(&id, &hash)
						if bytes.Equal(hash, GetHash(id)) {
							offset = id
						} else {
							break
						}
					}
					store.Set(lastId, offset)
					prev.Close()
					break
				}
			}
			ret = append(ret, block)
		}
	}
	return ret
}
