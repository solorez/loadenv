package main

import (
	"log"
	"os"
	"sort"
	"strings"
	"time"
	// "github.com/solorez/loadenv"
)

func main() {
	cfg := Config{
		FilePath:    ".env",
		HotReload:   true,
		ReloadDelay: 1 * time.Second,
		Logger:      log.New(os.Stdout, "[APP] ", log.LstdFlags),
	}

	if err := InitEnv(cfg); err != nil {
		log.Fatal("Failed to init env:", err)
	}
	defer Close()

	// 读取 .env 文件内容
	envContent, err := os.ReadFile(".env")
	if err != nil {
		log.Fatal("Failed to read .env file:", err)
	}

	// 解析 .env 文件内容
	envMap := make(map[string]string)
	lines := strings.Split(string(envContent), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			envMap[key] = value
		}
	}

	// 获取并排序 .env 文件中定义的环境变量
	var envVars []string
	for key := range envMap {
		if value, exists := os.LookupEnv(key); exists {
			envVars = append(envVars, key+"="+value)
		}
	}
	sort.Strings(envVars)

	// 打印 .env 文件中定义的环境变量
	log.Println("Environment Variables from .env file:")
	for _, envVar := range envVars {
		log.Println(envVar)
	}

	// 保持程序运行
	for {
		time.Sleep(5 * time.Minute)
	}
}
