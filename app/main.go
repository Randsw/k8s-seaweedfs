package main

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	s3Endpoint  = os.Getenv("S3_ENDPOINT")
	s3APIKey    = os.Getenv("S3_API_KEY")
	s3SecretKey = os.Getenv("S3_SECRET_KEY")
	bucketName  = "data-bucket"
	s3Client    *s3.Client
)

func initS3() {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(s3APIKey, s3SecretKey, "")),
	)
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		if s3Endpoint != "" {
			o.BaseEndpoint = aws.String(s3Endpoint)
		}
		o.UsePathStyle = true
	})
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("object-%03d.txt", i)
		content := fmt.Sprintf("content of %s", key)

		// Calculate MD5 manually
		hash := md5.Sum([]byte(content))
		contentMD5 := base64.StdEncoding.EncodeToString(hash[:])

		_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket:     aws.String(bucketName),
			Key:        aws.String(key),
			Body:       strings.NewReader(content),
			ContentMD5: aws.String(contentMD5),
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to put object %s: %v", key, err), http.StatusInternalServerError)
			return
		}
	}
	fmt.Fprintln(w, "Generated 100 objects in bucket "+bucketName)
}

func showHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := s3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		http.Error(w, "failed to list objects: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := `<html><body><h1>Objects in {{.Bucket}}</h1><ul>{{range .Objects}}<li>{{.}}</li>{{end}}</ul></body></html>`
	t := template.Must(template.New("list").Parse(tmpl))

	var names []string
	for _, o := range resp.Contents {
		if o.Key != nil {
			names = append(names, *o.Key)
		}
	}

	data := struct {
		Bucket  string
		Objects []string
	}{
		Bucket:  bucketName,
		Objects: names,
	}

	t.Execute(w, data)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Simple health check that doesn't require additional permissions
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

func bucketCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Check if bucket exists and is accessible
	_, err := s3Client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Bucket %s not accessible: %v", bucketName, err), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Bucket %s is accessible\n", bucketName)
}

func main() {
	initS3()

	// Removed ListBuckets call since user only has access to data-bucket

	http.HandleFunc("/generate", generateHandler)
	http.HandleFunc("/show", showHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/check-bucket", bucketCheckHandler)

	log.Println("Server listening on :8080")
	log.Printf("Using bucket: %s", bucketName)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
