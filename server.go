package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"

	// "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/internal/generated"

	// "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/internal/generated"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

// UploadFileToBucket ...
func UploadFileToBucket(bucketName string, filepath string, body string, client *s3.Client) {
	fileContent := strings.NewReader(body)
	_, err := client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filepath),
		Body:   fileContent,
		ACL:    "public-read",
		// Metadata:,
	})
	if err != nil {
		panic(err)
	}
}

// func PublicFile(bucket string, filepath string, client *s3.Client) {
// 	client.PutObjectAcl(context.TODO(), &s3.PutObjectAclInput{
// 		Bucket: aws.String(bucket),
// 		ACL:    "public-read",
// 	})
// }

func postFileToBucket(c echo.Context) error {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		fmt.Printf("Error %s\n", err)
	}

	fileContentBuff := c.Request().Body
	fileContent := BufferToString(fileContentBuff)
	// filePath := c.Param("filePath")
	filePathBytes, _ := base64.StdEncoding.DecodeString(c.Param("filePath"))
	filePath := bytes.NewBuffer(filePathBytes).String()
	bucketName := c.Param("bucket")

	client := s3.NewFromConfig(cfg)
	_, err = client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		panic(err)
	}

	UploadFileToBucket(bucketName, filePath, fileContent, client)
	return c.String(http.StatusOK, "")
}

// BufferToString ...
func BufferToString(buff io.ReadCloser) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(buff)
	respBytes := buf.String()

	respString := string(respBytes)
	return respString
}

func postFileToBlob(c echo.Context) error {
	fileContentBuff := c.Request().Body
	fileContent := BufferToString(fileContentBuff)
	// filePath := c.Param("filePath")
	filePathBytes, _ := base64.StdEncoding.DecodeString(c.Param("filePath"))
	filePath := bytes.NewBuffer(filePathBytes).String()
	containerName := c.Param("container")

	url := "https://" + os.Getenv("AZURE_STORAGE_ACCOUNT") + ".blob.core.windows.net/"
	ctx := context.Background()
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		panic(err)
	}
	client, err := azblob.NewClient(url, credential, nil)
	if err != nil {
		panic(err)
	}

	fileContentBytes := []byte(fileContent)
	_, err = client.UploadBuffer(ctx, containerName, filePath, fileContentBytes, &azblob.UploadBufferOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType: to.Ptr("text/html"),
		},
	})
	if err != nil {
		panic(err)
	}
	// defer url

	// u := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", os.Getenv("AZURE_STORAGE_ACCOUNT"), containerName, filePath)
	// blobClient, err := azblob.NewClient(u, credential, nil)
	// blobClient.ServiceClient().GetProperties(ctx, nil)

	// fmt.Printf("blobClient: %v\n", blobClient)
	// paper, err := client.ServiceClient().GetProperties(ctx, &service.GetPropertiesOptions{})

	// pipeline := client.(context.Background(), nil, nil)
	// container = client.ServiceClient()

	return c.String(http.StatusOK, "")
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("Error %s\n", err)
	}
	e := echo.New()
	e.POST("/aws/:bucket/:filePath", postFileToBucket)
	e.POST("/azure/:container/:filePath", postFileToBlob)
	e.Start(":" + os.Getenv("SERVER_PORT"))
}
