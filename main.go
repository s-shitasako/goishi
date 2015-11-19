package main

import (
  "crypto/tls"
  "fmt"
  "io"
  "net"
  "os"
  "strconv"
)

func main() {
  port, target, cert := loadArgs(os.Args)
  ln, err := tls.Listen("tcp", fmt.Sprintf(":%d", port), &tls.Config{
    Certificates: []tls.Certificate{cert},
  })
  if err != nil {
    fmt.Println("Failed to listen SSL port:", err)
    os.Exit(1)
  }
  for {
    conn, err := ln.Accept()
    if err != nil {
      fmt.Println("Error:", err)
      os.Exit(1)
    }
    go serve(conn, target)
  }
}

func serve(c net.Conn, target string) {
  notStarted := true
  defer func(){
    if err := recover(); err != nil {
      fmt.Println("Recovered from panic:", err)
    }
    if notStarted {
      c.Close()
    }
  }()
  fmt.Println("Access from", c.RemoteAddr())

  s, err := net.Dial("tcp", target)
  if err != nil {
    fmt.Println("Failed to access target:", err)
    return
  }

  convey := func(s1, s2 net.Conn, name string){
    defer func(){
      if err := recover(); err != nil {
        fmt.Println("Recovered from panic caused by", name, ":", err)
      }
    }()
    defer s1.Close()

    buf := make([]byte, 65536)
    for {
      l, err := s1.Read(buf)
      if err != nil && err != io.EOF {
        fmt.Println(name, "read error:", err)
        return
      }
      s2.Write(buf[:l])
    }
  }
  notStarted = false
  go convey(c, s, "client")
  go convey(s, c, "server")
}

func loadArgs(args []string) (int, string, tls.Certificate) {
  rhost := "localhost"
  certFile := "./cert.pem"
  keyFile := "./key.pem"
  switch len(args) {
  case 6:
    certFile = args[4]
    keyFile = args[5]
    fallthrough
  case 4:
    rhost = args[3]
    fallthrough
  case 3:
    rport, err := strconv.Atoi(args[2])
    if rport <= 0 || err != nil {
      fmt.Println("target-port is bad format.", err)
      os.Exit(1)
    }
    port, err := strconv.Atoi(args[1])
    if port <= 0 || err != nil {
      fmt.Println("ssl-port is bad format.", err)
      os.Exit(1)
    }
    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
      fmt.Println("Failed to load cert and/or key:", err)
      os.Exit(1)
    }
    return port, fmt.Sprintf("%s:%d", rhost, rport), cert
  default:
    fmt.Println("usage: goishi ssl-port target-port [ target-host [ cert-path key-path ] ]")
    fmt.Println("e.g. goishi 8080 8000")
    fmt.Println("e.g. goishi 443 80 www.example.com")
    fmt.Println("e.g. goishi 443 80 www.example.com cert/cert.pem cert/key.pem")
    os.Exit(1)
  }
  return 0, "", tls.Certificate{}
}
