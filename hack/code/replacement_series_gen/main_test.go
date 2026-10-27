// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package main

import (
	"reflect"
	"strings"
	"testing"
)

const sampleMarkdown = `# Guide

## Recommended Replacement VM Series

|Current VM Series | Target VM Series| Differences in Specification in Target VM*|
|--|--|--|
| D<br>Ds | Dsv5/Ddsv5<br>Dasv6 | Disk controller type: SCSI <br> Remote Storage: 3750 IOPS|
| Av2<br>Amv2 | Bsv2/Basv2 | Remote Storage Throughput: 3750 IOPS / 85 MBps
| NP-series<br>(Standard_NP10s<br>Standard_NP20s) | [NDv2](../ndv2-series.md)<br>[NCasT4_v3](../t4.md) | GPU type: V100 |

*Footnote.

## Next section

| ignored | ignored | ignored |
`

func TestParseReplacements(t *testing.T) {
	rows, err := parseReplacements(sampleMarkdown)
	if err != nil {
		t.Fatalf("parseReplacements: %v", err)
	}
	want := []replacement{
		{[]string{"D", "Ds"}, []string{"Dsv5/Ddsv5", "Dasv6"}, []string{"Disk controller type: SCSI", "Remote Storage: 3750 IOPS"}},
		{[]string{"Av2", "Amv2"}, []string{"Bsv2/Basv2"}, []string{"Remote Storage Throughput: 3750 IOPS / 85 MBps"}},
		{[]string{"NP"}, []string{"NDv2", "NCasT4_v3"}, []string{"GPU type: V100"}},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Errorf("rows = %#v, want %#v", rows, want)
	}
}

func TestParseReplacementsErrors(t *testing.T) {
	tests := map[string]string{
		"missing section":  "# Guide\n",
		"wrong cell count": "## Recommended Replacement VM Series\n\n| a | b |\n|--|--|\n| x | y |\n",
		"empty cell":       "## Recommended Replacement VM Series\n\n| a | b | c |\n|--|--|--|\n| D |  | diff |\n",
	}
	for name, md := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := parseReplacements(md); err == nil {
				t.Error("expected an error")
			}
		})
	}
}

func TestRenderIsValidGo(t *testing.T) {
	out, err := render([]replacement{{[]string{"D"}, []string{"Dsv5"}, []string{`a "quoted" diff`}}})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(string(out), "Code generated") {
		t.Error("missing generated-code header")
	}
}
