package controllers

import (
    "fmt"
    "net/http"
    "encoding/json"
    "strings"
    "errors"
    "log"
    "io"
    "io/ioutil"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "github.com/julienschmidt/httprouter"
    "../models"
)

//To control the Consumer resource
type ConsumerController struct{
    session *mgo.Session
}

//Session instantiation
func NewConsumerController(s *mgo.Session) *ConsumerController {
    return &ConsumerController{s}
}

//to fetch the coordinates of consumer with the help Google Maps Api by passing posted address
func fetchCoordinates(Consumer *models.Consumer) error {
    mapsUrl := "http://maps.google.com/maps/api/geocode/json?address=" + Consumer.Address + ", " + Consumer.City + ", " + Consumer.State
    mapsUrl = strings.Replace(mapsUrl, " ", "+", -1)

    res, err := http.Get(mapsUrl)
    if err != nil {
        return err
    }

    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        return err
    }
    defer res.Body.Close()

    var contents map[string]interface{}
    err = json.Unmarshal(body, &contents)
    if err != nil {
        return err
    }

    if !strings.EqualFold(contents["status"].(string), "OK") {
        return errors.New("Coordinates unavailable")
    }

    results := contents["results"].([]interface{})
    location := results[0].(map[string]interface{})["geometry"].(map[string]interface{})["location"]

    Consumer.Coordinate.Lat = location.(map[string]interface{})["lat"].(float64)
    Consumer.Coordinate.Lng = location.(map[string]interface{})["lng"].(float64)

    if err != nil {
        return err
    }

    return nil
}

//Consumer object creation in Consumer resource
func (cc ConsumerController) CreateConsumer(rw http.ResponseWriter, req *http.Request, _ httprouter.Params) {
    Consumer := models.Consumer{}
    json.NewDecoder(req.Body).Decode(&Consumer)
    Consumer.Id = bson.NewObjectId()
    err := fetchCoordinates(&Consumer)
    if err != nil {
		log.Println(err)
	}

    //Data persistence in MongoDB
    conn := cc.session.DB("cmpe273").C("imabdul")
    err = conn.Insert(Consumer)
    if err != nil {
        log.Println(err)
    }

    //response generation
    consumerJson, _ := json.Marshal(Consumer)
    if err != nil {
        rw.Header().Set("Content-Type", "plain/text")
        rw.WriteHeader(400)
        fmt.Fprintf(rw, "%s\n", err)
    } else {
        rw.Header().Set("Content-Type", "application/json")
        rw.WriteHeader(201)
        fmt.Fprintf(rw, "%s\n", consumerJson)
    }
}

func (cc ConsumerController) GetConsumer(rw http.ResponseWriter, _ *http.Request, param httprouter.Params) {
    consumer, err := fetchConsumerById(cc, param.ByName("id"))
    if err != nil {
		log.Println(err)
	}

    //response generation
    consumerJson, _ := json.Marshal(consumer)
    if err != nil {
        rw.Header().Set("Content-Type", "plain/text")
        rw.WriteHeader(400)
        fmt.Fprintf(rw, "%s\n", err)
    } else {
        rw.Header().Set("Content-Type", "application/json")
        rw.WriteHeader(200)
        fmt.Fprintf(rw, "%s\n", consumerJson)
    }
}

func (cc ConsumerController) UpdateConsumer(rw http.ResponseWriter, req *http.Request, param httprouter.Params) {
    updatedUsr, err := updateConsumerLocation(cc, param.ByName("id"), req.Body)
    if err != nil {
		log.Println(err)
	}

    //response generation
    consumerJson, _ := json.Marshal(updatedUsr)
    if err != nil {
        rw.Header().Set("Content-Type", "plain/text")
        rw.WriteHeader(400)
        fmt.Fprintf(rw, "%s\n", err)
    } else {
        rw.Header().Set("Content-Type", "application/json")
        rw.WriteHeader(201)
        fmt.Fprintf(rw, "%s\n", consumerJson)
    }
}

func (cc ConsumerController) RemoveConsumer(rw http.ResponseWriter, _ *http.Request, param httprouter.Params) {

    //consumer's existence check in db
    consumer, err := fetchConsumerById(cc, param.ByName("id"))
    if err != nil {
		log.Println(err)
        log.Println(consumer)
	}

    //consumer removed from the collection
    objId := bson.ObjectIdHex(param.ByName("id"))
    conn := cc.session.DB("cmpe273").C("imabdul")
    err = conn.Remove(bson.M{"id": objId})
    if err != nil {
        log.Println(err)
    }
    rw.Header().Set("Content-Type", "plain/text")
    if err != nil {
        rw.WriteHeader(400)
        fmt.Fprintf(rw, "%s\n", err)
    } else {
        rw.WriteHeader(200)
        fmt.Fprintf(rw, "Consumer ID=%s has been deleted", param.ByName("id"))
    }
}

func fetchConsumerById(cc ConsumerController, id string) (models.Consumer, error) {

    // Object ID existence check in Mongodb
    if !bson.IsObjectIdHex(id) {
        return models.Consumer{}, errors.New("Invalid Consumer ID")
    }
    objId := bson.ObjectIdHex(id)
    Consumer := models.Consumer{}
    conn := cc.session.DB("cmpe273").C("imabdul")
    err := conn.Find(bson.M{"id": objId}).One(&Consumer)
    if err != nil {
        return models.Consumer{}, errors.New("This Consumer Id doesn't exists")
    }

    return Consumer, nil
}

func updateConsumerLocation(cc ConsumerController, id string, contents io.Reader) (models.Consumer, error) {

    //consumer's existence check in db
    consumer, err := fetchConsumerById(cc, id)
    if err != nil {
        return models.Consumer{}, err
    }
    updConsumer := models.Consumer{}
    updConsumer.Id = consumer.Id
    updConsumer.Name = consumer.Name
    json.NewDecoder(contents).Decode(&updConsumer)

    //update and append coordinates
    err = fetchCoordinates(&updConsumer)
    if err != nil {
        return models.Consumer{}, err
    }

    //db connection and updating consumer
    objId := bson.ObjectIdHex(id)
    conn := cc.session.DB("cmpe273").C("imabdul")
    err = conn.Update(bson.M{"id": objId}, updConsumer)
    if err != nil {
        log.Println(err)
        return models.Consumer{}, errors.New("Given id is invalid")
    }
    return updConsumer, nil
}
