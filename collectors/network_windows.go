package collectors

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/StackExchange/wmi"
	"github.com/oliveagle/go-collectors/datapoint"
	"github.com/oliveagle/go-collectors/metadata"
	"github.com/oliveagle/go-collectors/slog"
)

func init() {
	collectors = append(collectors, &IntervalCollector{F: c_network_windows, init: winNetworkInit})

	c := &IntervalCollector{
		F: c_network_team_windows,
	}
	// Make sure MSFT_NetImPlatAdapter and MSFT_NetAdapterStatisticsSettingData
	// are valid WMI classes when initializing c_network_team_windows
	c.init = func() {
		var dstTeamNic []MSFT_NetLbfoTeamNic
		var dstStats []MSFT_NetAdapterStatisticsSettingData
		queryTeamAdapter = wmi.CreateQuery(&dstTeamNic, "")
		queryTeamStats = wmi.CreateQuery(&dstStats, "")
		c.Enable = func() bool {
			errTeamNic := queryWmiNamespace(queryTeamAdapter, &dstTeamNic, namespaceStandardCimv2)
			errStats := queryWmiNamespace(queryTeamStats, &dstStats, namespaceStandardCimv2)
			return errTeamNic == nil && errStats == nil
		}
	}
	collectors = append(collectors, c)
}

var (
	queryTeamStats         string
	queryTeamAdapter       string
	namespaceStandardCimv2 = "root\\StandardCimv2"
	interfaceExclusions    = regexp.MustCompile("isatap|Teredo")

	// instanceNameToUnderscore matches '#' '/' and '\' for replacing with '_'.
	instanceNameToUnderscore         = regexp.MustCompile("[#/\\\\]")
	mNicInstanceNameToInterfaceIndex = make(map[string]string)
)

// winNetworkToInstanceName converts a Network Adapter Name to the InstanceName
// that is used in Win32_PerfRawData_Tcpip_NetworkInterface.
func winNetworkToInstanceName(Name string) string {
	instanceName := Name
	instanceName = strings.Replace(instanceName, "(", "[", -1)
	instanceName = strings.Replace(instanceName, ")", "]", -1)
	instanceName = instanceNameToUnderscore.ReplaceAllString(instanceName, "_")
	return instanceName
}

// winNetworkInit maintains a mapping of InstanceName to InterfaceIndex
func winNetworkInit() {
	update := func() {
		var dstNetworkAdapter []Win32_NetworkAdapter
		q := wmi.CreateQuery(&dstNetworkAdapter, "WHERE PhysicalAdapter=True and MACAddress <> null")
		err := queryWmi(q, &dstNetworkAdapter)
		if err != nil {
			slog.Error(err)
			return
		}
		for _, nic := range dstNetworkAdapter {
			var iface = fmt.Sprint("Interface", nic.InterfaceIndex)
			//Get PnPName using Win32_PnPEntity class
			var pnpname = ""
			var escapeddeviceid = strings.Replace(nic.PNPDeviceID, "\\", "\\\\", -1)
			var filter = fmt.Sprintf("WHERE DeviceID='%s'", escapeddeviceid)
			var dstPnPName []Win32_PnPEntity
			q = wmi.CreateQuery(&dstPnPName, filter)
			err = queryWmi(q, &dstPnPName)
			if err != nil {
				slog.Error(err)
				return
			}
			for _, pnp := range dstPnPName { //Really should be a single item
				pnpname = pnp.Name
			}
			if pnpname == "" {
				slog.Errorf("%s cannot find Win32_PnPEntity %s", iface, filter)
				continue
			}

			//Convert to instance name (see http://goo.gl/jfq6pq )
			instanceName := winNetworkToInstanceName(pnpname)
			mNicInstanceNameToInterfaceIndex[instanceName] = iface
		}
	}
	update()
	go func() {
		for _ = range time.Tick(time.Minute * 5) {
			update()
		}
	}()
}

func c_network_windows() (datapoint.MultiDataPoint, error) {
	var dstStats []Win32_PerfRawData_Tcpip_NetworkInterface
	var q = wmi.CreateQuery(&dstStats, "")
	err := queryWmi(q, &dstStats)
	if err != nil {
		return nil, err
	}

	var md datapoint.MultiDataPoint
	for _, nicStats := range dstStats {
		if interfaceExclusions.MatchString(nicStats.Name) {
			continue
		}

		iface := mNicInstanceNameToInterfaceIndex[nicStats.Name]
		if iface == "" {
			continue
		}
		//This does NOT include TEAM network adapters. Those will go to os.net.bond
		tagsIn := datapoint.TagSet{"iface": iface, "direction": "in"}
		tagsOut := datapoint.TagSet{"iface": iface, "direction": "out"}
		Add(&md, "win.net.ifspeed", nicStats.CurrentBandwidth, datapoint.TagSet{"iface": iface}, metadata.Gauge, metadata.BitsPerSecond, descWinNetCurrentBandwidth)
		Add(&md, "win.net.bytes", nicStats.BytesReceivedPersec, tagsIn, metadata.Counter, metadata.BytesPerSecond, descWinNetBytesReceivedPersec)
		Add(&md, "win.net.bytes", nicStats.BytesSentPersec, tagsOut, metadata.Counter, metadata.BytesPerSecond, descWinNetBytesSentPersec)
		Add(&md, "win.net.packets", nicStats.PacketsReceivedPersec, tagsIn, metadata.Counter, metadata.PerSecond, descWinNetPacketsReceivedPersec)
		Add(&md, "win.net.packets", nicStats.PacketsSentPersec, tagsOut, metadata.Counter, metadata.PerSecond, descWinNetPacketsSentPersec)
		Add(&md, "win.net.dropped", nicStats.PacketsOutboundDiscarded, tagsOut, metadata.Counter, metadata.PerSecond, descWinNetPacketsOutboundDiscarded)
		Add(&md, "win.net.dropped", nicStats.PacketsReceivedDiscarded, tagsIn, metadata.Counter, metadata.PerSecond, descWinNetPacketsReceivedDiscarded)
		Add(&md, "win.net.errs", nicStats.PacketsOutboundErrors, tagsOut, metadata.Counter, metadata.PerSecond, descWinNetPacketsOutboundErrors)
		Add(&md, "win.net.errs", nicStats.PacketsReceivedErrors, tagsIn, metadata.Counter, metadata.PerSecond, descWinNetPacketsReceivedErrors)
		Add(&md, osNetBytes, nicStats.BytesReceivedPersec, tagsIn, metadata.Counter, metadata.BytesPerSecond, osNetBytesDesc)
		Add(&md, osNetBytes, nicStats.BytesSentPersec, tagsOut, metadata.Counter, metadata.BytesPerSecond, osNetBytesDesc)
		Add(&md, osNetPackets, nicStats.PacketsReceivedPersec, tagsIn, metadata.Counter, metadata.PerSecond, osNetPacketsDesc)
		Add(&md, osNetPackets, nicStats.PacketsSentPersec, tagsOut, metadata.Counter, metadata.PerSecond, osNetPacketsDesc)
		Add(&md, osNetDropped, nicStats.PacketsOutboundDiscarded, tagsOut, metadata.Counter, metadata.PerSecond, osNetDroppedDesc)
		Add(&md, osNetDropped, nicStats.PacketsReceivedDiscarded, tagsIn, metadata.Counter, metadata.PerSecond, osNetDroppedDesc)
		Add(&md, osNetErrors, nicStats.PacketsOutboundErrors, tagsOut, metadata.Counter, metadata.PerSecond, osNetErrorsDesc)
		Add(&md, osNetErrors, nicStats.PacketsReceivedErrors, tagsIn, metadata.Counter, metadata.PerSecond, osNetErrorsDesc)
	}
	return md, nil
}

const (
	descWinNetCurrentBandwidth         = "Estimate of the interface's current bandwidth in bits per second (bps). For interfaces that do not vary in bandwidth or for those where no accurate estimation can be made, this value is the nominal bandwidth."
	descWinNetBytesReceivedPersec      = "Bytes Received/sec is the rate at which bytes are received over each network adapter, including framing characters. Network Interface\\Bytes Received/sec is a subset of Network Interface\\Bytes Total/sec."
	descWinNetBytesSentPersec          = "Bytes Sent/sec is the rate at which bytes are sent over each network adapter, including framing characters. Network Interface\\Bytes Sent/sec is a subset of Network Interface\\Bytes Total/sec."
	descWinNetPacketsReceivedPersec    = "Packets Received/sec is the rate at which packets are received on the network interface."
	descWinNetPacketsSentPersec        = "Packets Sent/sec is the rate at which packets are sent on the network interface."
	descWinNetPacketsOutboundDiscarded = "Packets Outbound Discarded is the number of outbound packets that were chosen to be discarded even though no errors had been detected to prevent transmission. One possible reason for discarding packets could be to free up buffer space."
	descWinNetPacketsReceivedDiscarded = "Packets Received Discarded is the number of inbound packets that were chosen to be discarded even though no errors had been detected to prevent their delivery to a higher-layer protocol.  One possible reason for discarding packets could be to free up buffer space."
	descWinNetPacketsOutboundErrors    = "Packets Outbound Errors is the number of outbound packets that could not be transmitted because of errors."
	descWinNetPacketsReceivedErrors    = "Packets Received Errors is the number of inbound packets that contained errors preventing them from being deliverable to a higher-layer protocol."
)

type Win32_PnPEntity struct {
	Name string //Intel(R) Gigabit ET Quad Port Server Adapter #3
}

type Win32_NetworkAdapter struct {
	Description    string //Intel(R) Gigabit ET Quad Port Server Adapter (no index)
	InterfaceIndex uint32
	PNPDeviceID    string
}

type Win32_PerfRawData_Tcpip_NetworkInterface struct {
	CurrentBandwidth         uint32
	BytesReceivedPersec      uint32
	BytesSentPersec          uint32
	Name                     string
	PacketsOutboundDiscarded uint32
	PacketsOutboundErrors    uint32
	PacketsReceivedDiscarded uint32
	PacketsReceivedErrors    uint32
	PacketsReceivedPersec    uint32
	PacketsSentPersec        uint32
}

// c_network_team_windows will add metrics for team network adapters from
// MSFT_NetAdapterStatisticsSettingData for any adapters that are in
// MSFT_NetLbfoTeamNic and have a valid instanceName.
func c_network_team_windows() (datapoint.MultiDataPoint, error) {
	//TODO: minimal supported server: 2012, minimal supported client 2008
	var dstTeamNic []*MSFT_NetLbfoTeamNic
	err := queryWmiNamespace(queryTeamAdapter, &dstTeamNic, namespaceStandardCimv2)
	if err != nil {
		return nil, err
	}

	var dstStats []MSFT_NetAdapterStatisticsSettingData
	err = queryWmiNamespace(queryTeamStats, &dstStats, namespaceStandardCimv2)
	if err != nil {
		return nil, err
	}

	mDescriptionToTeamNic := make(map[string]*MSFT_NetLbfoTeamNic)
	for _, teamNic := range dstTeamNic {
		mDescriptionToTeamNic[teamNic.InterfaceDescription] = teamNic
	}

	var md datapoint.MultiDataPoint
	for _, nicStats := range dstStats {
		TeamNic := mDescriptionToTeamNic[nicStats.InterfaceDescription]
		if TeamNic == nil {
			continue
		}

		instanceName := winNetworkToInstanceName(nicStats.InterfaceDescription)
		iface := mNicInstanceNameToInterfaceIndex[instanceName]
		if iface == "" {
			continue
		}
		tagsIn := datapoint.TagSet{"iface": iface, "direction": "in"}
		tagsOut := datapoint.TagSet{"iface": iface, "direction": "out"}
		linkSpeed := math.Min(float64(TeamNic.ReceiveLinkSpeed), float64(TeamNic.Transmitlinkspeed))
		Add(&md, "win.net.bond.ifspeed", linkSpeed, datapoint.TagSet{"iface": iface}, metadata.Gauge, metadata.BitsPerSecond, descWinNetTeamlinkspeed)
		Add(&md, "win.net.bond.bytes", nicStats.ReceivedBytes, tagsIn, metadata.Counter, metadata.BytesPerSecond, descWinNetTeamReceivedBytes)
		Add(&md, "win.net.bond.bytes", nicStats.SentBytes, tagsOut, metadata.Counter, metadata.BytesPerSecond, descWinNetTeamSentBytes)
		Add(&md, "win.net.bond.bytes_unicast", nicStats.ReceivedUnicastBytes, tagsIn, metadata.Counter, metadata.BytesPerSecond, descWinNetTeamReceivedUnicastBytes)
		Add(&md, "win.net.bond.bytes_unicast", nicStats.SentUnicastBytes, tagsOut, metadata.Counter, metadata.BytesPerSecond, descWinNetTeamSentUnicastBytes)
		Add(&md, "win.net.bond.bytes_broadcast", nicStats.ReceivedBroadcastBytes, tagsIn, metadata.Counter, metadata.BytesPerSecond, descWinNetTeamReceivedBroadcastBytes)
		Add(&md, "win.net.bond.bytes_broadcast", nicStats.SentBroadcastBytes, tagsOut, metadata.Counter, metadata.BytesPerSecond, descWinNetTeamSentBroadcastBytes)
		Add(&md, "win.net.bond.bytes_multicast", nicStats.ReceivedMulticastBytes, tagsIn, metadata.Counter, metadata.BytesPerSecond, descWinNetTeamReceivedMulticastBytes)
		Add(&md, "win.net.bond.bytes_multicast", nicStats.SentMulticastBytes, tagsOut, metadata.Counter, metadata.BytesPerSecond, descWinNetTeamSentMulticastBytes)
		Add(&md, "win.net.bond.packets_unicast", nicStats.ReceivedUnicastPackets, tagsIn, metadata.Counter, metadata.PerSecond, descWinNetTeamReceivedUnicastPackets)
		Add(&md, "win.net.bond.packets_unicast", nicStats.SentUnicastPackets, tagsOut, metadata.Counter, metadata.PerSecond, descWinNetTeamSentUnicastPackets)
		Add(&md, "win.net.bond.dropped", nicStats.ReceivedDiscardedPackets, tagsIn, metadata.Counter, metadata.PerSecond, descWinNetTeamReceivedDiscardedPackets)
		Add(&md, "win.net.bond.dropped", nicStats.OutboundDiscardedPackets, tagsOut, metadata.Counter, metadata.PerSecond, descWinNetTeamOutboundDiscardedPackets)
		Add(&md, "win.net.bond.errs", nicStats.ReceivedPacketErrors, tagsIn, metadata.Counter, metadata.PerSecond, descWinNetTeamReceivedPacketErrors)
		Add(&md, "win.net.bond.errs", nicStats.OutboundPacketErrors, tagsOut, metadata.Counter, metadata.PerSecond, descWinNetTeamOutboundPacketErrors)
		Add(&md, "win.net.bond.packets_multicast", nicStats.ReceivedMulticastPackets, tagsIn, metadata.Counter, metadata.PerSecond, descWinNetTeamReceivedMulticastPackets)
		Add(&md, "win.net.bond.packets_multicast", nicStats.SentMulticastPackets, tagsOut, metadata.Counter, metadata.PerSecond, descWinNetTeamSentMulticastPackets)
		Add(&md, "win.net.bond.packets_broadcast", nicStats.ReceivedBroadcastPackets, tagsIn, metadata.Counter, metadata.PerSecond, descWinNetTeamReceivedBroadcastPackets)
		Add(&md, "win.net.bond.packets_broadcast", nicStats.SentBroadcastPackets, tagsOut, metadata.Counter, metadata.PerSecond, descWinNetTeamSentBroadcastPackets)
		//Todo: add os.net.bond metrics once we confirm they have the same metadata
	}
	return md, nil
}

const (
	descWinNetTeamlinkspeed                = "The link speed of the adapter in bits per second."
	descWinNetTeamReceivedBytes            = "The number of bytes of data received without errors through this interface. This value includes bytes in unicast, broadcast, and multicast packets."
	descWinNetTeamReceivedUnicastPackets   = "The number of unicast packets received without errors through this interface."
	descWinNetTeamReceivedMulticastPackets = "The number of multicast packets received without errors through this interface."
	descWinNetTeamReceivedBroadcastPackets = "The number of broadcast packets received without errors through this interface."
	descWinNetTeamReceivedUnicastBytes     = "The number of unicast bytes received without errors through this interface."
	descWinNetTeamReceivedMulticastBytes   = "The number of multicast bytes received without errors through this interface."
	descWinNetTeamReceivedBroadcastBytes   = "The number of broadcast bytes received without errors through this interface."
	descWinNetTeamReceivedDiscardedPackets = "The number of inbound packets which were chosen to be discarded even though no errors were detected to prevent the packets from being deliverable to a higher-layer protocol."
	descWinNetTeamReceivedPacketErrors     = "The number of incoming packets that were discarded because of errors."
	descWinNetTeamSentBytes                = "The number of bytes of data transmitted without errors through this interface. This value includes bytes in unicast, broadcast, and multicast packets."
	descWinNetTeamSentUnicastPackets       = "The number of unicast packets transmitted without errors through this interface."
	descWinNetTeamSentMulticastPackets     = "The number of multicast packets transmitted without errors through this interface."
	descWinNetTeamSentBroadcastPackets     = "The number of broadcast packets transmitted without errors through this interface."
	descWinNetTeamSentUnicastBytes         = "The number of unicast bytes transmitted without errors through this interface."
	descWinNetTeamSentMulticastBytes       = "The number of multicast bytes transmitted without errors through this interface."
	descWinNetTeamSentBroadcastBytes       = "The number of broadcast bytes transmitted without errors through this interface."
	descWinNetTeamOutboundDiscardedPackets = "The number of outgoing packets that were discarded even though they did not have errors."
	descWinNetTeamOutboundPacketErrors     = "The number of outgoing packets that were discarded because of errors."
)

type MSFT_NetLbfoTeamNic struct {
	Team                 string
	Name                 string
	ReceiveLinkSpeed     uint64
	Transmitlinkspeed    uint64
	InterfaceDescription string
}

type MSFT_NetAdapterStatisticsSettingData struct {
	InstanceID               string
	Name                     string
	InterfaceDescription     string
	ReceivedBytes            uint64
	ReceivedUnicastPackets   uint64
	ReceivedMulticastPackets uint64
	ReceivedBroadcastPackets uint64
	ReceivedUnicastBytes     uint64
	ReceivedMulticastBytes   uint64
	ReceivedBroadcastBytes   uint64
	ReceivedDiscardedPackets uint64
	ReceivedPacketErrors     uint64
	SentBytes                uint64
	SentUnicastPackets       uint64
	SentMulticastPackets     uint64
	SentBroadcastPackets     uint64
	SentUnicastBytes         uint64
	SentMulticastBytes       uint64
	SentBroadcastBytes       uint64
	OutboundDiscardedPackets uint64
	OutboundPacketErrors     uint64
}
