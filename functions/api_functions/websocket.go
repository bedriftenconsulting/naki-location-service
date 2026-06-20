package api_functions

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/naki/location-service/models"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type trackingClient struct {
	conn    *websocket.Conn
	visitID string
}

var (
	clients   = make(map[string][]*trackingClient)
	clientsMu sync.RWMutex
)

func HandleTrackingWebSocket(c *gin.Context) {
	visitID := c.Param("visit_id")
	if visitID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "message": "visit_id required"})
		return
	}

	_, err := GetActiveVisit(visitID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": 404, "message": "visit not active"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}

	client := &trackingClient{conn: conn, visitID: visitID}

	clientsMu.Lock()
	clients[visitID] = append(clients[visitID], client)
	clientsMu.Unlock()

	log.Printf("tracking client connected for visit %s (total: %d)", visitID, len(clients[visitID]))

	defer func() {
		conn.Close()
		removeClient(visitID, client)
	}()

	info, err := GetTrackingInfo(visitID)
	if err == nil {
		conn.WriteJSON(info)
	}

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func removeClient(visitID string, client *trackingClient) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	conns := clients[visitID]
	for i, c := range conns {
		if c == client {
			clients[visitID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}

	if len(clients[visitID]) == 0 {
		delete(clients, visitID)
	}

	log.Printf("tracking client disconnected for visit %s", visitID)
}

func notifyTrackingClients(visitID string, loc models.NurseLocation) {
	clientsMu.RLock()
	conns := clients[visitID]
	clientsMu.RUnlock()

	if len(conns) == 0 {
		return
	}

	info, err := GetTrackingInfo(visitID)
	if err != nil {
		return
	}

	for _, client := range conns {
		client.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if err := client.conn.WriteJSON(info); err != nil {
			log.Printf("failed to send tracking update: %v", err)
		}
	}
}

func CloseVisitTracking(visitID string) {
	clientsMu.Lock()
	conns := clients[visitID]
	delete(clients, visitID)
	clientsMu.Unlock()

	for _, client := range conns {
		client.conn.WriteJSON(gin.H{"status": "visit_completed", "visit_id": visitID})
		client.conn.Close()
	}

	log.Printf("closed all tracking connections for visit %s", visitID)
}
