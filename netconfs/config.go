/*****************************************************************************/
/**
* \file       config.go
* \author     gaoxj
* \date       2018/06/05
* \brief      配置数据
* \note       Copyright (c) 2000-2020  赛特斯信息科技股份有限公司
* \remarks    修改日志
******************************************************************************/

package main

import (
	"fmt"

	"github.com/pelletier/go-toml"
)

type ConfigData struct {
	netconfAddr string /*netconf server监听的地址*/
	username    string /*netconf server认证用户名*/
	password    string /*netconf server认证密码*/
}

func (c *ConfigData) Init() {
	c.netconfAddr = "172.19.8.182:2022"
	c.username = "admin"
	c.password = "cErtusnEt@2018"
}

func (c *ConfigData) Load() {
	config, err := toml.LoadFile("netconfs.conf")
	if err != nil {
		fmt.Println("Read config file error: ", err.Error(), ".Use default config instead!")
		return
	}

	if v := config.Get("netconf.local"); v != nil {
		c.netconfAddr = v.(string)
	}

	if v := config.Get("netconf.username"); v != nil {
		c.username = v.(string)
	}

	if v := config.Get("netconf.password"); v != nil {
		c.password = v.(string)
	}

	return
}
