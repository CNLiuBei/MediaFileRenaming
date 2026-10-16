package syncer

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultAPIBase = "https://tmdb.liubei.org/3"
const defaultQPS = 20

type Config struct {
	APIBase         string
	APIKey          string
	ReadAccessToken string
	QPS             float64
	Interval        time.Duration
	AutoStart       bool
	Language        string
}

func LoadConfig() (Config, error) {
	base := strings.TrimSpace(os.Getenv("TMDB_API_BASE_URL"))
	if base == "" {
		base = defaultAPIBase
	}
	parsed, err := url.Parse(base)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return Config{}, fmt.Errorf("TMDB_API_BASE_URL 必须是有效的 HTTP 或 HTTPS 地址")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	if !strings.HasSuffix(parsed.Path, "/3") {
		parsed.Path += "/3"
	}
	qps := float64(defaultQPS)
	if value := strings.TrimSpace(os.Getenv("TMDB_SYNC_QPS")); value != "" {
		qps, err = strconv.ParseFloat(value, 64)
		if err != nil || qps <= 0 || qps > 40 {
			return Config{}, fmt.Errorf("TMDB_SYNC_QPS 必须大于 0 且不高于 40")
		}
	}
	interval := 24 * time.Hour
	if value := strings.TrimSpace(os.Getenv("TMDB_SYNC_INTERVAL")); value != "" {
		interval, err = time.ParseDuration(value)
		if err != nil || interval < time.Hour {
			return Config{}, fmt.Errorf("TMDB_SYNC_INTERVAL 必须是至少 1 小时的时间间隔")
		}
	}
	autoStart := true
	if value := strings.TrimSpace(os.Getenv("TMDB_SYNC_AUTOSTART")); value != "" {
		autoStart, err = strconv.ParseBool(value)
		if err != nil {
			return Config{}, fmt.Errorf("TMDB_SYNC_AUTOSTART 必须是 true 或 false")
		}
	}
	return Config{
		APIBase:         strings.TrimRight(parsed.String(), "/"),
		APIKey:          strings.TrimSpace(os.Getenv("TMDB_API_KEY")),
		ReadAccessToken: strings.TrimSpace(os.Getenv("TMDB_READ_ACCESS_TOKEN")),
		QPS:             qps,
		Interval:        interval,
		AutoStart:       autoStart,
		Language:        "zh-CN",
	}, nil
}
