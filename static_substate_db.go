package substate

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/urfave/cli/v2"
)

const substateNamespace = "substatedir"

var (
	SubstateDbFlag = cli.StringFlag{
		Name:  "substate-db",
		Usage: "Data directory for substate recorder/replayer",
	}
	substateDir      = SubstateDbFlag.Value
	staticSubstateDB *DB
	RecordReplay     bool = false
)

// Deprecated: Please use NewDb instead. This function
// relies on a global variable which will be removed in the future.
func OpenSubstateDB() {
	fmt.Println("record-replay: OpenSubstateDB")
	backend, err := rawdb.NewLevelDBDatabase(substateDir, 1024, 100, substateNamespace, false)
	if err != nil {
		panic(fmt.Errorf("error opening substate leveldb %s: %v", substateDir, err))
	}
	fmt.Println("record-replay: opened successfully")
	staticSubstateDB = newSubstateDB(backend)
}

// Deprecated: Please use NewDb instead. This function
// relies on a global variable which will be removed in the future.
func OpenSubstateDBReadOnly() {
	fmt.Println("record-replay: OpenSubstateDB")
	backend, err := rawdb.NewLevelDBDatabase(substateDir, 1024, 100, substateNamespace, true)
	if err != nil {
		panic(fmt.Errorf("error opening substate leveldb %s: %v", substateDir, err))
	}
	staticSubstateDB = newSubstateDB(backend)
}

// Deprecated: Please use MakeDb instead. This function
// relies on a global variable which will be removed in the future.
func SetSubstateDbBackend(backend ethdb.Database) {
	fmt.Println("record-replay: SetSubstateDB")
	staticSubstateDB = newSubstateDB(backend)
}

// Deprecated: Please use NewDb or MakeDb to create new DB and call the method Close() from
// returned object. This function relies on a global variable which will be removed in the future.
func CloseSubstateDB() {
	defer fmt.Println("record-replay: CloseSubstateDB")

	err := staticSubstateDB.Close()
	if err != nil {
		panic(fmt.Errorf("error closing substate leveldb %s: %v", substateDir, err))
	}
}

// Deprecated: Please use NewDb or MakeDb to create new DB and call the method Compact() from
// returned object. This function relies on a global variable which will be removed in the future.
func CompactSubstateDB() {
	fmt.Println("record-replay: CompactSubstateDB")

	// compact entire DB
	err := staticSubstateDB.Compact(nil, nil)
	if err != nil {
		panic(fmt.Errorf("error compacting substate leveldb %s: %v", substateDir, err))
	}
}

// Deprecated: Please use NewInMemoryDb instead. This function
// relies on a global variable which will be removed in the future.
func OpenFakeSubstateDB() {
	backend := rawdb.NewMemoryDatabase()
	staticSubstateDB = newSubstateDB(backend)
}

// Deprecated: Please use NewInMemoryDb to create new in-memory DB and call the method Close() from
// returned object. This function relies on a global variable which will be removed in the future.
func CloseFakeSubstateDB() {
	staticSubstateDB.Close()
}

// Deprecated: This function relies on a global variable which will be removed in the future.
func SetSubstateDbFlags(ctx *cli.Context) {
	substateDir = ctx.String(SubstateDbFlag.Name)
	fmt.Printf("record-replay: --substatedir=%s\n", substateDir)
}

// Deprecated: This function relies on a global variable which will be removed in the future.
func SetSubstateDb(dir string) {
	substateDir = dir
}

func HasCode(codeHash common.Hash) bool {
	return staticSubstateDB.HasCode(codeHash)
}

func GetCode(codeHash common.Hash) []byte {
	return staticSubstateDB.GetCode(codeHash)
}

func PutCode(code []byte) {
	staticSubstateDB.PutCode(code)
}

func HasSubstate(block uint64, tx int) bool {
	return staticSubstateDB.HasSubstate(block, tx)
}

func GetSubstate(block uint64, tx int) *Substate {
	return staticSubstateDB.GetSubstate(block, tx)
}

func GetBlockSubstates(block uint64) map[int]*Substate {
	return staticSubstateDB.GetBlockSubstates(block)
}

func PutSubstate(block uint64, tx int, substate *Substate) {
	staticSubstateDB.PutSubstate(block, tx, substate)
}

func DeleteSubstate(block uint64, tx int) {
	staticSubstateDB.DeleteSubstate(block, tx)
}

func GetFirstSubstate() *Substate {
	return staticSubstateDB.GetFirstSubstate()
}

func GetLastSubstate() (*Substate, error) {
	return staticSubstateDB.GetLastSubstate()
}
