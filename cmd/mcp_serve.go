package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/yuhua2000/sheetpilot/internal/mcp"
)

var (
	transport string
	addr      string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start MCP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))

		srv, err := mcp.NewServer()
		if err != nil {
			return fmt.Errorf("create mcp server: %w", err)
		}

		switch transport {
		case "stdio":
			return srv.ServeStdio()
		case "sse":
			return srv.ServeSSE(addr)
		default:
			return fmt.Errorf("unsupported transport: %s", transport)
		}
	},
}

func init() {
	serveCmd.Flags().StringVarP(&transport, "transport", "t", "stdio", "Transport mode: stdio or sse")
	serveCmd.Flags().StringVarP(&addr, "addr", "a", ":8080", "Listen address for SSE mode")
}
