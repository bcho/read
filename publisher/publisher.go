package publisher

import (
	"fmt"
	"strings"

	"github.com/bcho/timespan"
)

type Publisher interface {
	Publish(timespan.Span, []string) (string, error)
}

func title(span timespan.Span, articles []string) string {
	return fmt.Sprintf(
		"Reading note on %s ~ %s",
		span.Start().Format("2006-01-02"),
		span.End().Format("2006-01-02"),
	)
}

func content(span timespan.Span, articles []string) string {
	return strings.Join(articles, "\n\n")
}
