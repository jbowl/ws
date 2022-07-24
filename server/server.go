package server

import (
	"context"
	"encoding/json"
	"github/jbowl/ws/types"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/jbowl/hodlapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// templ represents a single template
type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

// ServeHTTP handles the HTTP request.
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("ServeHTTP")
	t.once.Do(func() {
		log.Printf("t.templ")
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	t.templ.Execute(w, r)
}

// We'll need to define an Upgrader
// this will require a Read and Write buffer size
var upgrader = websocket.Upgrader{
	ReadBufferSize:  10240,
	WriteBufferSize: 10240,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Server struct {
	bc      *hodlapi.BreweryServiceClient
	Healthy *int64
}

func newBrewery(brewery *hodlapi.Brewery) types.BreweryResult {
	return types.BreweryResult{
		ID:          brewery.Id,
		Name:        brewery.Name,
		BreweryType: brewery.BreweryType,

		Street: brewery.Street,
		//	address_2: null,
		//	address_3: null,
		City:  brewery.City,
		State: brewery.State,

		CountryProvince: brewery.CountryProvince,
		PostalCode:      brewery.PostalCode,
		Country:         brewery.Country,
		Longitude:       brewery.Longitude,
		Latitude:        brewery.Latitude,
		Phone:           brewery.Phone,
		Website:         brewery.WebsiteUrl,
		Updated:         brewery.UpdatedAt,
		Created:         brewery.CreatedAt,
		//	updated_at: "2018-08-23T23:24:11.758Z",
		//	created_at: "2018-08-23T23:24:11.758Z"
	}
}

// define a reader which will listen for
// new messages being sent to our WebSocket
// endpoint
func reader(client hodlapi.BreweryServiceClient, conn *websocket.Conn) {
	for {
		// read in a message
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			//return
		}

		// expecting in the format "by_state=&...."
		filter := string(p)

		//filter, err := parse(uri)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		stream, err := client.ListBreweries(ctx, &hodlapi.Filter{Query: filter})

		if err != nil {
			log.Println(err)
			return
		}

		for {
			brewery, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}

			// copy proto to JSON
			res := newBrewery(brewery)

			buff, err := json.Marshal(res)

			// write response from gRPC Stream to client
			if err := conn.WriteMessage(websocket.TextMessage, buff); err != nil {
				log.Println(err)
				return
			}
		}
		// end
		// print out that message for clarity
		log.Println(string(p))

		if err := conn.WriteMessage(messageType, p); err != nil {
			if err := conn.WriteMessage(messageType, p); err != nil {
				log.Println(err)
				return
			}
		}

	}
}

func Breweries(client hodlapi.BreweryServiceClient) http.Handler {
	//func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// upgrade this connection to a WebSocket
			// connection
			ws, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Println(err)
			}

			log.Println("Client Connected")
			err = ws.WriteMessage(1, []byte("connected"))
			if err != nil {
				log.Println(err)
			}
			// listen indefinitely for new messages coming
			// through on our WebSocket connection
			reader(client, ws)
		})
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello"))
}

func (s *Server) Start(port string, apiAddr string) <-chan os.Signal {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	creds, err := credentials.NewClientTLSFromFile("./jbowl.cert", "")
	if err != nil {
		log.Fatalf("could not process the credentials: %v", err)
	}

	if creds != nil {
	} // dummy check unused var

	//	b, _ := ioutil.ReadFile("local_ca.cert")
	//	cp := x509.NewCertPool()
	//	if !cp.AppendCertsFromPEM(b) {

	//		return shutdown
	//return nil, errors.New("credentials: failed to append certificates")
	go func() {

		log.Printf("dialing")

		// easier local dev option by not encrypting
		if os.Getenv("INSECURE") == "TRUE" {
			creds = insecure.NewCredentials()
		}
		// secure use if TLS enabled on server
		conn, err := grpc.Dial(apiAddr, grpc.WithTransportCredentials(creds))
		//
		// insecure use if TLS not enabled on server
		//  will get this, Detail": "rpc error: code = Unavailable desc = connection closed before server preface received",
		// if not using secure

		if err != nil {
			log.Printf("fail to dial: %v", err)
			//	log.Fatalf("fail to dial: %v", err)
		}
		log.Printf("dialed")
		defer conn.Close()

		router := mux.NewRouter().StrictSlash(true)

		breweryClient := hodlapi.NewBreweryServiceClient(conn)

		// send html to client browser on root GET
		router.Handle("/", &templateHandler{filename: "breweries.html"})
		//
		router.Handle("/ws", Breweries(breweryClient))
		// ALB health check
		router.HandleFunc("/healthz", healthCheck)

		httpServer := http.Server{
			Addr:         ":" + port,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			IdleTimeout:  5 * time.Second,
			Handler:      router,
		}

		log.Fatal(httpServer.ListenAndServe())
	}()

	return shutdown
}
