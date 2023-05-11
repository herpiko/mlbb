package main

import (
	"bufio"
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"example_nrpc/proto/hello"

	"github.com/darmawan01/toldata"
	"github.com/nats-io/nats.go"
)

func main() {

	natsURL := nats.DefaultURL
	if len(os.Args) == 2 {
		natsURL = os.Args[1]
	}
	// Connect to the NATS server.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bus, err := toldata.NewBus(ctx, toldata.ServiceConfiguration{URL: natsURL})
	defer bus.Close()

	svc := hello.NewHelloServicesToldataClient(bus)

	// nc, err := nats.Connect(natsURL, nats.Timeout(5*time.Second))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer nc.Close()
	//
	// cli := hello.NewHelloServicesClient(nc)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		res, err := svc.Greeting(r.Context(), &hello.GreetingRequest{
			Firstname: "Rahmat",
			Lastname:  "Fathoni",
		})
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("coba"))
			return
		}
		w.Write([]byte(res.Fullname))
	})

	http.HandleFunc("/upload2", func(w http.ResponseWriter, r *http.Request) {
		b, err := os.ReadFile("golang.jpg")
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("coba"))
			return
		}

		resp, err := svc.Upload2(r.Context(), &hello.UploadRequest{
			Data: b,
		})
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("coba"))
			return
		}

		w.Write([]byte(resp.GetName()))
	})
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		log.Println("upload")

		file, _, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		stream, err := svc.Upload(r.Context()) // 10ms
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("coba"))
			return
		}

		reader := bufio.NewReader(file)
		//Set the buffer size dynamically
		// For instance, 30MB file with 1024 buffer size will take 30 seconds to upload
		// But 30MB file with 10024 buffer size will take 3 seconds to upload
		buffer := make([]byte, 1024)

		for { // 137ms
			n, err := reader.Read(buffer)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal("cannot read chunk to buffer: ", err)
			}

			stream.Send(&hello.UploadRequest{
				Data: buffer[:n],
			})

		}

		_, _ = stream.Done()

		return
		//w.Write([]byte(resp.GetName()))
	})
	http.HandleFunc("/upload-direct", func(w http.ResponseWriter, r *http.Request) {
		log.Println("upload-direct")

		file, _, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		reader := bufio.NewReader(file)
		buffer := make([]byte, 1024)

		for {
			_, err := reader.Read(buffer)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal("cannot read chunk to buffer: ", err)
			}
		}

		return
	})

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		err = http.ListenAndServe(":8080", nil)
		if err != nil {
			panic(err)
		}
	}()

	<-sig
	cancel()
}
