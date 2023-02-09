package torrentfile

import "context"

type TFInt interface {
	DownloadToFile(ctx context.Context, path string) error
	VerifyFiles(ctx context.Context, path string) error
}
