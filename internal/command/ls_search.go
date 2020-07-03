package command

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/tickstep/cloudpan189-go/cloudpan"
	"github.com/tickstep/cloudpan189-go/cmder/cmdtable"
	"github.com/tickstep/cloudpan189-go/cmder/cmdutil/converter"
	"github.com/tickstep/cloudpan189-go/internal/config"
	"github.com/tickstep/cloudpan189-go/library/text"
	"os"
	"strconv"
)

type (
	// LsOptions 列目录可选项
	LsOptions struct {
		Total bool
	}

	// SearchOptions 搜索可选项
	SearchOptions struct {
		Total   bool
		Recurse bool
	}
)

const (
	opLs int = iota
	opSearch
)

func RunLs(targetPath string, lsOptions *LsOptions, orderBy cloudpan.OrderBy, orderSort cloudpan.OrderSort)  {
	activeUser := config.Config.ActiveUser()
	targetPath = activeUser.PathJoin(targetPath)
	if targetPath[len(targetPath) - 1] == '/' {
		targetPath = text.Substr(targetPath, 0, len(targetPath) - 1)
	}

	targetPathInfo, err := activeUser.PanClient().FileInfoByPath(targetPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	searchParam := cloudpan.NewFileSearchParam()
	searchParam.FileId = targetPathInfo.FileId
	searchParam.OrderBy = orderBy
	searchParam.OrderSort = orderSort
	if targetPathInfo.IsFolder {
		fileResult, err := activeUser.PanClient().FileSearch(searchParam)
		if err != nil {
			fmt.Println(err)
			return
		}
		renderTable(opLs, lsOptions.Total, targetPath, fileResult.Data)
	}
}


func renderTable(op int, isTotal bool, path string, files cloudpan.FileList) {
	tb := cmdtable.NewTable(os.Stdout)
	var (
		fN, dN   int64
		showPath string
	)

	switch op {
	case opLs:
		showPath = "文件(目录)"
	case opSearch:
		showPath = "路径"
	}

	if isTotal {
		tb.SetHeader([]string{"#", "file_id", "文件大小", "创建日期", "修改日期", showPath})
		tb.SetColumnAlignment([]int{tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
		for k, file := range files {
			if file.IsFolder {
				tb.Append([]string{strconv.Itoa(k), file.FileId, "-", file.CreateTime, file.LastOpTime, file.FileName + cloudpan.PathSeparator})
				continue
			}

			switch op {
			case opLs:
				tb.Append([]string{strconv.Itoa(k), file.FileId, converter.ConvertFileSize(file.FileSize, 2), file.CreateTime, file.LastOpTime, file.FileName})
			case opSearch:
				tb.Append([]string{strconv.Itoa(k), file.FileId, converter.ConvertFileSize(file.FileSize, 2), file.CreateTime, file.LastOpTime, file.Path})
			}
		}
		fN, dN = files.Count()
		tb.Append([]string{"", "", "总: " + converter.ConvertFileSize(files.TotalSize(), 2), "", "", "", fmt.Sprintf("文件总数: %d, 目录总数: %d", fN, dN)})
	} else {
		tb.SetHeader([]string{"#", "文件大小", "修改日期", showPath})
		tb.SetColumnAlignment([]int{tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
		for k, file := range files {
			if file.IsFolder {
				tb.Append([]string{strconv.Itoa(k), "-", file.LastOpTime, file.FileName + cloudpan.PathSeparator})
				continue
			}

			switch op {
			case opLs:
				tb.Append([]string{strconv.Itoa(k), converter.ConvertFileSize(file.FileSize, 2), file.LastOpTime, file.FileName})
			case opSearch:
				tb.Append([]string{strconv.Itoa(k), converter.ConvertFileSize(file.FileSize, 2), file.LastOpTime, file.Path})
			}
		}
		fN, dN = files.Count()
		tb.Append([]string{"", "总: " + converter.ConvertFileSize(files.TotalSize(), 2), "", fmt.Sprintf("文件总数: %d, 目录总数: %d", fN, dN)})
	}

	tb.Render()

	if fN+dN >= 60 {
		fmt.Printf("\n当前目录: %s\n", path)
	}

	fmt.Printf("----\n")
}
