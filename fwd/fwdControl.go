/*****************************************************************************/
/**
* \file       fwdControl.go
* \author     gaoxj
* \date       2018/05/31
* \brief      本地任务控制
* \note       Copyright (c) 2000-2020  赛特斯信息科技股份有限公司
* \remarks    修改日志
******************************************************************************/

package main

func StartMainWork() {
	local := new(SocketData)
	local.Init()
	socketConfInit(&local.config)
}
