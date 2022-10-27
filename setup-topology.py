#!/usr/bin/env python
from mininet.net import Mininet
from mininet.cli import CLI
from mininet.link import Link, TCLink, Intf
from subprocess import Popen, PIPE
from mininet.log import setLogLevel
from mininet.node import Controller
from mininet.topo import Topo


if '__main__' == __name__:

    setLogLevel('info')
    net = Mininet(link=TCLink)
    key = "net.mptcp.mptcp_enabled"
    value = 1
    p = Popen("sysctl -w %s=%s" % (key, value),
            shell=True, stdout=PIPE, stderr=PIPE)
    stdout, stderr = p.communicate()
    print ("stdout=", stdout, "stderr=", stderr)
    
    
    c0 = net.addController("controller")

    sw1 = net.addSwitch('sw1')
    sw2 = net.addSwitch('sw2')
    sw3 = net.addSwitch('sw3')
    sw4 = net.addSwitch('sw4')
    
    MS = net.addHost('MS') #Main Server
    R1 = net.addHost('R1') #relay server 1
    R2 = net.addHost('R2') #relay server 2
    C = net.addHost('C')  # client
    router = net.addHost('router') #router
    linkopt = {'bw': 10}
    linkopt2 = {'bw': 100}
    
    net.addLink(router, sw1, cls=TCLink, **linkopt)
    net.addLink(router, sw2, cls=TCLink, **linkopt)
    net.addLink(router, sw3, cls=TCLink, **linkopt)
    net.addLink(router, sw4, cls=TCLink, **linkopt2)

    net.addLink(sw1, MS, cls=TCLink, **linkopt)
    net.addLink(sw2, MS, cls=TCLink, **linkopt)
    net.addLink(sw3, R1, cls=TCLink, **linkopt)
    net.addLink(sw3, R2, cls=TCLink, **linkopt)
    net.addLink(sw4, C, cls=TCLink, **linkopt2)
    net.build()
    router.cmd("ifconfig router-eth0 0")
    router.cmd("ifconfig router-etMS 0")
    router.cmd("ifconfig router-etR1 0")

    MS.cmd("ifconfig Sw1-eth0 0")
    MS.cmd("ifconfig Sw1-etMS 0")
    
    R1.cmd("ifconfig R1-eth0 0")
    
    R2.cmd("ifconfig R2-eth0 0")

    C.cmd("ifconfig C-eth0 0")
    
    router.cmd("echo 1 > /proc/sys/net/ipv4/ip_forward")
    router.cmd("ifconfig router-eth0 10.0.0.1 netmask 255.255.255.0")
    router.cmd("ifconfig router-etMS 10.0.1.1 netmask 255.255.255.0")
    router.cmd("ifconfig router-etR1 10.0.2.1 netmask 255.255.255.0")
    router.cmd("ifconfig router-etC 10.0.3.1 netmask 255.255.255.0")
    
    MS.cmd("ifconfig Sw1-eth0 10.0.0.2 netmask 255.255.255.0")
    MS.cmd("ifconfig Sw1-etMS 10.0.1.2 netmask 255.255.255.0")
    
    R1.cmd("ifconfig R1-eth0 10.0.2.2 netmask 255.255.255.0")

    R2.cmd("ifconfig R2-eth0 10.0.2.3 netmask 255.255.255.0")
    
    C.cmd("ifconfig C-eth0 10.0.3.2 netmask 255.255.255.0")
    
    MS.cmd("ip rule add from 10.0.0.2 table 1")
    MS.cmd("ip rule add from 10.0.1.2 table 2")
    MS.cmd("ip route add 10.0.0.0/24 dev Sw1-eth0 scope link table 1")
    MS.cmd("ip route add default via 10.0.0.1 dev Sw1-eth0 table 1")
    MS.cmd("ip route add 10.0.1.0/24 dev Sw1-etMS scope link table 2")
    MS.cmd("ip route add default via 10.0.1.1 dev Sw1-etMS table 2")
    MS.cmd("ip route add default scope global nexthop via 10.0.0.1 dev Sw1-eth0")
    
    R1.cmd("ip rule add from 10.0.2.2 table 1")
    R1.cmd("ip route add 10.0.2.0/24 dev R1-eth0 scope link table 1")
    R1.cmd("ip route add default via 10.0.2.1 dev R1-eth0 table 1")
    R1.cmd("ip route add default scope global nexthop via 10.0.2.1 dev R1-eth0")

    R2.cmd("ip rule add from 10.0.2.3 table 1")
    R2.cmd("ip route add 10.0.2.0/24 dev R2-eth0 scope link table 1")
    R2.cmd("ip route add default via 10.0.2.1 dev R2-eth0 table 1")
    R2.cmd("ip route add default scope global nexthop via 10.0.2.1 dev R2-eth0")
    
    C.cmd("ip rule add from 10.0.3.2 table 1")
    C.cmd("ip route add 10.0.3.0/24 dev C-eth0 scope link table 1")
    C.cmd("ip route add default via 10.0.3.1 dev C-eth0 table 1")
    C.cmd("ip route add default scope global nexthop via 10.0.3.1 dev C-eth0")
    
    controller = net.controllers[0]
    controller.start()
    
    sw1.start([controller])
    sw2.start([controller])
    sw3.start([controller])
    sw4.start([controller])

    CLI(net)
    net.stop()
