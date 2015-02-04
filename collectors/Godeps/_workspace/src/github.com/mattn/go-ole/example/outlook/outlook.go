package main

import (
	"fmt"
	"github.com/mattn/go-ole"
	"github.com/mattn/go-ole/oleutil"
)

func main() {
	ole.CoInitialize(0)
	unknown, _ := oleutil.CreateObject("Outlook.Application")
	outlook, _ := unknown.QueryInterface(ole.IID_IDispatch)
	ns := oleutil.MustCallMethod(outlook, "GetNamespace", "MAPI").ToIDispatch()
	folder := oleutil.MustCallMethod(ns, "GetDefaultFolder", 10).ToIDispatch()
	contacts := oleutil.MustCallMethod(folder, "Items").ToIDispatch()
	count := oleutil.MustGetProperty(contacts, "Count").Value().(int64)
	for i := int64(1); i <= count; i++ {
		item, err := oleutil.GetProperty(contacts, "Item", i)
		if err == nil && item.VT == ole.VT_DISPATCH {
			if value, err := oleutil.GetProperty(item.ToIDispatch(), "FullName"); err == nil {
				fmt.Println(value.Value())
			}
		}
	}
    oleutil.MustCallMethod(outlook, "Quit")
}
