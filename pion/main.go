package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/didip/tollbooth/v6"
	"github.com/pion/webrtc/v3"
	"github.com/rs/cors"
	"github.com/spf13/viper"
)

type PeerConnections struct {
	mutex       sync.RWMutex
	connections []*webrtc.PeerConnection
}

var (
	peerConnections PeerConnections
	config          Config
	debug           bool
)

type Config struct {
	Port           int     `mapstructure:"SERVER_PORT"`
	Addr           string  `mapstructure:"SERVER_ADDR"`
	CertFile       string  `mapstructure:"CERT_FILE"`
	KeyFile        string  `mapstructure:"KEY_FILE"`
	StunUrl        string  `mapstructure:"STUN_URL"`
	ExternalIP     string  `mapstructure:"EXTERNAL_IP"`
	AllowedOrigins string  `mapstructure:"ALLOWED_ORIGINS"`
	Debug          bool    `mapstructure:"DEBUG"`
	MinPort        uint16  `mapstructure:"MIN_PORT"`
	MaxPort        uint16  `mapstructure:"MAX_PORT"`
	MaxReqRate     float64 `mapstructure:"MAX_REQ_RATE"`
}

func LoadConfig() (config Config, err error) {
	viper.AutomaticEnv()

	if err := viper.BindEnv("SERVER_PORT"); err != nil {
		return Config{}, err
	}
	viper.SetDefault("SERVER_PORT", 443)

	if err := viper.BindEnv("SERVER_ADDR"); err != nil {
		return Config{}, err
	}
	viper.SetDefault("SERVER_ADDR", "127.0.0.1")

	if err := viper.BindEnv("STUN_URL"); err != nil {
		return Config{}, err
	}

	if err := viper.BindEnv("CERT_FILE"); err != nil {
		return Config{}, err
	}

	if err := viper.BindEnv("KEY_FILE"); err != nil {
		return Config{}, err
	}

	if err := viper.BindEnv("EXTERNAL_IP"); err != nil {
		return Config{}, err
	}

	if err := viper.BindEnv("ALLOWED_ORIGINS"); err != nil {
		return Config{}, err
	}

	if err := viper.BindEnv("DEBUG"); err != nil {
		return Config{}, err
	}
	viper.SetDefault("DEBUG", false)

	if err := viper.BindEnv("MIN_PORT"); err != nil {
		return Config{}, err
	}

	if err := viper.BindEnv("MAX_PORT"); err != nil {
		return Config{}, err
	}

	if err := viper.BindEnv("MAX_REQ_RATE"); err != nil {
		return Config{}, err
	}

	err = viper.Unmarshal(&config)
	return
}

func listPeerConnections() {
	fmt.Println("Active PeerConnections:")
	peerConnections.mutex.Lock()
	defer peerConnections.mutex.Unlock()

	for i := range peerConnections.connections {
		c := peerConnections.connections[i]
		//		id := c.getStatsID()
		//		fmt.Printf("PeerConnection ID: %s\n", id)
		fmt.Println(c)
	}
}

func offer(w http.ResponseWriter, r *http.Request) {
	var offer webrtc.SessionDescription

	if r == nil {
		return
	}

	if debug {
		fmt.Printf("Incoming offer from %s\n", r.RemoteAddr)
	}

	if r.Method != http.MethodPost {
		return
	}

	if r.Body == nil {
		if _, err := w.Write(nil); err != nil {
			fmt.Printf(" error writing response (%s)\n", err)
			return
		}
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
		fmt.Printf(" error decoding offer (%s)\n", err)
		return
	}

	webrtc_config := webrtc.Configuration{}

	if config.StunUrl != "" {
		if debug {
			fmt.Printf("Starting with STUN URL: %s\n", config.StunUrl)
		}
		webrtc_config = webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{
					URLs: []string{config.StunUrl},
				},
			},
		}
	}

	var settings = webrtc.SettingEngine{}

	if config.MinPort > 0 && config.MaxPort > 0 {
		if err := settings.SetEphemeralUDPPortRange(config.MinPort, config.MaxPort); err != nil {
			fmt.Printf("Error: cannot set UDP port range; min:%d, max:%d\n", config.MinPort, config.MaxPort)
		}
	}

	if config.ExternalIP != "" {
		settings.SetNAT1To1IPs([]string{config.ExternalIP}, webrtc.ICECandidateTypeHost)
	}
	pc, err := webrtc.NewAPI(webrtc.WithSettingEngine(settings)).NewPeerConnection(webrtc_config)

	if debug {
		fmt.Println("New peer connection created...")
	}
	if err != nil {
		fmt.Printf(" error creating a PeerConnection (%s)\n", err)
		return
	}

	peerConnections.mutex.Lock()
	peerConnections.connections = append(peerConnections.connections, pc)
	peerConnections.mutex.Unlock()

	pc.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		if debug {
			fmt.Printf("ICE connection state has changed to %s.\n", connectionState.String())
		}

		if connectionState == webrtc.ICEConnectionStateFailed {
			peerConnections.mutex.Lock()
			defer peerConnections.mutex.Unlock()

			for i := range peerConnections.connections {
				if peerConnections.connections[i] == pc {
					if err = pc.Close(); err != nil {
						fmt.Printf(" error closing PeerConnection (%s)\n", err)
						return
					}

					peerConnections.connections = append(peerConnections.connections[:i], peerConnections.connections[i+1:]...)
					break
				}
			}
		}
	})

	pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			if debug {
				fmt.Println("Gathering is complete...")
			}
			return
		}
		if debug {
			fmt.Printf("New ICE candidate: %s.\n", candidate.String())
		}
	})

	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		if debug {
			fmt.Printf("PeerConnection state has changed to %s.\n", state.String())
		}

		if state.String() == "closed" {
			if debug {
				listPeerConnections()
			}
		}
	})

	pc.OnSignalingStateChange(func(state webrtc.SignalingState) {
		if debug {
			fmt.Printf("Signaling state has changed to %s.\n", state.String())
		}
	})

	pc.OnDataChannel(func(d *webrtc.DataChannel) {
		if debug {
			fmt.Printf("New DataChannel %s %d\n", d.Label(), d.ID())
		}

		d.OnOpen(func() {
			if debug {
				fmt.Printf("Data channel '%s'-'%d' open.\n", d.Label(), d.ID())
			}
		})

		// Register text message handling
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			if debug {
				fmt.Printf("Message from DataChannel '%s': '%s'\n", d.Label(), string(msg.Data))
			}

			// Echo it back:
			if err := d.Send(msg.Data); err != nil {
				fmt.Printf(" error sending message (%s)\n", err)
			}
		})
	})

	if err := pc.SetRemoteDescription(offer); err != nil {
		fmt.Printf(" error setting remote description (%s)\n", err)
		return
	}

	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		fmt.Printf(" error creating answer (%s)\n", err)
		return
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatheringComplete := webrtc.GatheringCompletePromise(pc)

	err = pc.SetLocalDescription(answer)
	if err != nil {
		fmt.Printf(" error setting the local description (%s)\n", err)
		return
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	<-gatheringComplete

	response, err := json.Marshal(pc.LocalDescription())
	if err != nil {
		fmt.Printf(" error marshalling the response (%s)\n", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(response); err != nil {
		fmt.Printf(" error writing response (%s)\n", err)
		return
	}

	if debug {
		listPeerConnections()
	}
}

func main() {
	var err error
	config, err = LoadConfig()
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	host := config.Addr
	port := config.Port
	certFile := config.CertFile
	keyFile := config.KeyFile
	allowedOrigins := config.AllowedOrigins
	debug = config.Debug

	peerConnections = PeerConnections{}

	mux := http.NewServeMux()
	offerHandler := http.HandlerFunc(offer)

	if config.MaxReqRate > 0 {
		limiter := tollbooth.NewLimiter(config.MaxReqRate, nil)
		mux.Handle("/offer", tollbooth.LimitFuncHandler(limiter, offerHandler))
	} else {
		mux.Handle("/offer", offerHandler)
	}

	c := cors.New(cors.Options{})
	if allowedOrigins != "" {
		ao := strings.Split(allowedOrigins, ",")
		fmt.Println("Allowed origins:")
		fmt.Println(ao)
		c = cors.New(cors.Options{
			AllowedOrigins: ao,
			AllowedMethods: []string{http.MethodPost},
			Debug:          debug,
		})
	}

	handler := c.Handler(mux)

	addr := fmt.Sprintf("%s:%d", host, port)

	fmt.Printf("STUN URL: %s\n", config.StunUrl)
	fmt.Printf("Cert file: %s, key file: %s\n", certFile, keyFile)
	if debug {
		fmt.Println("Debug mode on")
	} else {
		fmt.Println("Debug mode off")
	}
	if config.MinPort > 0 && config.MaxPort > 0 {
		fmt.Printf("RTP range, min port: %d, max port: %d\n", config.MinPort, config.MaxPort)
	}
	if config.MaxReqRate > 0 {
		fmt.Printf("Max rate allowed: %f/sec\n", config.MaxReqRate)
	}
	fmt.Printf("Listening on %s...\n", addr)
	fmt.Println("=====================================")

	if keyFile != "" && certFile != "" {
		log.Fatal(http.ListenAndServeTLS(addr, certFile, keyFile, handler))
	} else {
		log.Fatal(http.ListenAndServe(addr, handler))
	}
}
