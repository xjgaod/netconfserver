/*****************************************************************************/
/**
* \file       netconfs.go
* \author     gaoxj
* \date       2018/06/05
* \brief      实现netconf server,和onos互通
* \note       Copyright (c) 2000-2020  赛特斯信息科技股份有限公司
* \remarks    修改日志
******************************************************************************/
package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/robfig/cron"
	"golang.org/x/crypto/ssh"
)

var completeNetconf chan int = make(chan int)

type NetConfServer struct {
	sync.RWMutex
	channels   map[ssh.Channel]int32
	listenAddr string
	username   string
	password   string
}

var gNetConfServer *NetConfServer

func netConfInit(config *ConfigData) {
	gNetConfServer = NewNetConfServer(config.netconfAddr)
	gNetConfServer.SetAuth(config.username, config.password)
	gNetConfServer.Serve()
	<-completeNetconf
	return
}

func netConfWrite(handle int32, data []byte) error {
	connection, err := gNetConfServer.getConnection(handle)
	if err != nil {
		return err
	}

	err = gNetConfServer.sendNotification(connection, data)
	return err
}

func NewNetConfServer(addr string) *NetConfServer {
	ncs := &NetConfServer{}
	ncs.channels = make(map[ssh.Channel]int32)
	ncs.listenAddr = addr

	return ncs
}

func (ncs *NetConfServer) SetAuth(u, p string) {
	ncs.username = u
	ncs.password = p
}

func (ncs *NetConfServer) Serve() {
	go func() {
		//defer log.DeferHdl()

		config := ncs.initConfig()
		listener, err := net.Listen("tcp", ncs.listenAddr)
		if err != nil {
			log.Fatalf("Failed to listen on %s (%s)", ncs.listenAddr, err)
		}

		// Accept all connections
		log.Printf("Listening on ", ncs.listenAddr)
		for {
			tcpConn, err := listener.Accept()
			if err != nil {
				log.Printf("Failed to accept incoming connection (%s)", err)
				continue
			}
			// Before use, a handshake must be performed on the incoming net.Conn.
			sshConn, chans, reqs, err := ssh.NewServerConn(tcpConn, config)
			if err != nil {
				log.Printf("Failed to handshake (%s)", err)
				continue
			}

			log.Printf("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())
			// Discard all global out-of-band Requests
			go ssh.DiscardRequests(reqs)
			// Accept all channels
			go ncs.handleClient(chans, sshConn)
		}
	}()
	completeNetconf <- 0
}

func (ncs *NetConfServer) initConfig() *ssh.ServerConfig {
	pKey := `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAzODqzuKMlPQv7Pl9d8gpUUoms9+aPQfCW2nGko1hcHwScr3q
l0uecpRUOjC646HhfRrZya6a8a1wqXFQmDvDXRRD6UIdCjAQw1RamrNFvfHqw86T
M9YD3wUEb0MYuk3U0GSw0KfbwUH9uXHlICbNWyTjXHqtc/UL52lkAtJZIrDeEUzR
yz6i82k1JcLQeYiT+59uHkrCKLUgc8ots1irteccAYL9qkCqBc0q8yJi2gdU5j4j
xcxUhAKVvblylcMnRWsrmlD83PKNGgTrqsNb4ixxKsJ5TknIYVnPIot3loifxJ2D
hd+5FUL3TxV9/tosoDNp0qyiIyWST4buL7MFnwIDAQABAoIBADcRMSugSubyZztL
p8SdQTW/N33bWOqUflHxsVTMuWbxgkfi8f3ubk0fvy/QzzrF1QS6RdVmn/DbjE+O
zHnOfmkzPDmi8ok9eBP4RU4VZ2Zxcodkq74gBPRZteJt52ST7OKCXeAHbhKYuSiy
C0LECyg3VwERXDOxppxxgFcd0KV452EPxarjBdSad56889YqzL4/FlvGKhyZkQEg
mSwNNi1IPEhnDRVwYsVBxuB3iGsNwntbSeTGxGRF4dXDLxTHOb0wTStSc/DA3yBS
5hsyLMB6MyU0ue8uqCcDuzLirbvPKj0n+Q++uNnAg55pHU6bL1zET/kYeA5sDMea
0V/PqcECgYEA6rP5ZQ7jceqqoe15UVPj/R25sgZ2md6iQkCGq66LK7PNt15u7KSV
aVS7bOP7XSP2bZqTAHJByzkj4ejZDX9j3FuNrUV8PwZ9jxv5rmy0KkdsfeFyz5I/
ESXu7HGh4C0244Pvg4m3Aa0lKFWZVYrnBHk5pSsrahf8gfZAHPx74uECgYEA33gj
HuqIQz/g6dk3zTPRZWVCx/uvY++9ptze0YUNLjzsTaamDwJr14VjtPysyU2g+f95
l1QOA7fsDomfwQnl/fvmcM56cZvZDdhYQhwP+rQgECDs94gIsNeimil+v/dSx/J1
qqailQO7MjyLYNSsvp0iXOHc9sb0hh1+UbxkeH8CgYEA2X7Gkjvl0d8hGMW0MwWG
tT0ipDMRHS4PN04MfnRVS75n2JGOQYWTX/TBavsqKPn2l0MzDqrTBbyB4AujeLqg
k8fT1soZhV5CZKgMDPN3Uea2R0Dw4CIqh32bl0kGNXQw9U2CW2b3THpjgKkyWu9J
ff/Ix6LlrH9l5BmK+FGRjIECgYAuhSjyh6pkLYkZxWFrc20U6ZaUYR2q9T6K3RH5
lfQfewlKRPXuy/c9P4R5Kdyib2migX+DdDkSpxgaEqZSHkhlrinTs/gjbGksC6yb
3pGpBBRkpyYNhaEhh1JPO3IqbkcqXpwGMXhJAyTWGWp+duebKsT7hv1j1hkTTlJ8
m3Zi6wKBgQCBAUR3ENdZGbKynVXmw0WRjeatxF9MaN70DJb9YduvZUkQOy57r3VD
fNPiJUyzI2RWHeq8da3WlmvQ+T1QYnljTtCXfciFEjOvUtiJNKauGqsQPpE9zszx
ckDayAMUvcIZMjT8EWF5bMr+0JXR3gYhLpWdQA7XJfI+0c55r0eNng==
-----END RSA PRIVATE KEY-----`
	config := &ssh.ServerConfig{
		//Define a function to run when a client attempts a password login
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			// Should use constant-time compare (or better, salt+hash) in a production setting.
			if c.User() == ncs.username && string(pass) == ncs.password {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
		// You may also explicitly allow anonymous client authentication, though anon bash
		// sessions may not be a wise idea
		// NoClientAuth: true,
	}

	private, err := ssh.ParsePrivateKey([]byte(pKey))
	if err != nil {
		log.Printf("Failed to parse private key")
	}

	config.AddHostKey(private)

	return config
}

func (ncs *NetConfServer) handleClient(chans <-chan ssh.NewChannel, conn *ssh.ServerConn) {
	// Service the incoming Channel channel in go routine
	for newChannel := range chans {
		go ncs.handleSession(newChannel, conn.RemoteAddr().String())
	}
	log.Printf("handleClient exit for ssh connection %s (%s)", conn.RemoteAddr(), conn.ClientVersion())
}

func (ncs *NetConfServer) readOneMsg(conn ssh.Channel) ([]byte, error) {
	var buf []byte
	var s byte = 0 //当前的状态
	var end bool   //是否要结束

	for !end {
		c := make([]byte, 1)
		_, err := conn.Read(c)
		if err != nil {
			return nil, err
		}
		switch s {
		case 0:
			if c[0] == ']' {
				s = 1
			}
		case 1:
			if c[0] == ']' {
				s = 2
			} else {
				s = 0
			}
		case 2:
			if c[0] == '>' {
				s = 3
			} else {
				s = 0
			}
		case 3:
			if c[0] == ']' {
				s = 4
			} else {
				s = 0
			}
		case 4:
			if c[0] == ']' {
				s = 5
			} else {
				s = 0
			}
		case 5:
			if c[0] == '>' {
				end = true
			} else {
				s = 0
			}
		}
		buf = append(buf, c[0])
	}

	return buf, nil
}

func (ncs *NetConfServer) handleSession(newChannel ssh.NewChannel, raddr string) {
	//defer log.DeferHdl()

	if t := newChannel.ChannelType(); t != "session" {
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}

	// At this point, we have the opportunity to reject the client's
	// request for another logical connection
	connection, requests, err := newChannel.Accept()
	if err != nil {
		log.Printf("Could not accept channel (%s)", err)
		return
	}

	// Sessions have out-of-band requests such as "shell", "pty-req" and "env"
	go func() {
		for req := range requests {
			switch req.Type {
			case "shell":
				// We only accept the default shell
				// (i.e. no command in the Payload)
				if len(req.Payload) == 0 {
					req.Reply(true, nil)
				}
			case "subsystem":
				log.Printf("Receive request %v\n", req)
				req.Reply(true, nil)
			default:
				log.Printf("Receive request %v\n", req)
			}
		}
	}()

	var sess *Session
	defer func() {
		if sess != nil {
			ncs.UserGone(connection)
			DelSession(sess)
			sess = nil
		}
		connection.Close()
	}()

	for true {
		buf, err := ncs.readOneMsg(connection)
		if err != nil {
			log.Printf("Netconf Read from connection fail (%s)", err)
			return
		}
		//log.Printf("Read from connection %d %s", len(buf), buf)
		if strings.Contains(string(buf), "<hello") {
			ncs.replyHello(connection)
			sess = ncs.UserNew(raddr, connection)
			ncs.notifySysmgr(connection)
		} else if strings.Contains(string(buf), "<close-session/>") {
			//exit
			log.Printf("Netconf session closed by client ", raddr)
			return
		} else if strings.Contains(string(buf), "</rpc>") && !strings.Contains(string(buf), "<rpc message-id=") {
			ncs.replyOk(connection, string(buf))
		} else if strings.Contains(string(buf), "/interfaces-state/interface[type='if_eth']") {
			ncs.replyInterface(connection, string(buf))
		} else {
			if sess != nil {
				sess.MsgUp(string(buf))
			} else {
				log.Printf("Receive msg, but sessin not setup for %s", raddr)
			}
		}
	}
}

func (ncs *NetConfServer) UserNew(raddr string, conn ssh.Channel) *Session {
	err, handle := ncs.allocIndex(conn)
	if err != nil {
		log.Printf("Could not alloc index for ", conn)
		return nil
	}

	return NewSession(raddr, handle)
}

func (ncs *NetConfServer) UserGone(conn ssh.Channel) {
	ncs.delIndex(conn)
}

func (ncs *NetConfServer) allocIndex(c ssh.Channel) (error, int32) {
	ncs.Lock()
	defer ncs.Unlock()

	if ii, ok := ncs.channels[c]; ok {
		return nil, ii
	}

	slot := make(map[int32]bool)
	for _, value := range ncs.channels {
		slot[value] = true
	}

	var ii int32
	for ii = 1; ; ii++ {
		if _, ok := slot[ii]; !ok {
			ncs.channels[c] = ii
			return nil, ii
		}
	}

	return fmt.Errorf("alloc index failed"), 0
}

func (ncs *NetConfServer) getIndex(c ssh.Channel) (error, int32) {
	ncs.RLock()
	defer ncs.RUnlock()

	if ii, ok := ncs.channels[c]; ok {
		return nil, ii
	} else {
		return fmt.Errorf("get index failed"), 0
	}
}

func (ncs *NetConfServer) delIndex(c ssh.Channel) error {
	ncs.Lock()
	defer ncs.Unlock()

	if _, ok := ncs.channels[c]; ok {
		delete(ncs.channels, c)
		return nil
	} else {
		return fmt.Errorf("del index failed")
	}

}

func (ncs *NetConfServer) getConnection(handle int32) (c ssh.Channel, err error) {
	ncs.RLock()
	defer ncs.RUnlock()

	var ii int32
	for c, ii = range ncs.channels {
		if ii == handle {
			return
		}
	}

	err = fmt.Errorf("Not find the connection %d", handle)
	return
}

func (ncs *NetConfServer) replyHello(user ssh.Channel) {
	user.Write([]byte(ncs.buildHello()))
}

func (ncs *NetConfServer) buildHello() []byte {
	var buildStr string
	xmlHeader := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>"
	xmlFooter := "]]>]]>"

	buildStr += xmlHeader
	buildStr += "\n"
	buildStr += "<hello xmlns=\"urn:ietf:params:xml:ns:netconf:base:1.0\">\n"
	buildStr += "  <capabilities>\n"

	buildStr += "    <capability>"
	buildStr += "urn:ietf:params:netconf:base:1.0"
	buildStr += "</capability>\n"

	buildStr += "    <capability>"
	buildStr += "urn:ietf:params:netconf:capability:notification:1.0"
	buildStr += "</capability>\n"

	buildStr += "    <capability>"
	buildStr += "urn:ietf:params:netconf:capability:interleave:1.0"
	buildStr += "</capability>\n"

	buildStr += "  </capabilities>\n"
	buildStr += "  </hello>\n"
	buildStr += xmlFooter

	return []byte(buildStr)
}

func (ncs *NetConfServer) sendNotification(user ssh.Channel, data []byte) error {
	msg := []byte(ncs.buildNotification(data))
	n, err := user.Write(msg)
	if err != nil {
		return err
	}
	if n != len(msg) {
		return fmt.Errorf("Expect send %d bytes, only send %d bytes", len(msg), n)
	}

	return nil
}

func (ncs *NetConfServer) buildNotification(data []byte) []byte {
	var buildStr string
	xmlHeader := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>"
	xmlFooter := "]]>]]>"

	buildStr += xmlHeader
	buildStr += "\n"
	buildStr += "<notification xmlns=\"urn:ietf:params:xml:ns:netconf:notification:1.0\">"
	// += "<notification>\n"
	buildStr += string(data)

	buildStr += "\n"
	buildStr += "  </notification>\n"
	buildStr += xmlFooter

	return []byte(buildStr)
}

func (ncs *NetConfServer) replyOk(user ssh.Channel, message string) {
	user.Write([]byte(ncs.buildOk(message)))
}

func (ncs *NetConfServer) replyInterface(user ssh.Channel, message string) {
	user.Write([]byte(ncs.buildInterface(message)))
}

func (ncs *NetConfServer) notifySysmgr(user ssh.Channel) {
	c := cron.New()
	spec := "*/5 * * * * ?"
	c.AddFunc(spec, func() {
		user.Write([]byte(ncs.buildSysmgr()))
	})
	c.Start()
}

func (ncs *NetConfServer) buildOk(message string) []byte {
	xmlHeader := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>"
	xmlFooter := "]]>]]>"
	spilString1 := "message-id="
	s := strings.Split(message, spilString1)
	spilString2 := ">"
	ss := strings.Split(s[1], spilString2)
	var buildStr string
	buildStr += xmlHeader
	buildStr += "<rpc-reply xmlns=\"urn:ietf:params:xml:ns:netconf:base:1.0\" message-id="
	buildStr += ss[0]
	buildStr += ">"
	buildStr += "<ok/>"
	buildStr += "</rpc-reply>"
	buildStr += xmlFooter
	//fmt.Printf(ss[0])
	return []byte(buildStr)
}

func (ncs *NetConfServer) buildSysmgr() []byte {
	data := `<?xml version="1.0" encoding="UTF-8"?>
<notification xmlns="urn:ietf:params:xml:ns:netconf:notification:1.0"><eventTime>2018-06-09T05:51:02.387635+00:00</eventTime>
<sysmgr-system-info-notify-msg xmlns='http://certusnet.com/nfv/flexbng/certus-flexbng-sysmgr'>
  <system_info>
    <cid>1</cid>
    <seq_id>1</seq_id>
    <container_role>5</container_role>
    <cpu_num>2</cpu_num>
    <cpu_freq>2194.916</cpu_freq>
    <memory_info>
      <memory_id>0</memory_id>
      <memory_freq>2666.67</memory_freq>
      <memory_type>DDRIII</memory_type>
      <memory_size>8G</memory_size>
    </memory_info>
    <memory_info>
      <memory_id>1</memory_id>
      <memory_freq>2666.67</memory_freq>
      <memory_type>DDRIII</memory_type>
      <memory_size>8G</memory_size>
    </memory_info>
    <System_version_str>flexedge-atm-v2.3-B2P2-debug</System_version_str>
    <Device_type_str>Red Hat: KVM</Device_type_str>
    <System_online_time>16145</System_online_time>
    <alert_flag>2</alert_flag>
    <global_cpu_5s_busy>112</global_cpu_5s_busy>
    <global_cpu_5s_total>544</global_cpu_5s_total>
    <global_cpu_1m_busy>1036</global_cpu_1m_busy>
    <global_cpu_1m_total>5839</global_cpu_1m_total>
    <global_cpu_5m_busy>5805</global_cpu_5m_busy>
    <global_cpu_5m_total>31545</global_cpu_5m_total>
    <global_memory_total>3899132</global_memory_total>
    <global_memory_used>3774524</global_memory_used>
    <global_memory_free>124608</global_memory_free>
    <global_disk_total>39265556</global_disk_total>
    <global_disk_used>14318396</global_disk_used>
    <global_disk_free>24947160</global_disk_free>
  </system_info>
</sysmgr-system-info-notify-msg>
</notification>
]]>]]>`
	return []byte(data)
}

func (ncs *NetConfServer) buildInterface(message string) []byte {
	xmlHeader := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>"
	xmlFooter := "]]>]]>"
	spilString1 := "message-id="
	s := strings.Split(message, spilString1)
	spilString2 := "xmlns="
	ss := strings.Split(s[1], spilString2)
	var buildStr string
	buildStr += xmlHeader
	buildStr += "<rpc-reply xmlns=\"urn:ietf:params:xml:ns:netconf:base:1.0\" message-id="
	buildStr += ss[0]
	buildStr += ">"
	buildStr += "<data><interfaces-state xmlns=\"urn:ietf:params:xml:ns:yang:ietf-interfaces\">"
	buildStr += "<interface><name>10gei-1/1/0</name><type>if_eth</type><ipv4-address>172.13.14.4/24</ipv4-address><description>WAN</description><admin-status>up</admin-status><phys-status>up</phys-status><oper-status>up</oper-status><detect-status>1</detect-status><alias-name>WAN1</alias-name><if-index>2</if-index><phys-address>aa:bb:cc:dd:e7:f3</phys-address><speed>10000000</speed><statistics><in-pkts>8</in-pkts><in-octets>648</in-octets><in-unicast-pkts>0</in-unicast-pkts><in-broadcast-pkts>0</in-broadcast-pkts><in-multicast-pkts>8</in-multicast-pkts><in-discards>0</in-discards><in-errors>0</in-errors><in-rate-bits>0</in-rate-bits><in-rate-packets>0</in-rate-packets><in-peak-rate>0</in-peak-rate><in-peak-time>-</in-peak-time><in-bandwidth-timestamp>0.00%</in-bandwidth-timestamp><out-pkts>0</out-pkts><out-octets>0</out-octets><out-unicast-pkts>0</out-unicast-pkts><out-broadcast-pkts>0</out-broadcast-pkts><out-multicast-pkts>0</out-multicast-pkts><out-discards>0</out-discards><out-errors>0</out-errors><out-rate-bits>0</out-rate-bits><out-rate-packets>0</out-rate-packets><out-peak-rate>0</out-peak-rate><out-peak-time>-</out-peak-time><out-bandwidth-timestamp>0.00%</out-bandwidth-timestamp><cur-time>2018-06-11 10:09:48</cur-time></statistics><parent-name>10gei-1/1/0</parent-name><is-physical>true</is-physical><switch-port>false</switch-port><is-virtualize>false</is-virtualize><mtu>1500</mtu><vrf-id>0</vrf-id></interface>"
	buildStr += "<interface><name>10gei-1/1/1</name><type>if_eth</type><ipv4-address>172.13.14.2/24</ipv4-address><description>WAN</description><admin-status>up</admin-status><phys-status>up</phys-status><oper-status>up</oper-status><detect-status>1</detect-status><alias-name>WAN2</alias-name><if-index>2</if-index><phys-address>aa:bb:cc:dd:e7:f4</phys-address><speed>10000000</speed><statistics><in-pkts>8</in-pkts><in-octets>648</in-octets><in-unicast-pkts>0</in-unicast-pkts><in-broadcast-pkts>0</in-broadcast-pkts><in-multicast-pkts>8</in-multicast-pkts><in-discards>0</in-discards><in-errors>0</in-errors><in-rate-bits>0</in-rate-bits><in-rate-packets>0</in-rate-packets><in-peak-rate>0</in-peak-rate><in-peak-time>-</in-peak-time><in-bandwidth-timestamp>0.00%</in-bandwidth-timestamp><out-pkts>0</out-pkts><out-octets>0</out-octets><out-unicast-pkts>0</out-unicast-pkts><out-broadcast-pkts>0</out-broadcast-pkts><out-multicast-pkts>0</out-multicast-pkts><out-discards>0</out-discards><out-errors>0</out-errors><out-rate-bits>0</out-rate-bits><out-rate-packets>0</out-rate-packets><out-peak-rate>0</out-peak-rate><out-peak-time>-</out-peak-time><out-bandwidth-timestamp>0.00%</out-bandwidth-timestamp><cur-time>2018-06-11 10:09:48</cur-time></statistics><parent-name>10gei-1/1/1</parent-name><is-physical>true</is-physical><switch-port>false</switch-port><is-virtualize>false</is-virtualize><mtu>1500</mtu><vrf-id>0</vrf-id></interface>"
	buildStr += "<interface><name>10gei-1/1/2</name><type>if_eth</type><ipv4-address>172.19.14.3/24</ipv4-address><description>LAN</description><admin-status>up</admin-status><phys-status>up</phys-status><oper-status>up</oper-status><detect-status>1</detect-status><alias-name>LAN1</alias-name><if-index>3</if-index><phys-address>aa:bb:cc:dd:e7:f5</phys-address><speed>10000000</speed><statistics><in-pkts>8</in-pkts><in-octets>648</in-octets><in-unicast-pkts>0</in-unicast-pkts><in-broadcast-pkts>0</in-broadcast-pkts><in-multicast-pkts>8</in-multicast-pkts><in-discards>0</in-discards><in-errors>0</in-errors><in-rate-bits>0</in-rate-bits><in-rate-packets>0</in-rate-packets><in-peak-rate>0</in-peak-rate><in-peak-time>-</in-peak-time><in-bandwidth-timestamp>0.00%</in-bandwidth-timestamp><out-pkts>0</out-pkts><out-octets>0</out-octets><out-unicast-pkts>0</out-unicast-pkts><out-broadcast-pkts>0</out-broadcast-pkts><out-multicast-pkts>0</out-multicast-pkts><out-discards>0</out-discards><out-errors>0</out-errors><out-rate-bits>0</out-rate-bits><out-rate-packets>0</out-rate-packets><out-peak-rate>0</out-peak-rate><out-peak-time>-</out-peak-time><out-bandwidth-timestamp>0.00%</out-bandwidth-timestamp><cur-time>2018-06-11 10:09:48</cur-time></statistics><parent-name>10gei-1/1/2</parent-name><is-physical>true</is-physical><switch-port>false</switch-port><is-virtualize>false</is-virtualize><mtu>1500</mtu><vrf-id>0</vrf-id></interface>"
	buildStr += "</interfaces-state></data>"
	buildStr += "</rpc-reply>"
	buildStr += xmlFooter
	//fmt.Printf(buildStr)
	return []byte(buildStr)
}
