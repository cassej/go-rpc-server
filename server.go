package main

import (
    "net"
    "bufio"
    "net/http"
    "io/ioutil"
)

type Request struct {
	Data    string      // Содержимое запроса
	Conn    net.Conn    // Соединение TCP, может быть nil для HTTP
	Writer  http.ResponseWriter // Writer для HTTP, может быть nil для TCP
}

func (r *Request) Response(message string) {
    if r.Conn != nil {
        // Обработка TCP соединения
        _, err := r.Conn.Write([]byte(message + "\n"))
        if err != nil {
            ErrorLog.Printf("Failed to write to TCP connection: %v", err)
        }
    } else if r.Writer != nil {
        // Обработка HTTP ответа
        messageBytes := []byte(message)

        if _, err := r.Writer.Write(messageBytes); err != nil {
            ErrorLog.Printf("Failed to write HTTP response: %v", err)
        }
    }
}

var requestQueue chan Request

func TCPServer () {
    address := config.Server.Host + ":" + config.Server.Port
	listener, err := net.Listen("tcp", address)

	if err != nil {
		ErrorLog.Fatal("Failed to start TCP server on %s: %v", address, err)
	}

	defer listener.Close()

	NoticeLog.Printf("TCP server listening on %s", address)
    for {
        conn, err := listener.Accept()
        if err != nil {
            ErrorLog.Printf("Failed to accept connection: %v", err)
            continue
        }


        defer conn.Close()
            reader := bufio.NewReader(conn)

            for {
                message, err := reader.ReadString('\n')  // Считывание до символа новой строки
                if err != nil {
                    ErrorLog.Printf("Failed to read from connection: %v", err)
                }

                if message == "" {
                    continue  // Пропускаем пустые сообщения
                }

                requestQueue <- Request{Data: message, Conn: conn}
            }
    }
}

func HTTPServer () {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "POST" {
            http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
            return
        }

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		if len(body) < 1 {
            http.Error(w, "Bad Request: Body is empty", http.StatusBadRequest)
            return
		} else {
            requestQueue <- Request{Data: string(body), Writer: w}
            return
		}
	})

	address := config.Server.Host + ":" + config.Server.Port
	NoticeLog.Printf("HTTP server listening on %s", address)

	if err := http.ListenAndServe(address, nil); err != nil {
		ErrorLog.Printf("Failed to start HTTP server on %s: %v", address, err)
	}
}