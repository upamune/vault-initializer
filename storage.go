package main

import (
	"bytes"
	"context"
	"io/ioutil"

	"cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
)

type Storage interface {
	Get(key string) ([]byte, error)
	Put(key string, body []byte) error
}

type S3 struct {
	BucketName string
	session    *session.Session
}

func NewS3(sess *session.Session, bucketName string) *S3 {
	return &S3{
		BucketName: bucketName,
		session:    sess,
	}
}

func (s *S3) Get(key string) ([]byte, error) {
	svc := s3.New(s.session)

	obj, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get an object from S3: key=%s", key)
	}
	defer obj.Body.Close()

	return ioutil.ReadAll(obj.Body) // FIXME(upamune)
}

func (s *S3) Put(key string, body []byte) error {
	svc := s3.New(s.session)

	_, err := svc.PutObject(&s3.PutObjectInput{
		Body:   bytes.NewReader(body),
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(key),
	})
	return errors.Wrapf(err, "failed to put an object to S3: key=%s", key)
}

// TODO
type GCS struct {
	client     *storage.Client
	bucketName string
}

func NewGCS(client *storage.Client, bucketName string) *GCS {
	return &GCS{
		client:     client,
		bucketName: bucketName,
	}
}

func (s *GCS) Get(key string) ([]byte, error) {
	bucket := s.client.Bucket(s.bucketName)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reader, err := bucket.Object(key).NewReader(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get an object from GCS: key=%s", key)
	}
	defer reader.Close()

	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (s *GCS) Put(key string, body []byte) error {
	bucket := s.client.Bucket(s.bucketName)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	writer := bucket.Object(key).NewWriter(ctx)
	defer writer.Close()

	if _, err := writer.Write(body); err != nil {
		return errors.Wrapf(err, "failed to put an object to GCS: kye=%s", key)
	}

	return nil
}
