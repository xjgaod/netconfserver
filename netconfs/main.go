// netconfs project main.go
package main

func main() {
	local := new(NetConfSData)
	local.Init()
	netConfInit(&local.config)
}
