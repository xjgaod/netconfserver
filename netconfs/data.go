/*****************************************************************************/
/**
* \file       data.go
* \author     gaoxj
* \date       2018/06/05
* \brief      加载本地数据
* \note       Copyright (c) 2000-2020  赛特斯信息科技股份有限公司
* \remarks    修改日志
******************************************************************************/

package main

type NetConfSData struct {
	config  ConfigData
	sessMgr SessionMgr
}

//实现LocalData接口
func (ncs *NetConfSData) Init() {
	/*加载配置*/
	ncs.config.Init()
	ncs.config.Load()

	ncs.sessMgr.Init()
}

func (ncd *NetConfSData) Uninit() {
}
