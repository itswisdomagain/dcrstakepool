package controllers

import (
	mrand "math/rand"
	"net/http"
	"reflect"
	"sort"
	"testing"

	"github.com/decred/dcrd/chaincfg/v2"
)

func TestGetNetworkName(t *testing.T) {
	// First test that "testnet3" is translated to "testnet"
	cfg := Config{
		NetParams: chaincfg.TestNet3Params(),
	}

	mc := MainController{
		Cfg: &cfg,
	}

	netName := mc.getNetworkName()
	if netName != "testnet" {
		t.Errorf("Incorrect network name: expected %s, got %s", "testnet",
			netName)
	}

	// ensure "mainnet" is unaltered
	mc.Cfg.NetParams = chaincfg.MainNetParams()
	netName = mc.getNetworkName()
	if netName != "mainnet" {
		t.Errorf("Incorrect network name: expected %s, got %s", "mainnet",
			netName)
	}
}

func randHashString() string {
	var b [64]byte
	const hexvals = "123456789abcdef"
	for i := range b {
		b[i] = hexvals[mrand.Intn(len(hexvals))]
	}
	return string(b[:])
}

func TestSortByTicketHeight(t *testing.T) {
	// Create a large list of tickets to sort, voted over many blocks
	ticketCount, maxTxHeight := 55000, int64(123000)

	ticketInfoLive := make([]TicketInfo, 0, ticketCount)
	for i := 0; i < ticketCount; i++ {
		ticketInfoLive = append(ticketInfoLive, TicketInfo{
			TicketHeight: uint32(mrand.Int63n(maxTxHeight)),
			Ticket:       randHashString(), // could be nothing unless we sort with it
		})
	}

	// Make a copy to sort with ref method
	ticketInfoLive2 := make([]TicketInfo, len(ticketInfoLive))
	copy(ticketInfoLive2, ticketInfoLive)

	// Sort with ByTicketHeight, the test subject
	sort.Sort(ByTicketHeight(ticketInfoLive))

	// Sort using convenience function added in go1.8
	sort.Slice(ticketInfoLive2, func(i, j int) bool {
		return ticketInfoLive2[i].TicketHeight < ticketInfoLive2[j].TicketHeight
	})
	// compare
	if !reflect.DeepEqual(ticketInfoLive, ticketInfoLive2) {
		t.Error("Sort with ByTicketHeight failed")
	}

	// Check if sorted using convenience function added in go1.8
	if !sort.SliceIsSorted(ticketInfoLive, func(i, j int) bool {
		return ticketInfoLive[i].TicketHeight < ticketInfoLive[j].TicketHeight
	}) {
		t.Error("Sort with ByTicketHeight failed")
	}
}

func BenchmarkSortByTicketHeight100(b *testing.B)   { benchmarkSortByTicketHeight(100, b) }
func BenchmarkSortByTicketHeight500(b *testing.B)   { benchmarkSortByTicketHeight(500, b) }
func BenchmarkSortByTicketHeight1000(b *testing.B)  { benchmarkSortByTicketHeight(1000, b) }
func BenchmarkSortByTicketHeight2500(b *testing.B)  { benchmarkSortByTicketHeight(2500, b) }
func BenchmarkSortByTicketHeight5000(b *testing.B)  { benchmarkSortByTicketHeight(5000, b) }
func BenchmarkSortByTicketHeight10000(b *testing.B) { benchmarkSortByTicketHeight(10000, b) }
func BenchmarkSortByTicketHeight20000(b *testing.B) { benchmarkSortByTicketHeight(20000, b) }

func benchmarkSortByTicketHeight(ticketCount int, b *testing.B) {
	// Create a large list of tickets to sort, voted over many blocks
	maxTxHeight := int64(53000)

	ticketInfoLive := make([]TicketInfo, 0, ticketCount)
	for i := 0; i < ticketCount; i++ {
		ticketInfoLive = append(ticketInfoLive, TicketInfo{
			TicketHeight: uint32(mrand.Int63n(maxTxHeight)),
			Ticket:       randHashString(), // could be nothing unless we sort with it
		})
	}

	for i := 0; i < b.N; i++ {
		// Make a copy to sort
		ticketInfoLive2 := make([]TicketInfo, len(ticketInfoLive))
		copy(ticketInfoLive2, ticketInfoLive)

		// Sort with ByTicketHeight, the test subject
		sort.Sort(ByTicketHeight(ticketInfoLive2))
	}
}

func BenchmarkSortBySpentByHeight100(b *testing.B)   { benchmarkSortBySpentByHeight(100, b) }
func BenchmarkSortBySpentByHeight500(b *testing.B)   { benchmarkSortBySpentByHeight(500, b) }
func BenchmarkSortBySpentByHeight1000(b *testing.B)  { benchmarkSortBySpentByHeight(1000, b) }
func BenchmarkSortBySpentByHeight2500(b *testing.B)  { benchmarkSortBySpentByHeight(2500, b) }
func BenchmarkSortBySpentByHeight5000(b *testing.B)  { benchmarkSortBySpentByHeight(5000, b) }
func BenchmarkSortBySpentByHeight10000(b *testing.B) { benchmarkSortBySpentByHeight(10000, b) }
func BenchmarkSortBySpentByHeight20000(b *testing.B) { benchmarkSortBySpentByHeight(20000, b) }

func benchmarkSortBySpentByHeight(ticketCount int, b *testing.B) {
	// Create a large list of tickets to sort, voted over many blocks
	maxTxHeight := int64(53000)

	ticketInfoVoted := make([]TicketInfoHistoric, 0, ticketCount)
	for i := 0; i < ticketCount; i++ {
		ticketInfoVoted = append(ticketInfoVoted, TicketInfoHistoric{
			Ticket:        randHashString(), // could be nothing unless we sort with it
			SpentBy:       randHashString(),
			SpentByHeight: uint32(mrand.Int63n(maxTxHeight)),
			TicketHeight:  uint32(mrand.Int63n(maxTxHeight)),
		})
	}

	for i := 0; i < b.N; i++ {
		// Make a copy to sort
		ticketInfoVoted2 := make([]TicketInfoHistoric, len(ticketInfoVoted))
		copy(ticketInfoVoted2, ticketInfoVoted)

		// Sort with BySpentByHeight, the test subject
		sort.Sort(BySpentByHeight(ticketInfoVoted2))
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name, realIPHeader, realAddr, remoteAddr, wantAddr string
	}{{
		name:         "has realIPHeader default name",
		realIPHeader: "X-Real-IP",
		realAddr:     "240.111.3.145:3000",
		wantAddr:     "240.111.3.145",
	}, {
		name:         "has realIPHeader no port",
		realIPHeader: "X-Real-IP",
		realAddr:     "240.111.3.145",
		wantAddr:     "240.111.3.145",
	}, {
		name:         "has realIPHeader custom name",
		realIPHeader: "the real ip",
		realAddr:     "240.111.3.145:5454",
		wantAddr:     "240.111.3.145",
	}, {
		name:         "has realIPHeader host name",
		realIPHeader: "X-Real-IP",
		realAddr:     "hosting service",
		wantAddr:     "hosting service",
	}, {
		name:       "no realIPHeader has remoteAddr",
		remoteAddr: "240.111.3.145:80",
		wantAddr:   "240.111.3.145",
	}, {
		name:       "no realIPHeader has remoteAddr no port",
		remoteAddr: "240.111.3.145",
		wantAddr:   "240.111.3.145",
	}, {
		name:       "no realIPHeader has remoteAddr host name",
		remoteAddr: "hosting service",
		wantAddr:   "hosting service",
	}, {
		name:     "no realIPHeader no remoteAddr",
		wantAddr: "",
	}}

	r, _ := http.NewRequest("GET", "", nil)
	for _, test := range tests {
		requestHeader := make(http.Header)
		if test.realIPHeader != "" {
			requestHeader.Add(test.realIPHeader, test.realAddr)
		}
		r.RemoteAddr = test.remoteAddr
		r.Header = requestHeader
		addr := getClientIP(r, test.realIPHeader)
		if addr != test.wantAddr {
			t.Fatalf("expected \"%v\" for \"%v\" but got \"%v\"", test.wantAddr, test.name, addr)
		}
	}
}
