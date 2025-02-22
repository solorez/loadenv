// Package loadenv 提供环境变量加载和热重载功能
package loadenv

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
)

var (
	once     sync.Once
	watcher  *fsnotify.Watcher
	closeCh  chan struct{}
	filePath string
	mu       sync.RWMutex
	logger   *log.Logger
)

// Config 配置参数
type Config struct {
	FilePath    string        // 环境文件路径
	HotReload   bool          // 是否启用热重载
	Logger      *log.Logger   // 自定义日志记录器
	ReloadDelay time.Duration // 重载延迟（防抖）
}

// InitEnv 初始化环境变量加载
func InitEnv(cfg Config) error {
	var initErr error
	once.Do(func() {
		// 设置默认值
		if cfg.FilePath == "" {
			cfg.FilePath = ".env"
		}
		if cfg.ReloadDelay == 0 {
			cfg.ReloadDelay = 2 * time.Second
		}
		if cfg.Logger == nil {
			cfg.Logger = log.New(os.Stdout, "[ENV] ", log.LstdFlags)
		}

		logger = cfg.Logger
		filePath = cfg.FilePath

		// 首次加载
		if err := load(); err != nil {
			initErr = err
			return
		}

		// 初始化监听器
		if cfg.HotReload {
			if err := initWatcher(cfg.FilePath); err != nil {
				initErr = err
				return
			}
			go watchEvents(cfg.ReloadDelay)
		}
	})
	return initErr
}

// load 实际加载环境变量的方法
func load() error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}

	mu.Lock()
	defer mu.Unlock()

	logger.Printf("Loading environment from: %s", absPath)
	return godotenv.Load(absPath)
}

// initWatcher 初始化文件监听
func initWatcher(path string) error {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	if err := watcher.Add(absPath); err != nil {
		return err
	}

	closeCh = make(chan struct{})
	logger.Printf("Starting hot reload watcher for: %s", absPath)
	return nil
}

func watchEvents(delay time.Duration) {
	defer watcher.Close()

	var (
		timer     *time.Timer
		lastEvent time.Time
	)

	// 读取当前 .env 文件内容
	oldEnvContent, err := os.ReadFile(".env")
	if err != nil {
		logger.Printf("Failed to read .env file: %v", err)
		return
	}
	oldEnv := parseEnvFile(string(oldEnvContent))

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// 防抖处理
			now := time.Now()
			if now.Sub(lastEvent) < delay {
				continue
			}

			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				if timer != nil {
					timer.Stop()
				}

				timer = time.AfterFunc(delay, func() {
					if err := load(); err != nil {
						logger.Printf("Reload failed: %v", err)
					} else {
						logger.Printf("Successfully reloaded environment file")

						// 读取新的 .env 文件内容
						newEnvContent, err := os.ReadFile(".env")
						if err != nil {
							logger.Printf("Failed to read updated .env file: %v", err)
							return
						}
						newEnv := parseEnvFile(string(newEnvContent))

						// 比较并输出变化的环境变量
						for key, newValue := range newEnv {
							oldValue, exists := oldEnv[key]
							if !exists {
								logger.Printf("New environment variable: %s = %s", key, newValue)
							} else if oldValue != newValue {
								logger.Printf("Environment variable changed: %s = %s (old value: %s)", key, newValue, oldValue)
							}
						}

						for key := range oldEnv {
							if _, exists := newEnv[key]; !exists {
								logger.Printf("Environment variable removed: %s", key)
							}
						}

						// 更新 oldEnv 为新的环境变量
						oldEnv = newEnv
					}
				})

				lastEvent = now
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			logger.Printf("Watcher error: %v", err)

		case <-closeCh:
			return
		}
	}
}

// 解析 .env 文件内容
func parseEnvFile(content string) map[string]string {
	env := make(map[string]string)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			env[key] = value
		}
	}
	return env
}

// Close 停止热重载监听
func Close() {
	if closeCh != nil {
		close(closeCh)
	}
}
