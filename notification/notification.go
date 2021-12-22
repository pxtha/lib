package notification

import (
	"crypto/ecdsa"
	"crypto/tls"
	"github.com/sirupsen/logrus"
	"log"
	Sync "sync"

	// android - FCM
	"github.com/NaySoftware/go-fcm"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
)

const (
	PLATFORM_APNs = 1
	PLATFORM_FCM  = 2
)

type AppNotification struct {
	// list of client
	clients map[string]*Client
	// config object loaded from file
	config Config
}

type Client struct {
	Platform int
	Mutex    *Sync.Mutex
	// android
	SenderID string

	// iOS
	AppBundleID string

	androidClient *fcm.FcmClient
	iOSClient     *apns2.Client
}

type Config struct {
	AndroidConfig *FCMConfig  `json:"android"`
	IOSConfig     *APNsConfig `json:"ios"`
}

type Message struct {
	Title       string
	Body        string
	PayloadData interface{}
	Tokens      []string
	Topic       string
	Sound       string
	Badge       string
}

func NewNotificationHelper(config Config) *AppNotification {
	android, iOS := &Client{}, &Client{}
	if config.AndroidConfig != nil {
		androidClient := FCMInitFromConfig(config.AndroidConfig)
		if androidClient != nil {
			android = &Client{
				Platform:      PLATFORM_FCM,
				Mutex:         &Sync.Mutex{},
				androidClient: androidClient,
				SenderID:      config.AndroidConfig.ClientID,
			}
		}
	}
	if config.IOSConfig != nil {
		iOSClient := APNsInitFromConfig(config.IOSConfig)
		if iOSClient != nil {
			iOS = &Client{
				Platform:    PLATFORM_APNs,
				Mutex:       &Sync.Mutex{},
				iOSClient:   iOSClient,
				AppBundleID: config.IOSConfig.AppBundleID,
			}
		}
	}
	return &AppNotification{
		clients: map[string]*Client{
			"FCM_" + config.AndroidConfig.ID: android,
			"APNs_" + config.IOSConfig.ID:    iOS,
		},
	}
}

func NewNotificationHelperFCM(config Config) *AppNotification {
	android := &Client{}
	if config.AndroidConfig != nil {
		androidClient := FCMInitFromConfig(config.AndroidConfig)
		if androidClient != nil {
			android = &Client{
				Platform:      PLATFORM_FCM,
				Mutex:         &Sync.Mutex{},
				androidClient: androidClient,
				SenderID:      config.AndroidConfig.ClientID,
			}
		}
		return &AppNotification{
			clients: map[string]*Client{
				"FCM_" + config.AndroidConfig.ID: android,
			},
		}
	}
	return &AppNotification{
		clients: map[string]*Client{},
	}
}

func (a *AppNotification) SendMessageForAll(msg *Message) {
	msg.Sound = "default"
	for key, client := range a.clients {
		log.Printf("SendMessageForAll key ", key)
		go a.SendMessage(client.Platform, key, msg)
	}
}

func (a *AppNotification) SendMessageForAndroid(msg *Message) {
	for key, client := range a.clients {
		if client.Platform == PLATFORM_FCM {
			go a.SendMessage(client.Platform, key, msg)
		}
	}
}

func (a *AppNotification) SendMessageForIOS(msg *Message) {
	for key, client := range a.clients {
		if client.Platform == PLATFORM_APNs {
			go a.SendMessage(client.Platform, key, msg)
		}
	}
}

func (a *AppNotification) SendMessage(platform int, clientID string, msg *Message) {
	if client, found := a.clients[clientID]; found {
		logrus.Infof("SendMessage client.Platform ", client.Platform)
		logrus.Infof("SendMessage token ", msg.Tokens)
		switch client.Platform {
		case PLATFORM_FCM:
			a._sendFCM(client, msg)
			break
		case PLATFORM_APNs:
			a._sendAPNs(client, msg)
			break
		default:
			log.Println("Unsupported platform ID platform ", platform)
			break
		}
	} else {
		log.Println("Unsupported platform ID client ", "Client not found")
	}
}

func (a *AppNotification) _sendAPNs(client *Client, msg *Message) {
	client.Mutex.Lock()
	// create payload data
	pData := payload.NewPayload()
	pData.AlertTitle(msg.Title)
	pData.AlertBody(msg.Body)
	pData.MutableContent()
	pData.Category(msg.Topic)
	msgData := msg.PayloadData.(map[string]interface{})
	for k, v := range msgData {
		pData.Custom(k, v)
	}
	// setup notification
	notification := &apns2.Notification{
		DeviceToken: "",
		Payload:     pData,
		Topic:       client.AppBundleID,
	}

	// send
	for _, deviceToken := range msg.Tokens {
		notification.DeviceToken = deviceToken
		status, err := client.iOSClient.Push(notification)
		if err != nil {
			logrus.Infof("send APNs error ", err.Error())
		} else {
			if status.Sent() {
				logrus.Infof("send APNs error ", "Sent successfully")
			} else {
				logrus.Errorf("send APNs error ", "Sent failed")
			}
		}
	}
	client.Mutex.Unlock()
}

func (a *AppNotification) _sendFCM(client *Client, msg *Message) {
	client.Mutex.Lock()
	// prepare and send message
	client.androidClient.SetPriority("high")
	client.androidClient.NewFcmRegIdsMsg(msg.Tokens, msg.PayloadData)
	client.androidClient.SetNotificationPayload(&fcm.NotificationPayload{
		Title: msg.Title,
		Body:  msg.Body,
		Sound: msg.Sound,
		Badge: msg.Badge,
		//TODO: if support for fcm APNs then need set other field here
	})
	status, err := client.androidClient.Send()
	if err == nil {
		logrus.Infof("send FCM ", "Sent successfully")
		logrus.Infof("Status Code", status.StatusCode)
		logrus.Infof("Success", status.Success)

	} else {
		logrus.Errorf("send FCM error", err.Error())
	}
	client.Mutex.Unlock()
}

//=================================================================
// Android service
// - Support server key
//=================================================================
type FCMConfig struct {
	ID        string `json:"id"`
	ServerKey string `json:"server_key"`
	ClientID  string `json:"client_id"`
}

func FCMInitFromConfig(config *FCMConfig) *fcm.FcmClient {
	if config.ServerKey == "" {
		log.Println("send FCMInitFromConfig error", "Error when init FCM client from config! Server key was empty please check again.")
		return nil
	}
	client := fcm.NewFcmClient(config.ServerKey)
	return client
}

//=================================================================
// iOS service
// - Support p12, pem, p8 key
//=================================================================
type APNsConfig struct {
	ID      string `json:"id"`
	KeyType string `json:"key_type"`
	// path of key file
	KeyFilePath string `json:"key_path"`
	KeyData     string `json:"key_data"`
	// not support for now
	KeyBase64     string
	KeyBase64Type string
	// password for .p12 and .pem
	Password string `json:"password"`
	// ID for .p8
	KeyID  string `json:"p8_key_id"`
	TeamID string `json:"p8_team_id"`
	// type
	IsProduction bool   `json:"is_production"`
	AppBundleID  string `json:"app_bundle_id"`
}

func APNsInitFromConfig(config *APNsConfig) *apns2.Client {
	switch config.KeyType {
	case "p12":
		certificateKey, err := certificate.FromP12File(config.KeyFilePath, config.Password)
		if err != nil {
			log.Println("platform APNs:Error when create certificate, Please check key file is correct or not ", err.Error())
			return nil
		}
		return apns_NewClient_P12_Pem(certificateKey, config)
	case "pem":
		certificateKey, err := certificate.FromPemFile(config.KeyFilePath, config.Password)
		if err != nil {
			log.Println("platform APNs:Error when create certificate, Please check key file is correct or not ", err.Error())
			return nil
		}
		return apns_NewClient_P12_Pem(certificateKey, config)
	case "p8":
		var err error
		var authKey *ecdsa.PrivateKey
		if config.KeyFilePath != "" {
			authKey, err = token.AuthKeyFromFile(config.KeyFilePath)
		} else {
			authKey, err = token.AuthKeyFromBytes([]byte(config.KeyData))
		}
		if err != nil {
			log.Println("platform APNs:Error when create authentication key, Please check key file is correct or not", err.Error())
			return nil
		}
		return apns_NewClient_P8(authKey, config)
	default:
		log.Println("platform APNs:Init APNs failed. Key was invalid. APNs only support for .p12, .pem, .p8")
		break
	}
	return nil
}

func apns_NewClient_P12_Pem(cer tls.Certificate, config *APNsConfig) *apns2.Client {
	if config.IsProduction {
		return apns2.NewClient(cer).Production()
	} else {
		return apns2.NewClient(cer).Development()
	}
}

func apns_NewClient_P8(authKey *ecdsa.PrivateKey, config *APNsConfig) *apns2.Client {
	// init token
	token := &token.Token{
		AuthKey: authKey,
		// KeyID from developer account (Certificates, Identifiers & Profiles -> Keys)
		KeyID: config.KeyID,
		// TeamID from developer account (View Account -> Membership)
		TeamID: config.TeamID,
	}
	// create client
	if config.IsProduction {
		return apns2.NewTokenClient(token).Production()
	} else {
		return apns2.NewTokenClient(token).Development()
	}
}
