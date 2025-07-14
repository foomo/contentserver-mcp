package scrape

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScrape(t *testing.T) {
	md, err := Scrape(context.Background(), "https://www.bestbytes.com/", "main")
	require.NoError(t, err)
	fmt.Println(md)
}
