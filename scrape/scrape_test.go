package scrape

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScrape(t *testing.T) {
	summary, md, err := Scrape(context.Background(), "https://www.bestbytes.com/referenzen", "main")
	require.NoError(t, err)
	fmt.Println(summary, md)
}
