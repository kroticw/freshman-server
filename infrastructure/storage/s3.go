package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/aws/smithy-go/logging"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type S3Driver struct {
	bucket   string
	basePath string
	svc      *s3.Client
	tracer   trace.Tracer
}

func NewS3(
	region string,
	endpoint string,
	accessKeyID string,
	secretAccessKey string,
	forcePathStyle bool,
	bucket string,
	basePath string,
	tracer trace.Tracer,
	logger *logrus.Logger,
) *S3Driver {
	loggerS3 := logging.LoggerFunc(func(classification logging.Classification, format string, v ...interface{}) {
		if classification == logging.Warn {
			logger.WithField("category", "s3").Warnf(format, v)
		} else if classification == logging.Debug {
			logger.WithField("category", "s3").Debugf(format, v)
		} else {
			logger.WithField("category", "s3").Printf(format, v)
		}
	})

	client := s3.New(s3.Options{
		Credentials: credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
		EndpointOptions: s3.EndpointResolverOptions{
			Logger:        loggerS3,
			LogDeprecated: false,
			DisableHTTPS:  !strings.HasPrefix(endpoint, "https://"),
		},
		BaseEndpoint: aws.String(endpoint),
		Region:       region,
		Logger:       loggerS3,
		UsePathStyle: forcePathStyle,
		HTTPClient: &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	})

	return &S3Driver{
		bucket:   bucket,
		basePath: basePath,
		tracer:   tracer,
		svc:      client,
	}
}

func (s *S3Driver) Exists(ctx context.Context, filename string) (exists bool, err error) {
	ctx, span := s.tracer.Start(ctx, "Exists")
	defer span.End()
	filename = getFilePath(filename, s.basePath)
	span.SetAttributes(attribute.String("aws.s3.key", filename), attribute.String("aws.s3.bucket", s.bucket))
	_, err = s.svc.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &s.bucket,
		Key:    &filename,
	})
	if err != nil {
		var nf *types.NotFound
		if errors.As(err, &nf) {
			return false, nil
		}
		span.SetStatus(codes.Error, err.Error())
		return false, err
	}
	return true, nil
}

func (s *S3Driver) IsLinkedExists(ctx context.Context, filename string, sourceFilename string) (exists bool, err error) {
	sourceFilename = strings.Replace(path.Base(sourceFilename), path.Ext(sourceFilename), "", -1)
	return s.Exists(ctx, path.Join(sourceFilename, filename))
}

func (s *S3Driver) Upload(ctx context.Context, filename string, file []byte) error {
	ctx, span := s.tracer.Start(ctx, "Upload")
	defer span.End()
	span.SetAttributes(attribute.String("aws.s3.key", filename), attribute.String("aws.s3.bucket", s.bucket))
	exists, err := s.Exists(ctx, filename)
	if err != nil {
		return err
	}
	if exists {
		return os.ErrExist
	}
	reader := bytes.NewReader(file)
	defer reader.Reset([]byte("")) //memory leak fixer
	filename = getFilePath(filename, s.basePath)
	_, err = s.svc.PutObject(ctx, &s3.PutObjectInput{
		Bucket:             &s.bucket,
		Key:                &filename,
		ACL:                types.ObjectCannedACLPrivate,
		Body:               reader,
		ContentDisposition: aws.String("attachment"),
	})
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}

func (s *S3Driver) UploadLinked(ctx context.Context, filename string, sourceFilename string, file []byte) error {
	fullFileName := getFullFileName(filename, sourceFilename)
	return s.Upload(ctx, getFilePath(fullFileName, s.basePath), file)
}

func (s *S3Driver) Get(ctx context.Context, filename string) ([]byte, error) {
	ctx, span := s.tracer.Start(ctx, "Get")
	defer span.End()
	filename = getFilePath(filename, s.basePath)
	span.SetAttributes(attribute.String("aws.s3.key", filename), attribute.String("aws.s3.bucket", s.bucket))
	results, err := s.svc.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filename),
	})
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		var opErr *smithy.OperationError
		if errors.As(err, &opErr) {
			var nf *types.NotFound
			if errors.As(err, &nf) {
				return nil, os.ErrNotExist
			}
		}
		return nil, err
	}
	defer results.Body.Close()

	return io.ReadAll(results.Body)
}

func (s *S3Driver) GetLinked(ctx context.Context, filename string, sourceFilename string) ([]byte, error) {
	return s.Get(ctx, getFullFileName(filename, sourceFilename))
}

func (s *S3Driver) Delete(ctx context.Context, filename string) error {
	ctx, span := s.tracer.Start(ctx, "Delete")
	defer span.End()

	// Сначала удаляем слинкованные файлы
	linkedFilesDeleteCtx, linkedFilesDeleteSpan := s.tracer.Start(ctx, "Delete linked files")
	sourceFilename := getFilePath(strings.Replace(path.Base(filename), path.Ext(filename), "", -1), s.basePath)
	_, err := s.svc.DeleteObject(linkedFilesDeleteCtx, &s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &sourceFilename,
	})
	if err != nil {
		linkedFilesDeleteSpan.SetStatus(codes.Error, err.Error())
		linkedFilesDeleteSpan.End()
		return err
	}
	linkedFilesDeleteSpan.End()
	// Удаляем оригинал
	sourceFileDeleteCtx, sourceFileDeleteSpan := s.tracer.Start(ctx, "Delete source file")
	filename = getFilePath(filename, s.basePath)
	_, err = s.svc.DeleteObject(sourceFileDeleteCtx, &s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &filename,
	})
	if err != nil {
		sourceFileDeleteSpan.SetStatus(codes.Error, err.Error())
		sourceFileDeleteSpan.End()
		return err
	}

	return nil
}

func (s *S3Driver) DeleteCache(ctx context.Context, filename string, sourceFilename string) error {
	sourceFileDeleteCtx, sourceFileDeleteSpan := s.tracer.Start(ctx, "Delete cache file")
	fillPath := getFilePath(getFullFileName(filename, sourceFilename), s.basePath)
	_, err := s.svc.DeleteObject(sourceFileDeleteCtx, &s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &fillPath,
	})
	if err != nil {
		sourceFileDeleteSpan.SetStatus(codes.Error, err.Error())
		sourceFileDeleteSpan.End()
		return err
	}

	return nil
}

func (s *S3Driver) GetSpaceUsage(_ context.Context) (usage int64, err error) {
	return -1, nil
}
