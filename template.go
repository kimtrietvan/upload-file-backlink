package main

import (
	"bytes"
	"context"
	"html/template"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TemplateID struct {
	TemplateID    string `bson:"templateid"`
	TemplateName  string `bson:templatename`
	TemplateData  string `bson:templatedata`
	TemplateTitle string `bson:templatetitle`
}

type templateLayout struct {
	TemplateIDs []TemplateID
	Title       string
}

type templateLoad struct {
	TemplateIDs  []TemplateID
	Title        string
	CurrentID    string
	CurrentData  string
	CurrentName  string
	CurrentTitle string
}

func updateTemplate(c echo.Context) error {
	template_id := c.Param("id")

	collection := connectToServer()

	// objectId, _ := primitive.ObjectIDFromHex(template_id)

	var result TemplateID
	err := collection.FindOne(context.TODO(), bson.M{"templateid": template_id}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return c.String(http.StatusBadRequest, "")
	}
	// json_map := make(map[string]interface{})
	// err = json.NewDecoder(c.Request().Body).Decode(&json_map)
	// if err != nil {
	// 	panic(err)
	// }
	update := bson.D{{"$set", bson.M{"templateTitle": c.FormValue("templateTitle"), "templateData": c.FormValue("templateData")}}}
	_, err = collection.UpdateOne(context.TODO(), bson.M{"templateid": template_id}, update)
	if err != nil {
		panic(err)
	}
	// fmt.Printf("updateResult: %v\n", updateResult)
	// return c.Response().Header().Add("Location", "123")
	return c.Redirect(301, "/template/"+template_id)

}

func loadTemplate(c echo.Context) error {
	template_id := c.Param("id")
	collection := connectToServer()
	cursor, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		panic(err)
	}

	var dataPassing templateLoad
	dataPassing.Title = "test"
	for cursor.Next(context.Background()) {
		var result TemplateID
		err := cursor.Decode(&result)
		if err != nil {
			panic(err)
		}

		dataPassing.TemplateIDs = append(dataPassing.TemplateIDs, result)

	}
	var result TemplateID
	// objectId, _ := primitive.ObjectIDFromHex(template_id)
	err = collection.FindOne(context.TODO(), bson.M{"templateid": template_id}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return c.String(http.StatusBadRequest, "")
	}

	dataPassing.CurrentID = result.TemplateID
	dataPassing.CurrentData = result.TemplateData
	dataPassing.CurrentName = result.TemplateName
	dataPassing.CurrentTitle = result.TemplateTitle
	dataPassing.Title = result.TemplateName + "edit"

	tmpl, err := template.ParseFiles("frontend/load.html")
	if err != nil {
		panic(err)
	}
	var tpl bytes.Buffer
	tmpl.Execute(&tpl, dataPassing)
	// fmt.Printf("%s\n", dataPassing.TemplateIDs[0].TemplateID)
	return c.HTML(http.StatusOK, tpl.String())
}

func indexTemplate(c echo.Context) error {
	collection := connectToServer()
	cursor, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		panic(err)
	}

	var dataPassing templateLayout
	dataPassing.Title = "test"
	for cursor.Next(context.Background()) {
		var result TemplateID
		err := cursor.Decode(&result)
		if err != nil {
			panic(err)
		}

		dataPassing.TemplateIDs = append(dataPassing.TemplateIDs, result)

	}
	tmpl, err := template.ParseFiles("frontend/index.html")
	if err != nil {
		panic(err)
	}
	var tpl bytes.Buffer
	tmpl.Execute(&tpl, dataPassing)
	// fmt.Printf("%s\n", dataPassing.TemplateIDs[0].TemplateID)
	return c.HTML(http.StatusOK, tpl.String())

}

func renameTemplate(c echo.Context) error {
	template_id := c.Param("id")
	name := c.Param("name")
	collection := connectToServer()

	// objectId, _ := primitive.ObjectIDFromHex(template_id)

	var result TemplateID
	err := collection.FindOne(context.TODO(), bson.M{"templateid": template_id}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return c.String(http.StatusBadRequest, "")
	}
	// json_map := make(map[string]interface{})
	// err = json.NewDecoder(c.Request().Body).Decode(&json_map)
	// if err != nil {
	// 	panic(err)
	// }
	update := bson.D{{"$set", bson.M{"templateName": name}}}
	_, err = collection.UpdateOne(context.TODO(), bson.M{"templateid": template_id}, update)
	if err != nil {
		panic(err)
	}
	return c.Redirect(301, "/template/"+template_id)
}

func addTemplate(c echo.Context) error {
	name := c.Param("name")
	newTemplate := TemplateID{TemplateName: name, TemplateData: "", TemplateTitle: "", TemplateID: primitive.NewObjectID().Hex()}
	collection := connectToServer()
	_, _ = collection.InsertOne(context.TODO(), newTemplate)
	return c.String(http.StatusOK, "")
}

func removeTemplate(c echo.Context) error {
	template_id := c.Param("id")
	collection := connectToServer()
	_, err := collection.DeleteOne(context.Background(), bson.M{"templateid": template_id})
	if err != nil {
		panic(err)
	}
	return c.String(http.StatusOK, "")
}
