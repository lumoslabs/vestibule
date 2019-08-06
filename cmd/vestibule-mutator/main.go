//go:generate go run -tags=build data/generate.go
package main

func main() {
  srv := kubernetes.NewWebhookServer()
}
