package apis

import (
	"github.com/pixie-sh/logger-go/logger"

	"github.com/pixie-sh/core-go/pkg/models/response_models"
	"github.com/pixie-sh/core-go/pkg/models/serializer"
)

func processRaw[T any](log logger.Interface, blobResp []byte, err error, dest *response_models.Response[T], withValidations ...bool) error {
	if err != nil {
		if len(blobResp) > 0 {
			if errDeserialize := serializer.Deserialize(blobResp, dest, false); errDeserialize != nil {
				log.
					With("error", err).
					With("error_deserialize", errDeserialize).
					With("blob_response", blobResp).
					Error("processRaw issue deserializing body")

				return err
			}

			if dest.Error != nil {
				return dest.Error
			}
		}

		return err
	}

	if err := serializer.Deserialize(blobResp, dest, withValidations...); err != nil {
		log.
			With("error", err).
			With("blob_response", blobResp).
			Error("processRawResponse issue deserializing body")

		return err
	}

	if dest.Error != nil {
		return dest.Error
	}

	return nil
}

func processUnknownRaw[T any](log logger.Interface, blobResp []byte, err error, dest *T, withValidations ...bool) error {
	if err != nil {
		if len(blobResp) > 0 {
			if errDeserialize := serializer.Deserialize(blobResp, dest, false); errDeserialize != nil {
				log.
					With("error", err).
					With("error_deserialize", errDeserialize).
					With("blob_response", blobResp).
					Error("processRaw issue deserializing body")

				return err
			}
		}

		return err
	}

	if err := serializer.Deserialize(blobResp, dest, withValidations...); err != nil {
		log.
			With("error", err).
			With("blob_response", blobResp).
			Error("processRawResponse issue deserializing body")

		return err
	}

	return nil
}
