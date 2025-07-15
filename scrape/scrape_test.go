package scrape

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScrape(t *testing.T) {
	summary, md, err := Scrape(context.Background(), http.DefaultClient, "https://globus-b.prod.mzg.bestbytes.net/damen", "main")
	require.NoError(t, err)
	fmt.Println(summary, md)
}
