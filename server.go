package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	// "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/internal/generated"

	// "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/internal/generated"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	// "go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/mongo/options"
)

type uploadLayout struct {
	Title template.HTML
	Body  template.HTML
}

// UploadFileToBucket ...
func UploadFileToBucket(bucketName string, filepath string, body string, client *s3.Client) {
	fileContent := strings.NewReader(body)
	_, err := client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(filepath),
		Body:        fileContent,
		ACL:         "public-read",
		ContentType: to.Ptr("text/html"),
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
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(c.Param("region")))
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

func postFileToBucketFromTemplate(c echo.Context) error {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(c.Param("region")))
	if err != nil {
		fmt.Printf("Error %s\n", err)
	}

	// fileContentBuff := c.Request().Body
	// fileContent := BufferToString(fileContentBuff)
	// filePath := c.Param("filePath")
	collection := connectToServer()
	template_id := c.Param("template_id")
	var result TemplateID
	// objectId, _ := primitive.ObjectIDFromHex(template_id)
	err = collection.FindOne(context.TODO(), bson.M{"templateid": template_id}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return c.String(http.StatusBadRequest, "")
	}

	tmpl, err := template.ParseFiles("upload/template.html")
	if err != nil {
		panic(err)
	}
	var tpl bytes.Buffer
	tmpl.Execute(&tpl, &uploadLayout{
		Title: template.HTML(result.TemplateTitle),
		Body:  template.HTML(result.TemplateData),
	})
	fileContent := tpl.String()
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

func postFileToBlobFromTemplate(c echo.Context) error {
	// fileContentBuff := c.Request().Body
	// fileContent := BufferToString(fileContentBuff)
	// filePath := c.Param("filePath")
	collection := connectToServer()
	template_id := c.Param("template_id")
	var result TemplateID
	// objectId, _ := primitive.ObjectIDFromHex(template_id)
	err := collection.FindOne(context.TODO(), bson.M{"templateid": template_id}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return c.String(http.StatusBadRequest, "")
	}

	tmpl, err := template.ParseFiles("upload/template.html")
	if err != nil {
		panic(err)
	}
	var tpl bytes.Buffer
	tmpl.Execute(&tpl, &uploadLayout{
		Title: template.HTML(result.TemplateTitle),
		Body:  template.HTML(result.TemplateData),
	})
	fileContent := tpl.String()
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
	e.Static("/public", "public")
	e.GET("/aws/:bucket/:filePath/:region/:template_id", postFileToBucketFromTemplate)
	e.POST("/aws/:bucket/:filePath/:region", postFileToBucket)
	e.GET("/azure/:container/:filePath/:template_id", postFileToBlobFromTemplate)
	e.POST("/azure/:container/:filePath", postFileToBlob)
	e.GET("/template", indexTemplate)
	e.GET("/template/:id", loadTemplate)
	e.POST("/template/:id/update", updateTemplate)
	e.GET("/template/:id/rename/:name", renameTemplate)
	e.GET("/template/:name/add", addTemplate)
	e.GET("/template/:id/delete", removeTemplate)
	e.Start(":" + os.Getenv("SERVER_PORT"))
}
