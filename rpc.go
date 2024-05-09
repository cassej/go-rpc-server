package main

import (
    "plugin"
    "path/filepath"
    "os"
    "strings"
)

type Method struct {
    Name        string
    //Contract    func() (Contract, error)  // Загружает структуру контракта
    //Validate    func(params map[string]interface{}) error // Валидирует параметры
    Execute     func(params map[string]interface{}) (interface{}) // Выполняет метод
}

var methods map[string]Method

type RPCRequest struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
	Id     *int                   `json:"id"`
}

type RPCResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  *RPCError   `json:"error,omitempty"`
	Id     *int        `json:"id"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type RPC struct {}

var rpc RPC

func (r RPC) init() {
    methods = make(map[string]Method)
    err := filepath.Walk("methods", func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if info.IsDir() || filepath.Ext(path) != ".so" {
            return nil
        }

        p, err := plugin.Open(path)
        if err != nil {
            return err
        }

        execute, err := p.Lookup("Method")
        if err != nil {
            return err
        }
        relPath, _ := filepath.Rel("methods", path)
        methodName := strings.TrimSuffix(relPath, filepath.Ext(relPath))
        methodName = strings.ReplaceAll(methodName, string(os.PathSeparator), ".")

        methods[methodName] = Method{
            Name:            methodName,
            Execute:         execute.(func(map[string]interface{}) (interface{}))}

        return nil
    })

    if err != nil {
        ErrorLog.Printf("Failed to init Methods: %v", err)
    }
}

func (r RPC) handle(request RPCRequest) *RPCResponse {
    if method, ok := methods[request.Method]; ok {
        return &RPCResponse{Id: request.Id, Result: method.Execute(request.Params)}
    }

    return &RPCResponse{Id: request.Id, Error: &RPCError{-32601, "Method not found"}}
}


