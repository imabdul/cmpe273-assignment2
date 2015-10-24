package main

import (
    "net/http"
    "github.com/julienschmidt/httprouter"
    "fmt"
    "log"
    "gopkg.in/mgo.v2"
    "./controllers"
)

//connection to mongoDB in Mongolab
func getMongoSession() *mgo.Session {

    session, err := mgo.Dial("mongodb://imabdul:imabdul@ds043694.mongolab.com:43694/cmpe273")
    if err != nil {
        panic(err)
    }

    return session
}

func main() {

    //router being instantiated
    router := httprouter.New()
    consumerController := controllers.NewConsumerController(getMongoSession())

    //routing GET, POST, PUT, DELETE
    router.GET("/locations/:id", consumerController.GetConsumer)
    router.POST("/locations", consumerController.CreateConsumer)
    router.PUT("/locations/:id", consumerController.UpdateConsumer)
    router.DELETE("/locations/:id", consumerController.RemoveConsumer)

    //server kick off
    fmt.Println("Server listening on 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

