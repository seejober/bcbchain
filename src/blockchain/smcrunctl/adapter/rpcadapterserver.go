package adapter

import (
	"common/socket"
	"strconv"

	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/log"
)

func start(port int, logger log.Logger) error {
	//call function getting IP address

	//server_addr = "http://<ip>:<port>"

	address := "tcp://0.0.0.0:" + strconv.Itoa(port)

	SetLogger(logger)

	// start server and wait forever
	svr, err := socket.NewServer(address, Routes, 120, logger)
	if err != nil {
		cmn.Exit(err.Error())
	}

	err = svr.Start()
	if err != nil {
		cmn.Exit(err.Error())
	}

	return nil
}
