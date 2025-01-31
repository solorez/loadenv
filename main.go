package main

import (
	"log"
	"os"
	"time"

	"github.com/solorez/loadenv/loadenv"
)

func main() {
	cfg := loadenv.Config{
		FilePath:    ".env",
		HotReload:   true,
		ReloadDelay: 1 * time.Second,
		Logger:      log.New(os.Stdout, "[APP] ", log.LstdFlags),
	}

	if err := loadenv.InitEnv(cfg); err != nil {
		log.Fatal("Failed to init env:", err)
	}
	defer loadenv.Close()

	// 保持程序运行
	for {
		time.Sleep(5 * time.Minute)
	}
}
