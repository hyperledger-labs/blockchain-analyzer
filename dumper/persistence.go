package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

// Implement this interface for custom persistence (e.g., writing the data into database instead of json), and use that instead of FileDumper. For example, see FileDumper.
type Persistent interface {
	PersistNonEndorserTx(NonEndorserTx) error
	PersistEndorserTx(EndorserTx) error
	PersistWrite(Write) error
	PersistBlock(Block) error
}

// This implementation of the Persistent interface writes data to separate json files.
type FileDumper struct {
	NonEndorserTxPath   string
	NonEndorserTxSeqNum uint64
	EndorserTxPath      string
	WritePath           string
	WriteSeqNum         uint64
	BlockPath           string
}

// Writes non-endorser transaction data to a json file. Uses an increasing sequence number for file naming.
func (fd *FileDumper) PersistNonEndorserTx(tx NonEndorserTx) error {
	err := fd.persistToFile(tx, path.Join(fd.NonEndorserTxPath, fmt.Sprintf("%d.json", fd.NonEndorserTxSeqNum)))
	fd.NonEndorserTxSeqNum = fd.NonEndorserTxSeqNum + 1
	return err
}

// Writes endorser transaction data to a json file. Uses the transaction ID for file naming.
func (fd *FileDumper) PersistEndorserTx(tx EndorserTx) error {
	return fd.persistToFile(tx, path.Join(fd.EndorserTxPath, fmt.Sprintf("%s.json", tx.TxID)))
}

// Writes write data to a json file. Uses an increasing sequence number for file naming.
func (fd *FileDumper) PersistWrite(w Write) error {
	err := fd.persistToFile(w, path.Join(fd.WritePath, fmt.Sprintf("%d.json", fd.WriteSeqNum)))
	fd.WriteSeqNum = fd.WriteSeqNum + 1
	return err
}

// Writes block data to a json file. Uses channel ID and block number for file naming.
func (fd *FileDumper) PersistBlock(b Block) error {
	return fd.persistToFile(b, path.Join(fd.BlockPath, fmt.Sprintf("%s-%d.json", b.ChannelID, b.BlockNumber)))
}

// Persists an object to a given json file. If the file does not exists, it creates.
// If the file already exists, it appends the new json to the end. NOTE: This breaks the json syntax of the file (missing "[", "]" and "," characters).
func (fd *FileDumper) persistToFile(object interface{}, file string) error {
	objectJSONBytes, err := json.Marshal(object)
	if err != nil {
		fmt.Println(err)
		return err
	}
	objectJSONString := string(objectJSONBytes)

	f, err := os.OpenFile(file,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer f.Close()
	_, err = f.WriteString(objectJSONString)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// Default implementation of FileDumper
var DefaultConfig = &FileDumper{
	NonEndorserTxPath:   "NonEndorserTx",
	NonEndorserTxSeqNum: 0,
	EndorserTxPath:      "EndorserTx",
	WritePath:           "Write",
	WriteSeqNum:         0,
	BlockPath:           "Block",
}
