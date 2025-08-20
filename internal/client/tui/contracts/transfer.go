package contracts

import (
	"context"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

type BinaryTransferCapable interface {
	UploadBinaryData(ctx context.Context, data *model.BinaryData, filePath string, progress chan<- int64) error
	DownloadBinaryData(ctx context.Context, dataID, destPath string, progress chan<- int64) error
}
