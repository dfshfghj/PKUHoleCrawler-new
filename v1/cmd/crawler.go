package main

import (
	"github.com/spf13/cobra"
)

func newCrawlerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "crawler",
		Short: "Run the crawler",
		Long:  `启动 PKU Hole 爬虫，支持一次性抓取、无限抓取和监控模式。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDaemon()
		},
	}

	cmd.Flags().BoolVarP(&daemonMode, "daemon", "D", false, "enable daemon (background) mode")
	cmd.Flags().IntVar(&startPage, "start", 1, "starting page number")
	cmd.Flags().IntVar(&pages, "pages", 0, "pages to crawl per round (0 = infinite)")
	cmd.Flags().IntVar(&interval, "interval", 1, "seconds between each page crawl")
	cmd.Flags().IntVar(&roundIntval, "round-interval", 60, "seconds between rounds in monitor mode")
	cmd.Flags().BoolVar(&resume, "resume", false, "resume from last crawled page")
	cmd.Flags().IntVarP(&monitorPages, "monitor", "k", 0, "monitor mode: loop crawl first N pages")

	return cmd
}
