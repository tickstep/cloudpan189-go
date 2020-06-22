package cmdtable

import (
	"github.com/olekukonko/tablewriter"
	"io"
)

type PCSTable struct {
	*tablewriter.Table
}

// NewTable 预设了一些配置
func NewTable(wt io.Writer) PCSTable {
	tb := tablewriter.NewWriter(wt)
	tb.SetAutoWrapText(false)
	tb.SetBorder(false)
	tb.SetHeaderLine(false)
	tb.SetColumnSeparator("")
	return PCSTable{tb}
}
