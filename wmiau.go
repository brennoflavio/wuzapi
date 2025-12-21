package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/appstate"
	"go.mau.fi/whatsmeow/proto/waCompanionReg"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// db field declaration as *sqlx.DB
type MyClient struct {
	WAClient       *whatsmeow.Client
	eventHandlerID uint32
	userID         string
	db             *sqlx.DB
	s              *server
}

// saveEvent stores an event in the events table
func saveEvent(db *sqlx.DB, userID string, eventType string, payload map[string]interface{}) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal event payload to JSON")
		return err
	}

	_, err = db.Exec(
		"INSERT INTO events (user_id, event_type, payload) VALUES ($1, $2, $3)",
		userID, eventType, string(jsonPayload),
	)
	if err != nil {
		log.Error().Err(err).Str("eventType", eventType).Str("userID", userID).Msg("Failed to save event to database")
		return err
	}

	log.Debug().Str("eventType", eventType).Str("userID", userID).Msg("Event saved to database")
	return nil
}

// processEvent stores the event in the database
func processEvent(mycli *MyClient, postmap map[string]interface{}) {
	eventType, ok := postmap["type"].(string)
	if !ok {
		log.Error().Msg("Event type is not a string in postmap")
		return
	}

	// Save event to database
	if err := saveEvent(mycli.db, mycli.userID, eventType, postmap); err != nil {
		log.Error().Err(err).Str("eventType", eventType).Msg("Failed to save event")
	}
}

func parseJID(arg string) (types.JID, bool) {
	if arg[0] == '+' {
		arg = arg[1:]
	}
	if !strings.ContainsRune(arg, '@') {
		return types.NewJID(arg, types.DefaultUserServer), true
	} else {
		recipient, err := types.ParseJID(arg)
		if err != nil {
			log.Error().Err(err).Msg("Invalid JID")
			return recipient, false
		} else if recipient.User == "" {
			log.Error().Err(err).Msg("Invalid JID no server specified")
			return recipient, false
		}
		return recipient, true
	}
}

func (s *server) startClient(userID string, textjid string) {
	log.Info().Str("userid", userID).Str("jid", textjid).Msg("Starting websocket connection to Whatsapp")

	// Connection retry constants
	const maxConnectionRetries = 3
	const connectionRetryBaseWait = 5 * time.Second

	var deviceStore *store.Device
	var err error

	// First handle the device store initialization
	if textjid != "" {
		jid, _ := parseJID(textjid)
		deviceStore, err = container.GetDevice(context.Background(), jid)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get device")
			deviceStore = container.NewDevice()
		}
	} else {
		log.Warn().Msg("No jid found. Creating new device")
		deviceStore = container.NewDevice()
	}

	if deviceStore == nil {
		log.Warn().Msg("No store found. Creating new one")
		deviceStore = container.NewDevice()
	}

	clientLog := waLog.Stdout("Client", *waDebug, false)

	// Create the client with initialized deviceStore
	var client *whatsmeow.Client
	if *waDebug != "" {
		client = whatsmeow.NewClient(deviceStore, clientLog)
	} else {
		client = whatsmeow.NewClient(deviceStore, nil)
	}

	// Now we can use the client with the manager
	clientManager.SetWhatsmeowClient(client)

	store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_DESKTOP.Enum()
	store.DeviceProps.Os = osName

	mycli := MyClient{client, 1, userID, s.db, s}
	mycli.eventHandlerID = mycli.WAClient.AddEventHandler(mycli.myEventHandler)

	// Store the MyClient in clientManager
	clientManager.SetMyClient(&mycli)

	if client.Store.ID == nil {
		// No ID stored, new login
		qrChan, err := client.GetQRChannel(context.Background())
		if err != nil {
			// This error means that we're already logged in, so ignore it.
			if !errors.Is(err, whatsmeow.ErrQRStoreContainsID) {
				log.Error().Err(err).Msg("Failed to get QR channel")
				return
			}
		} else {
			err = client.Connect() // Must connect to generate QR code
			if err != nil {
				log.Error().Err(err).Msg("Failed to connect client")
				return
			}

			for evt := range qrChan {
				if evt.Event == "code" {
					log.Info().Msg("QR code received")
					// Store encoded/embeded base64 QR on database for retrieval with the /qr endpoint
					image, _ := qrcode.Encode(evt.Code, qrcode.Medium, 256)
					base64qrcode := "data:image/png;base64," + base64.StdEncoding.EncodeToString(image)
					sqlStmt := `UPDATE users SET qrcode=$1 WHERE id=$2`
					_, err := s.db.Exec(sqlStmt, base64qrcode, userID)
					if err != nil {
						log.Error().Err(err).Msg(sqlStmt)
					} else {
						updateGlobalUser("Qrcode", base64qrcode)
						log.Info().Str("qrcode", base64qrcode).Msg("update global user with qr code")
					}

					// Store QR code event
					postmap := make(map[string]interface{})
					postmap["event"] = evt.Event
					postmap["qrCodeBase64"] = base64qrcode
					postmap["type"] = "QR"

					processEvent(&mycli, postmap)

				} else if evt.Event == "timeout" {
					// Clear QR code from DB on timeout
					// Store QR timeout event before cleanup
					postmap := make(map[string]interface{})
					postmap["event"] = evt.Event
					postmap["type"] = "QRTimeout"
					processEvent(&mycli, postmap)

					sqlStmt := `UPDATE users SET qrcode='' WHERE id=$1`
					_, err := s.db.Exec(sqlStmt, userID)
					if err != nil {
						log.Error().Err(err).Msg(sqlStmt)
					} else {
						updateGlobalUser("Qrcode", "")
					}
					log.Warn().Msg("QR timeout killing channel")
					clientManager.DeleteWhatsmeowClient()
					clientManager.DeleteMyClient()
					select {
					case killchannel[userID] <- true:
					default:
					}
				} else if evt.Event == "success" {
					log.Info().Msg("QR pairing ok!")
					// Clear QR code after pairing
					sqlStmt := `UPDATE users SET qrcode='', connected=1 WHERE id=$1`
					_, err := s.db.Exec(sqlStmt, userID)
					if err != nil {
						log.Error().Err(err).Msg(sqlStmt)
					} else {
						updateGlobalUser("Qrcode", "")
					}
				} else {
					log.Info().Str("event", evt.Event).Msg("Login event")
				}
			}
		}

	} else {
		// Already logged in, just connect
		log.Info().Msg("Already logged in, just connect")

		// Retry logic with linear backoff
		var lastErr error

		for attempt := 0; attempt < maxConnectionRetries; attempt++ {
			if attempt > 0 {
				waitTime := time.Duration(attempt) * connectionRetryBaseWait
				log.Warn().
					Int("attempt", attempt+1).
					Int("max_retries", maxConnectionRetries).
					Dur("wait_time", waitTime).
					Msg("Retrying connection after delay")
				time.Sleep(waitTime)
			}

			err = client.Connect()
			if err == nil {
				log.Info().
					Int("attempt", attempt+1).
					Msg("Successfully connected to WhatsApp")
				break
			}

			lastErr = err
			log.Warn().
				Err(err).
				Int("attempt", attempt+1).
				Int("max_retries", maxConnectionRetries).
				Msg("Failed to connect to WhatsApp")
		}

		if lastErr != nil {
			log.Error().
				Err(lastErr).
				Str("userid", userID).
				Int("attempts", maxConnectionRetries).
				Msg("Failed to connect to WhatsApp after all retry attempts")

			clientManager.DeleteWhatsmeowClient()
			clientManager.DeleteMyClient()

			sqlStmt := `UPDATE users SET qrcode='', connected=0 WHERE id=$1`
			_, dbErr := s.db.Exec(sqlStmt, userID)
			if dbErr != nil {
				log.Error().Err(dbErr).Msg("Failed to update user status after connection error")
			}

			// Use the existing mycli instance from outer scope
			postmap := make(map[string]interface{})
			postmap["event"] = "ConnectFailure"
			postmap["error"] = lastErr.Error()
			postmap["type"] = "ConnectFailure"
			postmap["attempts"] = maxConnectionRetries
			postmap["reason"] = "Failed to connect after retry attempts"
			processEvent(&mycli, postmap)

			return
		}
	}

	// Keep connected client live until disconnected/killed
	for {
		select {
		case <-killchannel[userID]:
			log.Info().Str("userid", userID).Msg("Received kill signal")
			client.Disconnect()
			clientManager.DeleteWhatsmeowClient()
			clientManager.DeleteMyClient()
			sqlStmt := `UPDATE users SET qrcode='', connected=0 WHERE id=$1`
			_, err := s.db.Exec(sqlStmt, userID)
			if err != nil {
				log.Error().Err(err).Msg(sqlStmt)
			}
			delete(killchannel, userID)
			return
		default:
			time.Sleep(1000 * time.Millisecond)
			//log.Info().Str("jid",textjid).Msg("Loop the loop")
		}
	}
}

func fileToBase64(filepath string) (string, string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", "", err
	}
	mimeType := http.DetectContentType(data)
	return base64.StdEncoding.EncodeToString(data), mimeType, nil
}

func (mycli *MyClient) myEventHandler(rawEvt interface{}) {
	txtid := mycli.userID
	postmap := make(map[string]interface{})
	postmap["event"] = rawEvt

	switch evt := rawEvt.(type) {
	case *events.AppStateSyncComplete:
		postmap["type"] = "AppStateSyncComplete"
		if len(mycli.WAClient.Store.PushName) > 0 && evt.Name == appstate.WAPatchCriticalBlock {
			err := mycli.WAClient.SendPresence(context.Background(), types.PresenceAvailable)
			if err != nil {
				log.Warn().Err(err).Msg("Failed to send available presence")
			} else {
				log.Info().Msg("Marked self as available")
			}
		}
	case *events.Connected, *events.PushNameSetting:
		postmap["type"] = "Connected"
		if len(mycli.WAClient.Store.PushName) == 0 {
			break
		}
		// Send presence available when connecting and when the pushname is changed.
		// This makes sure that outgoing messages always have the right pushname.
		err := mycli.WAClient.SendPresence(context.Background(), types.PresenceAvailable)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to send available presence")
		} else {
			log.Info().Msg("Marked self as available")
		}
		sqlStmt := `UPDATE users SET connected=1 WHERE id=$1`
		_, err = mycli.db.Exec(sqlStmt, mycli.userID)
		if err != nil {
			log.Error().Err(err).Msg(sqlStmt)
			return
		}
	case *events.PairSuccess:
		log.Info().Str("userid", mycli.userID).Str("ID", evt.ID.String()).Str("BusinessName", evt.BusinessName).Str("Platform", evt.Platform).Msg("QR Pair Success")
		jid := evt.ID
		sqlStmt := `UPDATE users SET jid=$1 WHERE id=$2`
		_, err := mycli.db.Exec(sqlStmt, jid, mycli.userID)
		if err != nil {
			log.Error().Err(err).Msg(sqlStmt)
			return
		}

		postmap["type"] = "PairSuccess"

		txtid = globalUser.Get("Id")
		updateGlobalUser("Jid", fmt.Sprintf("%s", jid))
		log.Info().Str("jid", jid.String()).Str("userid", txtid).Msg("User information set")
	case *events.StreamReplaced:
		postmap["type"] = "StreamReplaced"
		log.Info().Msg("Received StreamReplaced event")
	case *events.Message:
		lastMessageCache.Set(mycli.userID, &evt.Info, cache.DefaultExpiration)

		postmap["type"] = "Message"
		metaParts := []string{fmt.Sprintf("pushname: %s", evt.Info.PushName), fmt.Sprintf("timestamp: %s", evt.Info.Timestamp)}
		if evt.Info.Type != "" {
			metaParts = append(metaParts, fmt.Sprintf("type: %s", evt.Info.Type))
		}
		if evt.Info.Category != "" {
			metaParts = append(metaParts, fmt.Sprintf("category: %s", evt.Info.Category))
		}
		if evt.IsViewOnce {
			metaParts = append(metaParts, "view once")
		}
		if evt.IsViewOnce {
			metaParts = append(metaParts, "ephemeral")
		}

		log.Info().Str("id", evt.Info.ID).Str("source", evt.Info.SourceString()).Str("parts", strings.Join(metaParts, ", ")).Msg("Message Received")

		if !*skipMedia {
			// try to get Image if any
			img := evt.Message.GetImageMessage()
			if img != nil {
				// Download the image
				data, err := mycli.WAClient.Download(context.Background(), img)
				if err != nil {
					log.Error().Err(err).Msg("Failed to download image")
					return
				}

				// Get sender JID for inbox/outbox determination
				isIncoming := !evt.Info.IsFromMe
				contactJID := evt.Info.Sender.String()
				if evt.Info.IsGroup {
					contactJID = evt.Info.Chat.String()
				}

				// Save file to disk
				fileData, err := GetFileManager().SaveFile(
					txtid,
					contactJID,
					evt.Info.ID,
					data,
					img.GetMimetype(),
					evt.Info.ID,
					isIncoming,
				)
				if err != nil {
					log.Error().Err(err).Msg("Failed to save image to disk")
				} else {
					postmap["file"] = fileData
				}

				log.Info().Str("path", fileData.Path).Msg("Image processed")
			}

			// try to get Audio if any
			audio := evt.Message.GetAudioMessage()
			if audio != nil {
				// Download the audio
				data, err := mycli.WAClient.Download(context.Background(), audio)
				if err != nil {
					log.Error().Err(err).Msg("Failed to download audio")
					return
				}

				// Get sender JID for inbox/outbox determination
				isIncoming := !evt.Info.IsFromMe
				contactJID := evt.Info.Sender.String()
				if evt.Info.IsGroup {
					contactJID = evt.Info.Chat.String()
				}

				// Save file to disk
				fileData, err := GetFileManager().SaveFile(
					txtid,
					contactJID,
					evt.Info.ID,
					data,
					audio.GetMimetype(),
					evt.Info.ID,
					isIncoming,
				)
				if err != nil {
					log.Error().Err(err).Msg("Failed to save audio to disk")
				} else {
					postmap["file"] = fileData
				}

				log.Info().Str("path", fileData.Path).Msg("Audio processed")
			}

			// try to get Document if any
			document := evt.Message.GetDocumentMessage()
			if document != nil {
				// Download the document
				data, err := mycli.WAClient.Download(context.Background(), document)
				if err != nil {
					log.Error().Err(err).Msg("Failed to download document")
					return
				}

				// Get sender JID for inbox/outbox determination
				isIncoming := !evt.Info.IsFromMe
				contactJID := evt.Info.Sender.String()
				if evt.Info.IsGroup {
					contactJID = evt.Info.Chat.String()
				}

				// Save file to disk
				fileData, err := GetFileManager().SaveFile(
					txtid,
					contactJID,
					evt.Info.ID,
					data,
					document.GetMimetype(),
					evt.Info.ID,
					isIncoming,
				)
				if err != nil {
					log.Error().Err(err).Msg("Failed to save document to disk")
				} else {
					postmap["file"] = fileData
				}

				log.Info().Str("path", fileData.Path).Msg("Document processed")
			}

			// try to get Video if any
			video := evt.Message.GetVideoMessage()
			if video != nil {
				// Download the video
				data, err := mycli.WAClient.Download(context.Background(), video)
				if err != nil {
					log.Error().Err(err).Msg("Failed to download video")
					return
				}

				// Get sender JID for inbox/outbox determination
				isIncoming := !evt.Info.IsFromMe
				contactJID := evt.Info.Sender.String()
				if evt.Info.IsGroup {
					contactJID = evt.Info.Chat.String()
				}

				// Save file to disk
				fileData, err := GetFileManager().SaveFile(
					txtid,
					contactJID,
					evt.Info.ID,
					data,
					video.GetMimetype(),
					evt.Info.ID,
					isIncoming,
				)
				if err != nil {
					log.Error().Err(err).Msg("Failed to save video to disk")
				} else {
					postmap["file"] = fileData
				}

				log.Info().Str("path", fileData.Path).Msg("Video processed")
			}

			sticker := evt.Message.GetStickerMessage()
			if sticker != nil {
				// Download the sticker
				data, err := mycli.WAClient.Download(context.Background(), sticker)
				if err != nil {
					log.Error().Err(err).Msg("Failed to download sticker")
					return
				}

				// Get sender JID for inbox/outbox determination
				isIncoming := !evt.Info.IsFromMe
				contactJID := evt.Info.Sender.String()
				if evt.Info.IsGroup {
					contactJID = evt.Info.Chat.String()
				}

				// Save file to disk
				fileData, err := GetFileManager().SaveFile(
					txtid,
					contactJID,
					evt.Info.ID,
					data,
					sticker.GetMimetype(),
					evt.Info.ID,
					isIncoming,
				)
				if err != nil {
					log.Error().Err(err).Msg("Failed to save sticker to disk")
				} else {
					postmap["file"] = fileData
				}

				// useful metadata (optional, but handy)
				postmap["isSticker"] = true
				postmap["stickerAnimated"] = sticker.GetIsAnimated()

				log.Info().Str("path", fileData.Path).Msg("Sticker processed")
			}

		}

	case *events.Receipt:
		postmap["type"] = "ReadReceipt"
		//if evt.Type == events.ReceiptTypeRead || evt.Type == events.ReceiptTypeReadSelf {
		if evt.Type == types.ReceiptTypeRead || evt.Type == types.ReceiptTypeReadSelf {
			log.Info().Strs("id", evt.MessageIDs).Str("source", evt.SourceString()).Str("timestamp", fmt.Sprintf("%v", evt.Timestamp)).Msg("Message was read")
			//if evt.Type == events.ReceiptTypeRead {
			if evt.Type == types.ReceiptTypeRead {
				postmap["state"] = "Read"
			} else {
				postmap["state"] = "ReadSelf"
			}
			//} else if evt.Type == events.ReceiptTypeDelivered {
		} else if evt.Type == types.ReceiptTypeDelivered {
			postmap["state"] = "Delivered"
			log.Info().Str("id", evt.MessageIDs[0]).Str("source", evt.SourceString()).Str("timestamp", fmt.Sprintf("%v", evt.Timestamp)).Msg("Message delivered")
		} else {
			// Skip storing inactive or other delivery types
			return
		}
	case *events.Presence:
		postmap["type"] = "Presence"
		if evt.Unavailable {
			postmap["state"] = "offline"
			if evt.LastSeen.IsZero() {
				log.Info().Str("from", evt.From.String()).Msg("User is now offline")
			} else {
				log.Info().Str("from", evt.From.String()).Str("lastSeen", fmt.Sprintf("%v", evt.LastSeen)).Msg("User is now offline")
			}
		} else {
			postmap["state"] = "online"
			log.Info().Str("from", evt.From.String()).Msg("User is now online")
		}
	case *events.HistorySync:
		postmap["type"] = "HistorySync"

	case *events.AppState:
		log.Info().Str("index", fmt.Sprintf("%+v", evt.Index)).Str("actionValue", fmt.Sprintf("%+v", evt.SyncActionValue)).Msg("App state event received")
	case *events.LoggedOut:
		postmap["type"] = "LoggedOut"
		log.Info().Str("reason", evt.Reason.String()).Msg("Logged out")
		defer func() {
			// Use a non-blocking send to prevent a deadlock if the receiver has already terminated.
			select {
			case killchannel[mycli.userID] <- true:
			default:
			}
		}()
		sqlStmt := `UPDATE users SET connected=0 WHERE id=$1`
		_, err := mycli.db.Exec(sqlStmt, mycli.userID)
		if err != nil {
			log.Error().Err(err).Msg(sqlStmt)
			return
		}
	case *events.ChatPresence:
		postmap["type"] = "ChatPresence"
		log.Info().Str("state", fmt.Sprintf("%s", evt.State)).Str("media", fmt.Sprintf("%s", evt.Media)).Str("chat", evt.MessageSource.Chat.String()).Str("sender", evt.MessageSource.Sender.String()).Msg("Chat Presence received")
	case *events.CallOffer:
		postmap["type"] = "CallOffer"
		log.Info().Str("event", fmt.Sprintf("%+v", evt)).Msg("Got call offer")
	case *events.CallAccept:
		postmap["type"] = "CallAccept"
		log.Info().Str("event", fmt.Sprintf("%+v", evt)).Msg("Got call accept")
	case *events.CallTerminate:
		postmap["type"] = "CallTerminate"
		log.Info().Str("event", fmt.Sprintf("%+v", evt)).Msg("Got call terminate")
	case *events.CallOfferNotice:
		postmap["type"] = "CallOfferNotice"
		log.Info().Str("event", fmt.Sprintf("%+v", evt)).Msg("Got call offer notice")
	case *events.CallRelayLatency:
		postmap["type"] = "CallRelayLatency"
		log.Info().Str("event", fmt.Sprintf("%+v", evt)).Msg("Got call relay latency")
	case *events.Disconnected:
		postmap["type"] = "Disconnected"
		log.Info().Str("reason", fmt.Sprintf("%+v", evt)).Msg("Disconnected from Whatsapp")
	case *events.ConnectFailure:
		postmap["type"] = "ConnectFailure"
		log.Error().Str("reason", fmt.Sprintf("%+v", evt)).Msg("Failed to connect to Whatsapp")
	case *events.UndecryptableMessage:
		postmap["type"] = "UndecryptableMessage"
		log.Warn().Str("info", evt.Info.SourceString()).Msg("Undecryptable message received")
	case *events.MediaRetry:
		postmap["type"] = "MediaRetry"
		log.Info().Str("messageID", evt.MessageID).Msg("Media retry event")
	case *events.GroupInfo:
		postmap["type"] = "GroupInfo"
		log.Info().Str("jid", evt.JID.String()).Msg("Group info updated")
	case *events.JoinedGroup:
		postmap["type"] = "JoinedGroup"
		log.Info().Str("jid", evt.JID.String()).Msg("Joined group")
	case *events.Picture:
		postmap["type"] = "Picture"
		log.Info().Str("jid", evt.JID.String()).Msg("Picture updated")
	case *events.BlocklistChange:
		postmap["type"] = "BlocklistChange"
		log.Info().Str("jid", evt.JID.String()).Msg("Blocklist changed")
	case *events.Blocklist:
		postmap["type"] = "Blocklist"
		log.Info().Msg("Blocklist received")
	case *events.KeepAliveRestored:
		postmap["type"] = "KeepAliveRestored"
		log.Info().Msg("Keep alive restored")
	case *events.KeepAliveTimeout:
		postmap["type"] = "KeepAliveTimeout"
		log.Warn().Msg("Keep alive timeout")
	case *events.ClientOutdated:
		postmap["type"] = "ClientOutdated"
		log.Warn().Msg("Client outdated")
	case *events.TemporaryBan:
		postmap["type"] = "TemporaryBan"
		log.Info().Msg("Temporary ban")
	case *events.StreamError:
		postmap["type"] = "StreamError"
		log.Error().Str("code", evt.Code).Msg("Stream error")
	case *events.PairError:
		postmap["type"] = "PairError"
		log.Error().Msg("Pair error")
	case *events.PrivacySettings:
		postmap["type"] = "PrivacySettings"
		log.Info().Msg("Privacy settings updated")
	case *events.UserAbout:
		postmap["type"] = "UserAbout"
		log.Info().Str("jid", evt.JID.String()).Msg("User about updated")
	case *events.OfflineSyncCompleted:
		postmap["type"] = "OfflineSyncCompleted"
		log.Info().Msg("Offline sync completed")
	case *events.OfflineSyncPreview:
		postmap["type"] = "OfflineSyncPreview"
		log.Info().Msg("Offline sync preview")
	case *events.IdentityChange:
		postmap["type"] = "IdentityChange"
		log.Info().Str("jid", evt.JID.String()).Msg("Identity changed")
	case *events.NewsletterJoin:
		postmap["type"] = "NewsletterJoin"
		log.Info().Str("jid", evt.ID.String()).Msg("Newsletter joined")
	case *events.NewsletterLeave:
		postmap["type"] = "NewsletterLeave"
		log.Info().Str("jid", evt.ID.String()).Msg("Newsletter left")
	case *events.NewsletterMuteChange:
		postmap["type"] = "NewsletterMuteChange"
		log.Info().Str("jid", evt.ID.String()).Msg("Newsletter mute changed")
	case *events.NewsletterLiveUpdate:
		postmap["type"] = "NewsletterLiveUpdate"
		log.Info().Msg("Newsletter live update")
	case *events.FBMessage:
		postmap["type"] = "FBMessage"
		log.Info().Str("info", evt.Info.SourceString()).Msg("Facebook message received")
	default:
		postmap["type"] = fmt.Sprintf("%T", evt)
		log.Warn().Str("event", fmt.Sprintf("%+v", evt)).Msg("Unhandled event")
	}

	processEvent(mycli, postmap)
}
