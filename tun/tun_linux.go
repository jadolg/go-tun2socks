package tun

import (
	"io"
	"net"

	"github.com/jackpal/gateway"
	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
)

func ParseIPv4Mask(s string) net.IPMask {
	mask := net.ParseIP(s)
	if mask == nil {
		return nil
	}
	return net.IPv4Mask(mask[12], mask[13], mask[14], mask[15])
}

func OpenTunDevice(name, addr, gw, mask string, dnsServers []string, persist bool, SkipGatewayFor string) (io.ReadWriteCloser, error) {
	cfg := water.Config{
		DeviceType: water.TUN,
	}
	cfg.Name = name
	cfg.Persist = persist
	tunDev, err := water.New(cfg)
	if err != nil {
		return nil, err
	}
	name = tunDev.Name()

	tunInterface, err := netlink.LinkByName(name)
	if err != nil {
		return nil, err
	}

	var address = &net.IPNet{IP: net.ParseIP(addr), Mask: ParseIPv4Mask(mask)}

	err = netlink.AddrAdd(tunInterface, &netlink.Addr{IPNet: address})
	if err != nil {
		return nil, err
	}

	err = netlink.LinkSetUp(tunInterface)
	if err != nil {
		return nil, err
	}

	if SkipGatewayFor != "" {
		defaultGateway, err := gateway.DiscoverGateway()
		if err != nil {
			return nil, err
		}

		_, skipDst, err := net.ParseCIDR(SkipGatewayFor)
		if err != nil {
			return nil, err
		}

		err = netlink.RouteAdd(&netlink.Route{
			Dst: skipDst,
			Gw:  defaultGateway,
		})

		err = netlink.RouteAdd(&netlink.Route{
			Gw: net.ParseIP(gw),
		})

		if err != nil {
			return nil, err
		}
	}

	return tunDev, nil
}
