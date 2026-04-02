package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"treehole/internal/crawler"
	"treehole/internal/db"
	"treehole/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func init() {
	logFile, err := os.OpenFile("crawler.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		log.SetOutput(logFile)
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	}
}

var (
	dbPath       string
	daemonMode   bool
	startPage    int
	pages        int
	interval     int
	roundIntval  int
	resume       bool
	monitorPages int
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "treehole",
		Short: "PKU Hole Crawler & API Server",
		Long:  `PKU Hole 爬虫、TUI 交互式界面和 API 服务器的统一工具。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTUI()
		},
	}

	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "./test-data.db", "database file path")

	rootCmd.AddCommand(newServerCmd())
	rootCmd.AddCommand(newCrawlerCmd())

	return rootCmd
}

func initDB() (*db.Database, func(), error) {
	database, err := db.NewDatabase(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("初始化数据库失败: %w", err)
	}

	cleanup := func() {
		database.Checkpoint()
		database.Close()
	}

	return database, cleanup, nil
}

func runTUI() error {
	database, cleanup, err := initDB()
	if err != nil {
		return err
	}
	defer cleanup()

	client, cfg, err := tui.InitClientForTUI()
	if err != nil {
		return fmt.Errorf("初始化客户端失败: %w", err)
	}

	model := tui.NewModel(database, client, cfg)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI运行错误: %w", err)
	}

	return nil
}

func runDaemon() error {
	database, cleanup, err := initDB()
	if err != nil {
		return err
	}
	defer cleanup()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	client, _, err := tui.InitClientForTUI()
	if err != nil {
		return fmt.Errorf("初始化客户端失败: %w", err)
	}

	if resume {
		count, _ := database.GetPostCount()
		if count > 0 {
			startPage = (count / 100) + 1
			log.Printf("[Daemon] 断点续爬模式: 从第 %d 页开始 (已有 %d 条帖子)", startPage, count)
		}
	}

	monitorMode := monitorPages > 0
	if monitorMode {
		log.Printf("[Daemon] 监控模式启动: 循环抓取前 %d 页, 每轮间隔 %ds", monitorPages, roundIntval)
	} else if pages == 0 {
		log.Printf("[Daemon] 无限抓取模式启动: 从第 %d 页开始, 每页间隔 %ds", startPage, interval)
	} else {
		log.Printf("[Daemon] 一次性抓取模式启动: 从第 %d 页开始, 抓取 %d 页", startPage, pages)
	}

	page := startPage
	round := 0
	totalPosts := 0
	totalComments := 0

	for {
		select {
		case <-sigCh:
			log.Printf("[Daemon] 收到退出信号，正在优雅停止...")
			return nil
		default:
		}

		if monitorMode {
			page = 1
		}

		round++
		log.Printf("[Daemon] 开始第 %d 轮抓取", round)

		crawled := 0
		limit := pages
		if monitorMode {
			limit = monitorPages
		}

		for {
			select {
			case <-sigCh:
				log.Printf("[Daemon] 收到退出信号，正在优雅停止...")
				return nil
			default:
			}

			if limit > 0 && crawled >= limit {
				break
			}

			result, err := crawler.FetchAndSave(client, database, page)
			if err != nil {
				log.Printf("[Daemon] 第 %d 页抓取失败: %v", page, err)
				time.Sleep(time.Duration(interval) * time.Second)
				page++
				continue
			}

			totalPosts += result.PostCount
			totalComments += result.CommentCount
			crawled++

			pc, _ := database.GetPostCount()
			cc, _ := database.GetCommentCount()
			log.Printf("[Daemon] 第 %d 页完成: +%d帖子 +%d评论 | 总计: %d帖子 %d评论",
				page, result.PostCount, result.CommentCount, pc, cc)

			page++

			if limit > 0 && crawled >= limit {
				break
			}

			time.Sleep(time.Duration(interval) * time.Second)
		}

		if !monitorMode && pages > 0 {
			log.Printf("[Daemon] 抓取完成! 共处理 %d 页, +%d帖子 +%d评论", crawled, totalPosts, totalComments)
			return nil
		}

		if monitorMode {
			log.Printf("[Daemon] 第 %d 轮完成, 等待 %ds 后开始下一轮...", round, roundIntval)
		}

		select {
		case <-sigCh:
			log.Printf("[Daemon] 收到退出信号，正在优雅停止...")
			return nil
		case <-time.After(time.Duration(roundIntval) * time.Second):
		}
	}
}
