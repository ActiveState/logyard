package server

import (
	"github.com/ActiveState/doozer"
	"github.com/srid/doozerconfig"
	"github.com/srid/log2"
)

type clusterConfig struct {
	Endpoint string `doozer:"endpoint"`
}

var Config *clusterConfig

func Init(conn *doozer.Conn, rev int64) {
	Config = new(clusterConfig)
	cfg := doozerconfig.New(conn, rev, Config, "/cluster/config/")
	err := cfg.Load()
	if err != nil {
		log2.Fatal(err)
	}

	go cfg.Monitor("/cluster/config/*", func(change *doozerconfig.Change, err error) {
		if err != nil {
			log2.Errorf("Unable to process cluster config change in doozer: %s", err)
			return
		}
	})
}
