/*****************************************************************************/
/**
* \file       config.go
* \author     gaoxj
* \date       2016/04/21
* \brief      配置数据
* \note       Copyright (c) 2000-2020  赛特斯信息科技股份有限公司
* \remarks    修改日志
******************************************************************************/

package main

import (
	"fmt"

	toml "github.com/pelletier/go-toml"
)

type ConfigData struct {
	netconfAddr string /*netconf server监听的地址*/
	onosAddr    string /*onos监听的地址*/
}

func (c *ConfigData) Init() {
	c.netconfAddr = "172.19.8.182:2022"
	c.onosAddr = "172.19.106.227:4334"
}

func (c *ConfigData) Load() {
	config, err := toml.LoadFile("socket.conf")
	if err != nil {
		fmt.Println("Read config file error: ", err.Error(), ".Use default config instead!")
		return
	}
	if v := config.Get("netconf.ip"); v != nil {
		c.netconfAddr = v.(string)
	}

	if v := config.Get("onos.ip"); v != nil {
		c.onosAddr = v.(string)
	}
}
