package command

import (
	"fmt"
	"os"
	"text/tabwriter"

	device "github.com/gotracker/gosound"
	"github.com/spf13/cobra"

	"gotracker/internal/output"
)

func init() {
	deviceCmd.AddCommand(deviceListCmd)
}

var (
	deviceListCmd = &cobra.Command{
		Use:   "list",
		Short: "List the available output devices",
		Long:  `List the available output devices.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var pmap []map[string]output.DeviceInfo
			for k, v := range output.GetOutputDevices() {
				for v.Priority >= len(pmap) {
					pmap = append(pmap, make(map[string]output.DeviceInfo))
				}
				pmap[v.Priority][k] = v
			}

			if len(pmap) > 0 {
				fmt.Println("Devices listed in ascending priority (higher priority value = more likely to be chosen):")
				fmt.Println()
				tw := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
				fmt.Fprintln(tw, "DEVICE\tPriority\tKind")
				fmt.Fprintln(tw, "======\t========\t====")
				for _, m := range pmap {
					for k, v := range m {
						var kind string
						switch v.Kind {
						case device.KindNone:
							kind = "None"
						case device.KindSoundCard:
							kind = "Sound Card"
						case device.KindFile:
							kind = "File Writer"
						default:
							kind = "Unknown"
						}
						fmt.Fprintf(tw, "%v\t%s\t%d\n", k, kind, v.Priority)
					}
				}
				fmt.Fprintln(tw)
				if err := tw.Flush(); err != nil {
					return err
				}
			} else {
				fmt.Println("no valid devices to list!")
			}
			return nil
		},
	}
)
