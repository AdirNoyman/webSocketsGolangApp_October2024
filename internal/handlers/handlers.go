package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
)

// Create a channel that will handle (only) WsJsonPayload structs
var WebSocketChannel = make(chan WsJsonPayload)

// Create a map that is a dictionary pairs of WsConnection as a key and username as a value. This will use as a pool of all the current websockets for the connected users
var ConnectedClients = make(map[WebSocketConnection]string)

var views = jet.NewSet(
	// Load the html template from the folder
	jet.NewOSFileSystemLoader("./html"),
	// Use hot reload (development only!)
	jet.InDevelopmentMode(),
)

var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func Home(w http.ResponseWriter, r *http.Request) {

	err := renderPage(w, "home.jet", nil)

	if err != nil {

		log.Println("There was an error 😫 => ", err)

	}
}

type WebSocketConnection struct {
	*websocket.Conn
}

// WsJsonResponse defines the response sent back from the server to the client(user)
type WsJsonResponse struct {
	Action         string   `json:"action"`
	Message        string   `json:"message"`
	MessageType    string   `json:"Message_type"`
	ConnectedUsers []string `json:"connected_users"`
}

// WsJsonPayload defines the message sent by the client(user), to the websocket
type WsJsonPayload struct {
	Action     string              `json:"action"`
	Username   string              `json:"username"`
	Message    string              `json:"message"`
	Connection WebSocketConnection `json:"-"`
}

// WebSocketEndPoint Upgrades Http connection to a Web Socket connection
func WebSocketEndPoint(w http.ResponseWriter, r *http.Request) {

	// Upgrade the regular http connection to a websocket connection
	ws, err := upgradeConnection.Upgrade(w, r, nil)

	if err != nil {

		log.Println("Error upgrading connection to WS 😫")
	}

	log.Println("Client connected to WS endpoint 😎🤘")

	var response WsJsonResponse
	response.Message = `<em><small>Connected to server 🤓🤘</small></em>`

	// Add a new user websocket connection to the websockets connections pool
	conn := WebSocketConnection{Conn: ws}
	ConnectedClients[conn] = ""

	err = ws.WriteJSON(response)

	if err != nil {
		log.Println("Error sending back a response form WS 😫")
	}

	// Start the ListenForWebSocket routine (that will run forever of till it stopped by us or by an error). Every time a new user joins, this routine will be launched and start to listen for messages coming from the user
	go ListenForWebSocket(&conn)

}

// Create a Go routine that listens to new websockets connection that are created when new clients(users) hit the Websockets end point. So it starts listening for any new user that is joining
func ListenForWebSocket(conn *WebSocketConnection) {

	// What to do if this Go routine stops
	defer func() {

		if routineError := recover(); routineError != nil {

			log.Println("Error in ListenForWebSockets routine 😫 => ", fmt.Sprintf("%v", routineError))
		}
	}()

	// payload = message from the client user
	var payload WsJsonPayload

	for {

		err := conn.ReadJSON(&payload)

		if err != nil {

			// If now payload => do nothing
		} else {

			payload.Connection = *conn
			WebSocketChannel <- payload
		}

	}

}

// Create a go routine that will listen to the Websockets activity channel
func ListenToTheWsChannelActivity() {

	var response WsJsonResponse

	for {
		// read event data from the channel
		event := <-WebSocketChannel

		switch event.Action {

		case "user_joined":
			// Get a list of all active chat users and send it back via broadcast
			// Update the list of connected users => add the new pair of websocket connection and user name, to the pool of connected users
			ConnectedClients[event.Connection] = event.Username
			users := getListOfConnectedUsers()
			response.Action = "list_users"
			response.ConnectedUsers = users
			broadcastToAll(response)
		
		case "user_left":
			response.Action = "list_users"
			delete(ConnectedClients, event.Connection)
			users := getListOfConnectedUsers()
			response.ConnectedUsers = users
			broadcastToAll(response)
		
		case "user_broadcasted_msg":
			response.Action = "broadcast"
			response.Message = fmt.Sprintf("<strong>%s</strong>: %s", event.Username, event.Message)
			broadcastToAll(response)

		}

		// response.Action = "Got Here 🤘🤓"
		// response.Message = fmt.Sprintf("Some fucking message 🤪 and action was %s", event.Action)
		// broadcastToAll(response)
	}

}

func getListOfConnectedUsers() []string {

	var userList []string

	for _, user := range ConnectedClients {

		if (user != "") {

			userList = append(userList, user)
		}
	}

	sort.Strings(userList)
	return userList
}

// Broadcast to all connected clients(users) what I've received from the web sockets channel
func broadcastToAll(response WsJsonResponse) {

	// Loop the websockets connections pool and for every client, do this...
	for client := range ConnectedClients {

		err := client.WriteJSON(response)

		if err != nil {

			// Maybe the client has left the chat..
			log.Println("Websocket err")
			// Close the client's connection
			_ = client.Close()
			// Remove the Web socket connection of the client who left
			delete(ConnectedClients, client)

		}
	}

}

// Render template function
func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {

	view, err := views.GetTemplate(tmpl)

	if err != nil {

		log.Println("There was an error 😫 => ", err)
		return err
	}

	err = view.Execute(w, data, nil)

	if err != nil {

		log.Println("There was an error 😫 => ", err)
		return err
	}

	return nil

}
