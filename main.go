package main

import (
	"fmt"
	"net/url"

	"github.com/webview/webview"

	_ "embed"
)

//go:embed index.html
var index string

func main() {
	initCAN()
	initEngine()
	debug := true
	println(index)
	w := webview.New(debug)
	defer w.Destroy()
	w.SetTitle("Liquid Propulsion CAN Tester")
	w.SetSize(800, 600, webview.HintNone)
	w.Bind("state", func() *FakeEngine {
		return CurrentFakeEngine
	})
	w.Bind("addNode", CurrentFakeEngine.addNode)
	w.Bind("removeNode", CurrentFakeEngine.removeNode)
	w.Bind("addSensor", CurrentFakeEngine.addSensor)
	w.Bind("removeSensor", CurrentFakeEngine.removeSensor)
	w.Navigate(fmt.Sprintf("data:text/html,%s", url.PathEscape(index)))
	w.Run()
}
