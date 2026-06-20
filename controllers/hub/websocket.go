package hub

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"meow.net/utils"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 16384,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func HubNegotiate(w http.ResponseWriter, r *http.Request) {
	log.Printf("[HUB] negotiate")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"connectionId":     generateConnectionId(),
		"negotiateVersion": 0,
		"availableTransports": []map[string]interface{}{
			{
				"transport":       "WebSockets",
				"transferFormats": []string{"Text", "Binary"},
			},
			{
				"transport":       "ServerSentEvents",
				"transferFormats": []string{"Text"},
			},
			{
				"transport":       "LongPolling",
				"transferFormats": []string{"Text", "Binary"},
			},
		},
	})
}

func generateConnectionId() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func HubWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[HUB] upgrade error: %v", err)
		return
	}
	defer conn.Close()

	playerId := 0
	tokenStr := r.URL.Query().Get("access_token")
	if tokenStr == "" {
		tokenStr = utils.GetBearerToken(r)
	}
	if tokenStr != "" {
		if sub, err := utils.ParseSubFromJWT(tokenStr); err == nil {
			if id, err := strconv.Atoi(sub); err == nil {
				playerId = id
			}
		}
	}
	log.Printf("[HUB] WebSocket connected pid=%d", playerId)

	state := &connState{
		playerId:  playerId,
		conn:      conn,
		playerIds: map[int]bool{},
		send:      make(chan []byte, sendQueueSize),
		done:      make(chan struct{}),
	}
	if !hubRegister(state) {
		log.Printf("[HUB] connection limit reached pid=%d", playerId)
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseTryAgainLater, "connection limit reached"))
		return
	}
	defer func() {
		wentOffline := hubUnregister(state)
		log.Printf("[HUB] disconnected pid=%d offline=%v", playerId, wentOffline)
		if wentOffline && playerId != 0 {
			ClearPlayerInstance(playerId)
			MarkPlayerOffline(playerId)
			ClearLoginLock(playerId)
		}
		conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseGoingAway, ""),
			time.Now().Add(2*time.Second))
		state.close()
	}()

	go state.writePump()

	state.writeFrame([]byte(`{"SessionId":1}`))

	const wsReadTimeout = 5 * time.Minute
	conn.SetReadDeadline(time.Now().Add(wsReadTimeout))
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Printf("[HUB] read err pid=%d: %v", playerId, err)
			} else {
				log.Printf("[HUB] read closed pid=%d: %v", playerId, err)
			}
			break
		}
		conn.SetReadDeadline(time.Now().Add(wsReadTimeout))

		msg = []byte(strings.TrimRight(string(msg), "\x1e"))

		var payload map[string]interface{}
		if err := json.Unmarshal(msg, &payload); err != nil {
			log.Printf("[HUB] parse error: %v raw: %s", err, msg)
			continue
		}

		if _, ok := payload["protocol"]; ok {
			log.Printf("[HUB] handshake complete pid=%d", playerId)
			state.writeFrame([]byte("{}\x1e"))
			if playerId != 0 {
				state.sendInitialState()
				HubBroadcastPresence(playerId)
			}
			continue
		}

		if t, ok := payload["type"].(float64); ok && t == 6 {
			pong, _ := json.Marshal(map[string]interface{}{"type": 6})
			state.writeFrame(append(pong, 0x1e))
			continue
		}

		target, _ := payload["target"].(string)
		invId, hasInvId := payload["invocationId"]
		if !hasInvId {
			log.Printf("[HUB] non-invocation: type=%v target=%s", payload["type"], target)
			continue
		}

		if target != "SubscribeToPlayers" {
			log.Printf("[HUB] invocation: %s id=%v args=%v", target, invId, payload["arguments"])
		}

		switch target {
		case "heartbeat2":
			presence := BuildPresence(playerId)
			resp, _ := json.Marshal(map[string]interface{}{
				"type":         3,
				"invocationId": invId,
				"result": map[string]interface{}{
					"Id":  playerId,
					"Msg": presence,
				},
			})
			state.writeFrame(append(resp, 0x1e))
			state.writeFrame(NotifFrame("PresenceUpdate", presence))

		case "playerSubscriptions/v1/update", "SubscribeToPlayers":
			resp, _ := json.Marshal(map[string]interface{}{
				"type":         3,
				"invocationId": invId,
				"result":       nil,
			})
			state.writeFrame(append(resp, 0x1e))

			args, _ := payload["arguments"].([]interface{})
			if len(args) == 0 {
				continue
			}
			argMap, _ := args[0].(map[string]interface{})
			rawIds, _ := argMap["PlayerIds"].([]interface{})
			intIds := make([]int, 0, len(rawIds))
			for _, pid := range rawIds {
				if id, ok := pid.(float64); ok {
					intIds = append(intIds, int(id))
				}
			}

			added := state.updateSubscriptions(intIds)
			for _, pid := range added {
				state.writeFrame(NotifFrame("PresenceUpdate", BuildPresence(pid)))
			}

		default:
			log.Printf("[HUB] unknown invocation: %s", target)
			resp, _ := json.Marshal(map[string]interface{}{
				"type":         3,
				"invocationId": invId,
				"result":       nil,
			})
			state.writeFrame(append(resp, 0x1e))
		}
	}
}
