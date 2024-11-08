package main

import (
	"github.com/aarzilli/golua/lua"
	"github.com/mrnavastar/goluahttp"
)

const code = `
	local http = require("http")

	local req, err = http.get("https://www.google.com")
	print(req["body"])
	`

func main() {
	L := lua.NewState()
	defer L.Close()
	L.OpenLibs()
	
	L.RegisterLibrary("http", goluahttp.HTTP)

	if err := L.DoString(code); err != nil {
		panic(err)
	}
}
