package goluahttp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aarzilli/golua/lua"
	"github.com/fiatjaf/lunatico"
)

type luaResponse struct {
	Status int
	Body string
	Size int
	Url string
}

var client = http.Client{}

var HTTP = map[string]lua.LuaGoFunction{
	"get": get,
	"delete": delete,
	"head": head,
	"patch": patch,
	"post": post,
	"put": put,
}

func get(L *lua.State) int {
	return doRequestAndPush(L, "get")
}

func delete(L *lua.State) int {
	return doRequestAndPush(L, "delete")
}

func head(L *lua.State) int {
	return doRequestAndPush(L, "head")
}

func patch(L *lua.State) int {
	return doRequestAndPush(L, "patch")
}

func post(L *lua.State) int {
	return doRequestAndPush(L, "post")
}

func put(L *lua.State) int {
	return doRequestAndPush(L, "put")
}

func forOption(options map[string]interface{}, option string, f func(string, string)) {
	if ops, ok := options[option].(map[string]string); ok {
		for op, value := range ops {
			f(op, value)
		}
	}
}

func doRequest(L *lua.State, method string) (res luaResponse, err error) {
	req, err := http.NewRequest(strings.ToUpper(method), L.ToString(1), nil)
	if err != nil {
		return
	}

	options := lunatico.ReadTable(L, -1).(map[string]interface{})
	if options != nil {
		forOption(options, "cookies", func(cookie string, value string) {
			req.AddCookie(&http.Cookie{Name: cookie, Value: value})
		})
		forOption(options, "query", func(query string, value string) {
			req.URL.Query().Add(query, value)
		})
		forOption(options, "headers", func(header string, value string) {
			req.Header.Set(header, value)
		})

		if body, ok := options["body"].(string); ok {
			req.ContentLength = int64(len(body))
			req.Body = io.NopCloser(strings.NewReader(body))
		}

		if timeout, ok := options["timeout"].(int64); ok {
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond * time.Duration(timeout))
			req = req.WithContext(ctx)
			defer cancel()
		}

		if user, ok := options["user"].(string); ok {
			if pass, ok := options["pass"].(string); ok {
				req.SetBasicAuth(user, pass)
			} 
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var lresp luaResponse
	lresp.Status = resp.StatusCode
	lresp.Body = string(body)
	lresp.Size = len(lresp.Body)
	return lresp, nil
}

func doRequestAndPush(L *lua.State, method string) int {
	resp, err := doRequest(L, method)
	if err != nil {
		L.PushNil()
		L.PushString(fmt.Sprintf("%s", err))
		return 2
	}

	lunatico.PushAny(L, resp)
	return 1
}
