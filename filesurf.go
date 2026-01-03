package main

import (
    "flag"
    "fmt"
    "log"
    "net/http"
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "crypto/x509"
    "crypto/x509/pkix"
    "encoding/pem"
    "math/big"
    "path/filepath"
    "time"
    "os"
)


const DEFAULTCERTNAME string = "filesurf.pem"
const DEFAULTKEYNAME string = "filesurf.key"
const BUFFER_SIZE int64 = 20 << 20 // 20 MB buffer size for file upload


func main() {
    // Command-line flags
    port := flag.String("p", "8090", "Port to listen on.")
    baseDir := flag.String("d", ".", "Base directory to serve.")
    enableHTTPS := flag.Bool("s", false, "Enable HTTPS.")
    certFile := flag.String("cert", "", "Path to TLS certificate file (required if -s)")
    keyFile := flag.String("key", "", "Path to TLS key file (required if -s)")

    flag.Parse()

    if *port == "" {
        fmt.Fprintln(os.Stderr, "Error: -p is required, choose a valid port number.")
	flag.Usage()
        os.Exit(1)
    }
    //TODO: check if port is available

    // Validate base directory
    info, err := os.Stat(*baseDir)
    if err != nil {
        log.Fatalf("Failed to access directory %s: %v", *baseDir, err)
    }
    if !info.IsDir() {
        log.Fatalf("Provided path is not a directory: %s", *baseDir)
    }

    // Validate HTTPS options
    if *enableHTTPS {
        log.Printf("Secure mode enabled, using HTTPS.")

        if *certFile == "" && *keyFile == "" {
            log.Printf("No key or cert defined, using default self-generated.")
            *certFile = DEFAULTCERTNAME
            *keyFile = DEFAULTKEYNAME
            _, erri := os.Stat(*certFile)
            _, errk := os.Stat(*keyFile)

            if erri != nil || errk != nil {
                   log.Printf("Defaults not found. Generating new cert.")
                    generateKeyAndCert()
            }
        } else if *certFile == "" || *keyFile == "" {
            log.Fatal("HTTPS enabled but -cert or -key not provided")
        }
    }

    // File server with directory listing and upload endpoint
    fileServer := http.FileServer(http.Dir(*baseDir))
    http.Handle("/", fileServer)
    http.HandleFunc("/upload", fileUploadHandler )

    address := ":" + *port

    log.Printf("Serving directory: %s", *baseDir)
    log.Printf("Use endpoint /upload for uploading. POST request with form field 'content' = your data")

    if *enableHTTPS {
        log.Printf("HTTPS enabled on https://0.0.0.0%s/", address)
        err = http.ListenAndServeTLS(address, *certFile, *keyFile, nil)
    } else {
        log.Printf("HTTP enabled on http://0.0.0.0%s/", address)
        err = http.ListenAndServe(address, nil)
    }

    if err != nil {
        log.Fatalf("Server failed: %v", err)
    }

}


func generateKeyAndCert() {
    priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil {
        panic(err)
    }

    notBefore := time.Now()
    notAfter := notBefore.Add(365 * 24 * time.Hour)

    serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
    if err != nil {
        panic(err)
    }

    template := x509.Certificate{
        SerialNumber: serialNumber,
        Subject: pkix.Name{
            Organization: []string{"Filesurf"},
        },
        NotBefore:             notBefore,
        NotAfter:              notAfter,
        KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
        ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
        BasicConstraintsValid: true,
    }

    derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
    if err != nil {
        panic(err)
    }

    certOut, err := os.Create(DEFAULTCERTNAME)
    if err != nil {
        panic(err)
    }
    pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
    certOut.Close()

    keyOut, err := os.Create(DEFAULTKEYNAME)
    if err != nil {
        panic(err)
    }
    pem.Encode(keyOut, pemBlockForKey(priv))
    keyOut.Close()
}


func pemBlockForKey(priv *ecdsa.PrivateKey) *pem.Block {
    b, err := x509.MarshalECPrivateKey(priv)
    if err != nil {
        panic(err)
    }
    return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
}


func fileUploadHandler (w http.ResponseWriter, r *http.Request) {
    r.ParseMultipartForm( BUFFER_SIZE )

    // Retrieve the file from form data
    file, handler, err := r.FormFile("content")
    if err != nil {
        http.Error(w, "Error retrieving the contentfile", http.StatusBadRequest)
        return
    }
    defer file.Close()

    var sstr = prettyPrintSize(handler.Size)

    fmt.Fprintf(w, "Uploaded File: %s\n", handler.Filename)
    fmt.Fprintf(w, "File Size: %s\n", sstr)
    fmt.Fprintf(w, "MIME Header: %v\n", handler.Header)

    // Now let’s save it locally
    dst, err := createFile(handler.Filename)
    if err != nil {
        http.Error(w, "Error saving the file", http.StatusInternalServerError)
        return
    }
    defer dst.Close()

    // Copy the uploaded file to the destination file
    if _, err := dst.ReadFrom(file); err != nil {
        http.Error(w, "Error saving the file", http.StatusInternalServerError)
    }    

    log.Printf("Received file: %s, Size: %s", handler.Filename, sstr)
}


func createFile(filename string) (*os.File, error) {
    // Create an uploads directory if it doesn’t exist
    if _, err := os.Stat("uploads"); os.IsNotExist(err) {
        os.Mkdir("uploads", 0755)
    }

    // Build the file path and create it
    dst, err := os.Create(filepath.Join("uploads", filename))
    if err != nil {
        return nil, err
    }

    return dst, nil
}


func prettyPrintSize(s int64) (ps string) {
    const unit = 1024
    if s < unit {
        return fmt.Sprintf("%dB", s)
    }

    div, exp := float64(unit), 0
    for n := float64(s) / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }

    return fmt.Sprintf("%.1f%cB",
        float64(s)/div,
        "KMGTPE"[exp],
    )
}
