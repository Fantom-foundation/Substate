module github.com/Fantom-foundation/Substate

go 1.15

require (
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/davecgh/go-spew v1.1.1
	github.com/ethereum/go-ethereum v1.10.25
	github.com/google/gofuzz v1.1.1-0.20200604201612-c04b05f3adfa
	github.com/jedisct1/go-minisign v0.0.0-20230811132847-661be99b8267
	github.com/stretchr/testify v1.8.0
	github.com/syndtr/goleveldb v1.0.1-0.20210305035536-64b5b1c73954
	github.com/urfave/cli/v2 v2.24.4
	golang.org/x/crypto v0.17.0
	golang.org/x/sys v0.15.0
)

replace github.com/ethereum/go-ethereum => github.com/Fantom-foundation/go-ethereum-substate v1.1.1-0.20230110052435-1ac0bdd8f402
