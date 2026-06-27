package main

import (
	"context"
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
	s3Endpoint  = os.Getenv("S3_ENDPOINT") // e.g., http://localhost:8333
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

		// Use strings.NewReader which implements io.ReadSeeker
		_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
			Body:   strings.NewReader(content), // This is seekable
		})
		if err != nil {
			http.Error(w, "failed to put object: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	fmt.Fprintln(w, "Generated 100 objects in bucket "+bucketName)
}

func showHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := s3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{Bucket: aws.String(bucketName)})
	if err != nil {
		http.Error(w, "failed to list objects: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl := `<html><body><h1>Objects in {{.Bucket}}</h1><ul>{{range .Objects}}<li>{{.}}</li>{{end}}</ul></body></html>`
	t := template.Must(template.New("list").Parse(tmpl))
	var names []string
	for _, o := range resp.Contents {
		names = append(names, *o.Key)
	}
	data := struct {
		Bucket  string
		Objects []string
	}{Bucket: bucketName, Objects: names}
	t.Execute(w, data)
}

func main() {
	initS3()
	http.HandleFunc("/generate", generateHandler)
	http.HandleFunc("/show", showHandler)
	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
