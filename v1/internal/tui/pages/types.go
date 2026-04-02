package pages

type CrawlState int

const (
	CrawlStopped CrawlState = iota
	CrawlRunning
	CrawlError
)
