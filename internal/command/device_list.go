package command

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	device "github.com/gotracker/gosound"
	"github.com/spf13/cobra"

	"gotracker/internal/output"
)

var (
	deviceListAddHeader bool   = true
	deviceListFormat    string = "human"
)

func init() {
	deviceListCmd.Flags().BoolVarP(&deviceListAddHeader, "add-header", "H", deviceListAddHeader, "add header row(s) for formats that support it")
	deviceListCmd.Flags().StringVarP(&deviceListFormat, "format", "f", deviceListFormat, "format of output {human, csv, json}")

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

			var err error
			switch deviceListFormat {
			case "json":
				jw := json.NewEncoder(os.Stdout)
				list := []map[string]interface{}{}
				if err = deviceListSerialized(pmap, func(vals map[string]interface{}) error {
					list = append(list, vals)
					return nil
				}); err != nil {
					break
				}
				err = jw.Encode(list)
			case "csv":
				fieldOrder := []string{"device", "kind", "priority", "is_default"}
				cw := csv.NewWriter(os.Stdout)
				if deviceListAddHeader {
					if err = cw.Write(fieldOrder); err != nil {
						break
					}
				}
				if err = deviceListSerialized(pmap, func(vals map[string]interface{}) error {
					var fields []string
					for _, f := range fieldOrder {
						fields = append(fields, fmt.Sprintf("%v", vals[f]))
					}
					return cw.Write(fields)
				}); err != nil {
					break
				}
				cw.Flush()
			default:
				err = deviceListHuman(pmap)
			}

			return err
		},
	}
)

type recordFunc func(vals map[string]interface{}) error

func deviceListSerialized(pmap []map[string]output.DeviceInfo, recordFunc recordFunc) error {
	for _, m := range pmap {
		for k, v := range m {
			var kind string
			switch v.Kind {
			case device.KindNone:
				kind = "none"
			case device.KindSoundCard:
				kind = "sound-card"
			case device.KindFile:
				kind = "file-writer"
			default:
				kind = "unknown"
			}
			vals := make(map[string]interface{})
			vals["device"] = k
			vals["kind"] = kind
			vals["priority"] = v.Priority
			vals["is_default"] = (k == output.DefaultOutputDeviceName)
			if err := recordFunc(vals); err != nil {
				return err
			}
		}
	}
	return nil
}

func deviceListHuman(pmap []map[string]output.DeviceInfo) error {
	if len(pmap) == 0 {
		fmt.Println("no valid devices to list!")
		return nil
	}

	fmt.Println("Devices are listed in ascending priority (greater priority value = more likely to be chosen):")
	fmt.Println()

	tw := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	if deviceListAddHeader {
		fmt.Fprintln(tw, "DEVICE\tKind\tPriority\tDefault?")
		fmt.Fprintln(tw, "======\t====\t========\t========")
	}
	if err := deviceListSerialized(pmap, func(vals map[string]interface{}) error {
		name := vals["device"].(string)
		kind := vals["kind"].(string)
		priority := vals["priority"].(int)
		var defaultStr string
		if vals["is_default"].(bool) {
			defaultStr = "*"
		}
		_, err := fmt.Fprintf(tw, "%s\t%s\t%d\t%s\n", name, kind, priority, defaultStr)
		return err
	}); err != nil {
		return err
	}
	fmt.Fprintln(tw)

	return tw.Flush()
}
