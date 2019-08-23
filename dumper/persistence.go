package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type Persistent interface {
	PersistNonEndorserTx(NonEndorserTx) error
	PersistEndorserTx(EndorserTx) error
	PersistWrite(Write) error
	PersistBlock(Block) error
}

type FileDumper struct {
	NonEndorserTxPath   string
	NonEndorserTxSeqNum uint64
	EndorserTxPath      string
	WritePath           string
	WriteSeqNum         uint64
	BlockPath           string
}

func (fd FileDumper) PersistNonEndorserTx(tx NonEndorserTx) error {
	err := fd.persistToFile(tx, path.Join(fd.NonEndorserTxPath, fmt.Sprintf("%d.json", fd.NonEndorserTxSeqNum)))
	fd.NonEndorserTxSeqNum = fd.NonEndorserTxSeqNum + 1
	return err
}

func (fd FileDumper) PersistEndorserTx(tx EndorserTx) error {
	return fd.persistToFile(tx, path.Join(fd.EndorserTxPath, fmt.Sprintf("%s.json", tx.TxID)))
}

func (fd FileDumper) PersistWrite(w Write) error {
	err := fd.persistToFile(w, path.Join(fd.WritePath, fmt.Sprintf("%d.json", fd.WriteSeqNum)))
	fd.WriteSeqNum = fd.WriteSeqNum + 1
	return err
}

func (fd FileDumper) PersistBlock(b Block) error {
	return fd.persistToFile(b, path.Join(fd.BlockPath, fmt.Sprintf("%s-%d.json", b.ChannelID, b.BlockNumber)))
}

func (fd FileDumper) persistToFile(object interface{}, file string) error {
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

var DefaultConfig = FileDumper{
	NonEndorserTxPath:   "NonEndorserTx",
	NonEndorserTxSeqNum: 0,
	EndorserTxPath:      "EndorserTx",
	WritePath:           "Write",
	WriteSeqNum:         0,
	BlockPath:           "Block",
}
