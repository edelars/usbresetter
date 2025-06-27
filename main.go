package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/google/gousb"
	probing "github.com/prometheus-community/pro-bing"
)

const t1 = time.Second * time.Duration(10)

func main() {
	ticker := time.NewTicker(t1)
	pinger, err := probing.NewPinger("192.168.1.1")
	if err != nil {
		panic(err)
	}

	// Listen for Ctrl-C.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			pinger.Stop()
			ticker.Stop()
		}
	}()

	pinger.OnRecv = func(pkt *probing.Packet) {
		fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v\n",
			pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt)
	}
	pinger.Count = 3

	pinger.OnFinish = func(stats *probing.Statistics) {
		if stats.PacketLoss == 3 {
			reset()
		}
	}

	fmt.Println("usbresetter started")

	reset()

	for {
		select {
		case <-ticker.C:
			pinger.Count = 3

			err = pinger.Run()
			if err != nil {
				fmt.Println(err)
			}

			ticker.Reset(t1)
		}
	}
}

func reset() {
	ctx := gousb.NewContext()
	defer ctx.Close() // Ensure the context is closed when done

	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		return desc.Vendor == 0x0bda && desc.Product == 0xf179
	})
	if err != nil {
		fmt.Println("no device found")
		// Handle error
	}
	if len(devs) == 0 {
		fmt.Println("no device found: len 0")
		// Handle case where device is not found
	}

	// Assuming you want to reset the first found device
	dev := devs[0]
	fmt.Println("Trying to reset %w", dev.Reset().Error())
	defer dev.Close() // Close the device handle when done
}
