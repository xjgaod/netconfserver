/*****************************************************************************/
/**
* \file       fwdData.go
* \author     gaoxj
* \date       2018/05/31
* \brief      本地数据
* \note       Copyright (c) 2000-2020  赛特斯信息科技股份有限公司
* \remarks    修改日志
******************************************************************************/

package main

type SocketData struct {
	config ConfigData
}

//实现LocalData接口
func (std *SocketData) Init() {
	/*加载配置*/
	std.config.Init()
	std.config.Load()
}

func (std *SocketData) Uninit() {
}
